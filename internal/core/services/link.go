package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/jackc/pgx/v4"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/event"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/jsonschema"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	linkState "github.com/polygonid/sh-id-platform/pkg/link"
	"github.com/polygonid/sh-id-platform/pkg/network"
	"github.com/polygonid/sh-id-platform/pkg/notifications"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
)

var (
	// ErrLinkAlreadyActive link is already active
	ErrLinkAlreadyActive = errors.New("link is already active")
	// ErrLinkAlreadyInactive link is already inactive
	ErrLinkAlreadyInactive = errors.New("link is already inactive")
	// ErrLinkAlreadyExpired - link already expired
	ErrLinkAlreadyExpired = errors.New("cannot issue a credential for an expired link")
	// ErrLinkMaxExceeded - link max exceeded
	ErrLinkMaxExceeded = errors.New("cannot issue a credential for an expired link")
	// ErrLinkInactive - link inactive
	ErrLinkInactive = errors.New("cannot issue a credential for an inactive link")
	// ErrClaimAlreadyIssued - claim already issued
	ErrClaimAlreadyIssued = errors.New("the claim was already issued for the user")
	// ErrUnsupportedCredentialStatusType - unsupported credential status type
	ErrUnsupportedCredentialStatusType = errors.New("unsupported credential status type")
)

// Link - represents a link in the issuer node
type Link struct {
	storage          *db.Storage
	claimsService    ports.ClaimsService
	qrService        ports.QrStoreService
	claimRepository  ports.ClaimsRepository
	linkRepository   ports.LinkRepository
	schemaRepository ports.SchemaRepository
	loader           loader.DocumentLoader
	sessionManager   ports.SessionRepository
	publisher        pubsub.Publisher
	identityService  ports.IdentityService
	networkResolver  network.Resolver
}

// NewLinkService - constructor
func NewLinkService(storage *db.Storage, claimsService ports.ClaimsService, qrService ports.QrStoreService, claimRepository ports.ClaimsRepository, linkRepository ports.LinkRepository, schemaRepository ports.SchemaRepository, ld loader.DocumentLoader, sessionManager ports.SessionRepository, publisher pubsub.Publisher, identityService ports.IdentityService, networkResolver network.Resolver) ports.LinkService {
	return &Link{
		storage:          storage,
		claimsService:    claimsService,
		qrService:        qrService,
		claimRepository:  claimRepository,
		linkRepository:   linkRepository,
		schemaRepository: schemaRepository,
		loader:           ld,
		sessionManager:   sessionManager,
		publisher:        publisher,
		identityService:  identityService,
		networkResolver:  networkResolver,
	}
}

// Save - save a new credential
func (ls *Link) Save(
	ctx context.Context,
	did w3c.DID,
	maxIssuance *int,
	validUntil *time.Time,
	schemaID uuid.UUID,
	credentialExpiration *time.Time,
	credentialSignatureProof bool,
	credentialMTPProof bool,
	credentialSubject domain.CredentialSubject,
	refreshService *verifiable.RefreshService,
	displayMethod *verifiable.DisplayMethod,
	credentialStatusType verifiable.CredentialStatusType,
) (*domain.Link, error) {
	schemaDB, err := ls.schemaRepository.GetByID(ctx, did, schemaID)
	if err != nil {
		return nil, err
	}

	if err := ls.validateCredentialSubjectAgainstSchema(ctx, credentialSubject, schemaDB); err != nil {
		log.Error(ctx, "validating credential subject", "err", err)
		return nil, ErrParseClaim
	}
	if err = ls.validateRefreshService(refreshService, credentialExpiration); err != nil {
		log.Error(ctx, "validating refresh service", "err", err)
		return nil, err
	}
	if err = ls.validateDisplayMethod(displayMethod); err != nil {
		log.Error(ctx, "validating display method", "err", err)
		return nil, err
	}

	link := domain.NewLink(did, maxIssuance, validUntil, schemaID, credentialExpiration, credentialSignatureProof, credentialMTPProof, credentialSubject, refreshService, displayMethod, &credentialStatusType)
	_, err = ls.linkRepository.Save(ctx, ls.storage.Pgx, link)
	if err != nil {
		return nil, err
	}

	link.Schema = schemaDB

	return link, nil
}

// Activate - activates or deactivates a credential link
func (ls *Link) Activate(ctx context.Context, issuerID w3c.DID, linkID uuid.UUID, active bool) error {
	link, err := ls.linkRepository.GetByID(ctx, issuerID, linkID)
	if err != nil {
		return err
	}

	if link.Active && active {
		return ErrLinkAlreadyActive
	}

	if !link.Active && !active {
		return ErrLinkAlreadyInactive
	}

	link.Active = active
	_, err = ls.linkRepository.Save(ctx, ls.storage.Pgx, link)
	return err
}

// GetByID returns a link by id and issuerDID
func (ls *Link) GetByID(ctx context.Context, issuerID w3c.DID, id uuid.UUID) (*domain.Link, error) {
	link, err := ls.linkRepository.GetByID(ctx, issuerID, id)
	if err != nil {
		if errors.Is(err, repositories.ErrLinkDoesNotExist) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	return link, nil
}

// GetAll returns all links from issueDID of type lType filtered by query string
func (ls *Link) GetAll(ctx context.Context, issuerDID w3c.DID, status ports.LinkStatus, query *string) ([]domain.Link, error) {
	return ls.linkRepository.GetAll(ctx, issuerDID, status, query)
}

// Delete - delete a link by id
func (ls *Link) Delete(ctx context.Context, id uuid.UUID, did w3c.DID) error {
	return ls.linkRepository.Delete(ctx, id, did)
}

// CreateQRCode - generates a qr code for a link
func (ls *Link) CreateQRCode(ctx context.Context, issuerDID w3c.DID, linkID uuid.UUID, serverURL string) (*ports.CreateQRCodeResponse, error) {
	link, err := ls.GetByID(ctx, issuerDID, linkID)
	if err != nil {
		return nil, err
	}

	err = ls.validate(ctx, link)
	if err != nil {
		return nil, err
	}

	sessionID := uuid.New().String()
	reqID := uuid.New().String()
	qrCode := &protocol.AuthorizationRequestMessage{
		From:     issuerDID.String(),
		ID:       reqID,
		ThreadID: reqID,
		Typ:      packers.MediaTypePlainMessage,
		Type:     protocol.AuthorizationRequestMessageType,
		Body: protocol.AuthorizationRequestMessageBody{
			CallbackURL: fmt.Sprintf("%s/v1/credentials/links/callback?sessionID=%s&linkID=%s", serverURL, sessionID, linkID.String()),
			Reason:      authReason,
			Scope:       make([]protocol.ZeroKnowledgeProofRequest, 0),
		},
	}

	err = ls.sessionManager.Set(ctx, sessionID, *qrCode)
	if err != nil {
		return nil, err
	}

	err = ls.sessionManager.SetLink(ctx, linkState.CredentialStateCacheKey(linkID.String(), sessionID), *linkState.NewStatePending())
	if err != nil {
		return nil, err
	}

	raw, err := json.Marshal(qrCode)
	if err != nil {
		return nil, err
	}

	id, err := ls.qrService.Store(ctx, raw, DefaultQRBodyTTL)
	if err != nil {
		return nil, err
	}

	return &ports.CreateQRCodeResponse{
		SessionID: sessionID,
		QrCode:    ls.qrService.ToURL(serverURL, id),
		QrID:      id,
		Link:      link,
	}, nil
}

// IssueOrFetchClaim - Create a new claim
func (ls *Link) IssueOrFetchClaim(ctx context.Context, sessionID string, issuerDID w3c.DID, userDID w3c.DID, linkID uuid.UUID, hostURL string) (*protocol.CredentialsOfferMessage, error) {
	link, err := ls.linkRepository.GetByID(ctx, issuerDID, linkID)
	if err != nil {
		log.Error(ctx, "cannot fetch the link", "err", err)
		return nil, err
	}

	if link.CredentialStatusType == nil {
		link.CredentialStatusType = common.ToPointer(verifiable.Iden3commRevocationStatusV1)
	}

	resolverPrefix, err := common.ResolverPrefix(&issuerDID)
	if err != nil {
		log.Error(ctx, "error getting resolver prefix", "err", err)
		return nil, err
	}

	rhsSettings, err := ls.networkResolver.GetRhsSettings(ctx, resolverPrefix)
	if err != nil {
		log.Error(ctx, "error getting rhs settings", "err", err)
		return nil, err
	}

	if !ls.networkResolver.IsCredentialStatusTypeSupported(rhsSettings, *link.CredentialStatusType) {
		log.Error(ctx, "unsupported credential status type", "type", link.CredentialStatusType)
		return nil, ErrUnsupportedCredentialStatusType
	}

	issuedByUser, err := ls.claimRepository.GetClaimsIssuedForUser(ctx, ls.storage.Pgx, issuerDID, userDID, linkID)
	if err != nil {
		log.Error(ctx, "cannot fetch the claims issued for the user", "err", err, "issuerDID", issuerDID, "userDID", userDID)
		return nil, err
	}

	if err := ls.validate(ctx, link); err != nil {
		setLinkError := ls.sessionManager.SetLink(ctx, linkState.CredentialStateCacheKey(linkID.String(), sessionID), *linkState.NewStateError(err))
		if setLinkError != nil {
			log.Error(ctx, "cannot set the state", "err", setLinkError)
			return nil, setLinkError
		}

		return nil, err
	}

	var credentialIssuedID uuid.UUID
	var credentialIssued *domain.Claim

	schema, err := ls.schemaRepository.GetByID(ctx, issuerDID, link.SchemaID)
	if err != nil {
		log.Error(ctx, "cannot fetch the schema", "err", err)
		return nil, err
	}

	claimRequestProofs := ports.ClaimRequestProofs{
		BJJSignatureProof2021:      link.CredentialSignatureProof,
		Iden3SparseMerkleTreeProof: link.CredentialMTPProof,
	}
	if len(issuedByUser) == 0 {
		link.CredentialSubject["id"] = userDID.String()

		claimReq := ports.NewCreateClaimRequest(&issuerDID,
			nil,
			schema.URL,
			link.CredentialSubject,
			link.CredentialExpiration,
			schema.Type,
			nil, nil, nil,
			claimRequestProofs,
			&linkID,
			true,
			*link.CredentialStatusType,
			link.RefreshService,
			nil,
			link.DisplayMethod,
		)

		credentialIssued, err = ls.claimsService.CreateCredential(ctx, claimReq)
		if err != nil {
			log.Error(ctx, "cannot create the claim", "err", err.Error())
			return nil, err
		}

		err = ls.storage.Pgx.BeginFunc(ctx,
			func(tx pgx.Tx) error {
				link.IssuedClaims += 1
				_, err := ls.linkRepository.Save(ctx, ls.storage.Pgx, link)
				if err != nil {
					return err
				}

				credentialIssuedID, err = ls.claimRepository.Save(ctx, ls.storage.Pgx, credentialIssued)
				if err != nil {
					return err
				}

				return nil
			})
		if err != nil {
			return nil, err
		}
	} else {
		credentialIssuedID = issuedByUser[0].ID
		credentialIssued = issuedByUser[0]
	}

	credentialIssued.ID = credentialIssuedID
	if link.CredentialSignatureProof {
		err = ls.publisher.Publish(ctx, event.CreateCredentialEvent, &event.CreateCredential{CredentialIDs: []string{credentialIssued.ID.String()}, IssuerID: issuerDID.String()})
		if err != nil {
			log.Error(ctx, "publish CreateCredentialEvent", "err", err.Error(), "credential", credentialIssued.ID.String())
		}
	}

	if link.CredentialSignatureProof {
		credOffer, err := notifications.NewOfferMsg(fmt.Sprintf("%s/v1/agent", hostURL), credentialIssued)
		return credOffer, err
	} else {
		if credentialIssued.MTPProof.Bytes != nil {
			credOffer, err := notifications.NewOfferMsg(fmt.Sprintf("%s/v1/agent", hostURL), credentialIssued)
			return credOffer, err
		}
		log.Info(ctx, "credential issued without MTP proof. Publishing state have to be done", "credential", credentialIssued.ID.String())
		return nil, nil
	}
}

// ProcessCallBack - process the callback.
func (ls *Link) ProcessCallBack(ctx context.Context, message string, sessionID uuid.UUID, linkID uuid.UUID, hostURL string) (*protocol.CredentialsOfferMessage, error) {
	arm, err := ls.identityService.Authenticate(ctx, message, sessionID, hostURL)
	if err != nil {
		log.Error(ctx, "error authenticating", "err", err.Error())
		return nil, err
	}

	userDID, err := w3c.ParseDID(arm.From)
	if err != nil {
		log.Error(ctx, "parsing user did", "err", err)
		return nil, err
	}

	issuerDID, err := w3c.ParseDID(arm.To)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err)
		return nil, err
	}

	offer, err := ls.IssueOrFetchClaim(ctx, sessionID.String(), *issuerDID, *userDID, linkID, hostURL)
	if err != nil {
		log.Error(ctx, "error issuing claim", "err", err)
		return nil, err
	}
	return offer, nil
}

// GetQRCode - return the link qr code.
func (ls *Link) GetQRCode(ctx context.Context, sessionID uuid.UUID, issuerID w3c.DID, linkID uuid.UUID) (*ports.GetQRCodeResponse, error) {
	link, err := ls.GetByID(ctx, issuerID, linkID)
	if err != nil {
		log.Error(ctx, "error fetching the link from the database", "err", err)
		return nil, err
	}

	linkStateInCache, err := ls.sessionManager.GetLink(ctx, linkState.CredentialStateCacheKey(linkID.String(), sessionID.String()))
	if err != nil {
		log.Error(ctx, "error fetching the link state from the cache", "err", err)
		return nil, err
	}
	return &ports.GetQRCodeResponse{
		State: &linkStateInCache,
		Link:  link,
	}, nil
}

func (ls *Link) validate(ctx context.Context, link *domain.Link) error {
	if link.ValidUntil != nil && time.Now().UTC().After(*link.ValidUntil) {
		log.Debug(ctx, "cannot issue a credential for an expired link")
		return ErrLinkAlreadyExpired
	}

	if link.MaxIssuance != nil && *link.MaxIssuance <= link.IssuedClaims {
		log.Debug(ctx, "cannot dispatch more credentials for this link")
		return ErrLinkMaxExceeded
	}

	if !link.Active {
		log.Debug(ctx, "cannot dispatch credentials for inactive link")
		return ErrLinkInactive
	}

	return nil
}

func (ls *Link) validateCredentialSubjectAgainstSchema(ctx context.Context, cSubject domain.CredentialSubject, schemaDB *domain.Schema) error {
	return jsonschema.ValidateCredentialSubject(ctx, ls.loader, schemaDB.URL, schemaDB.Type, cSubject)
}

func (ls *Link) validateRefreshService(rs *verifiable.RefreshService, expiration *time.Time) error {
	if rs == nil {
		return nil
	}
	if expiration == nil {
		return ErrRefreshServiceLacksExpirationTime
	}

	if rs.ID == "" {
		return ErrRefreshServiceLacksURL
	}
	_, err := url.ParseRequestURI(rs.ID)
	if err != nil {
		return ErrRefreshServiceLacksURL
	}

	switch rs.Type {
	case verifiable.Iden3RefreshService2023:
		return nil
	default:
		return ErrUnsupportedRefreshServiceType
	}
}

func (ls *Link) validateDisplayMethod(dm *verifiable.DisplayMethod) error {
	if dm == nil {
		return nil
	}

	if dm.ID == "" {
		return ErrDisplayMethodLacksURL
	}
	if _, err := url.ParseRequestURI(dm.ID); err != nil {
		return ErrDisplayMethodLacksURL
	}

	switch dm.Type {
	case verifiable.Iden3BasicDisplayMethodV1:
		return nil
	default:
		return ErrUnsupportedDisplayMethodType
	}
}
