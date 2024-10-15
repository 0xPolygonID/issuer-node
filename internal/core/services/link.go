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
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/json_schema"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/network"
	"github.com/polygonid/sh-id-platform/internal/notifications"
	"github.com/polygonid/sh-id-platform/internal/pubsub"
	"github.com/polygonid/sh-id-platform/internal/qrlink"
	"github.com/polygonid/sh-id-platform/internal/repositories"
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
)

// Link - represents a link in the issuer node
type Link struct {
	cfg              config.UniversalLinks
	storage          *db.Storage
	claimsService    ports.ClaimService
	qrService        ports.QrStoreService
	claimRepository  ports.ClaimRepository
	linkRepository   ports.LinkRepository
	schemaRepository ports.SchemaRepository
	loader           loader.DocumentLoader
	sessionManager   ports.SessionRepository
	publisher        pubsub.Publisher
	identityService  ports.IdentityService
	networkResolver  network.Resolver
}

// NewLinkService - constructor
func NewLinkService(storage *db.Storage, claimsService ports.ClaimService, qrService ports.QrStoreService, claimRepository ports.ClaimRepository, linkRepository ports.LinkRepository, schemaRepository ports.SchemaRepository, ld loader.DocumentLoader, sessionManager ports.SessionRepository, publisher pubsub.Publisher, identityService ports.IdentityService, networkResolver network.Resolver, cfg config.UniversalLinks) ports.LinkService {
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
		cfg:              cfg,
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
) (*domain.Link, error) {
	schemaDB, err := ls.schemaRepository.GetByID(ctx, did, schemaID)
	if err != nil {
		return nil, err
	}

	if err := ls.validateCredentialSubjectAgainstSchema(ctx, credentialSubject, schemaDB); err != nil {
		log.Error(ctx, "validating credential subject", "err", err, "subject", credentialSubject, "schema-id", schemaDB.ID, "schema-type", schemaDB.Type)
		return nil, ErrInvalidCredentialSubject
	}
	if err = ls.validateRefreshService(refreshService, credentialExpiration); err != nil {
		log.Error(ctx, "validating refresh service", "err", err)
		return nil, err
	}
	if err = ls.validateDisplayMethod(displayMethod); err != nil {
		log.Error(ctx, "validating display method", "err", err)
		return nil, err
	}

	link := domain.NewLink(did, maxIssuance, validUntil, schemaID, credentialExpiration, credentialSignatureProof, credentialMTPProof, credentialSubject, refreshService, displayMethod)
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
func (ls *Link) GetByID(ctx context.Context, issuerDID w3c.DID, id uuid.UUID, serverURL string) (*domain.Link, error) {
	link, err := ls.linkRepository.GetByID(ctx, issuerDID, id)
	if err != nil {
		if errors.Is(err, repositories.ErrLinkDoesNotExist) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	if link.AuthorizationRequestMessage == nil {
		reqID := uuid.New().String()
		authorizationRequestMessage := &protocol.AuthorizationRequestMessage{
			From:     issuerDID.String(),
			ID:       reqID,
			ThreadID: reqID,
			Typ:      packers.MediaTypePlainMessage,
			Type:     protocol.AuthorizationRequestMessageType,
			Body: protocol.AuthorizationRequestMessageBody{
				CallbackURL: fmt.Sprintf(ports.LinksCallbackURL, serverURL, issuerDID.String(), link.ID.String()),
				Reason:      authReason,
				Scope:       make([]protocol.ZeroKnowledgeProofRequest, 0),
			},
		}
		if err := ls.linkRepository.AddAuthorizationRequest(ctx, link.ID, issuerDID, authorizationRequestMessage); err != nil {
			log.Error(ctx, "cannot add the authorization request", "err", err)
			return nil, err
		}

		link.AuthorizationRequestMessage = &pgtype.JSONB{}
		err = link.AuthorizationRequestMessage.Set(authorizationRequestMessage)
		if err != nil {
			log.Error(ctx, "cannot assign the authorization", "err", err)
			return nil, err
		}
	}

	ls.addLinksToLink(link, serverURL, issuerDID)
	return link, nil
}

// GetAll returns all links from issueDID of type lType filtered by query string
func (ls *Link) GetAll(ctx context.Context, issuerDID w3c.DID, status ports.LinkStatus, query *string, serverURL string) ([]*domain.Link, error) {
	links, err := ls.linkRepository.GetAll(ctx, issuerDID, status, query)
	if err != nil {
		return links, err
	}

	for _, link := range links {
		ls.addLinksToLink(link, serverURL, issuerDID)
	}
	return links, nil
}

func (ls *Link) addLinksToLink(link *domain.Link, serverURL string, issuerDID w3c.DID) {
	if link.AuthorizationRequestMessage != nil {
		link.DeepLink = qrlink.NewDeepLink(serverURL, link.ID, &issuerDID)
		link.UniversalLink = qrlink.NewUniversal(ls.cfg.BaseUrl, serverURL, link.ID, &issuerDID)
	}
}

// Delete - delete a link by id
func (ls *Link) Delete(ctx context.Context, id uuid.UUID, did w3c.DID) error {
	return ls.linkRepository.Delete(ctx, id, did)
}

// CreateQRCode - generates a qr code for a link
func (ls *Link) CreateQRCode(ctx context.Context, issuerDID w3c.DID, linkID uuid.UUID, serverURL string) (*ports.CreateQRCodeResponse, error) {
	link, err := ls.GetByID(ctx, issuerDID, linkID, serverURL)
	if err != nil {
		log.Error(ctx, "cannot fetch the link", "err", err)
		return nil, err
	}

	err = ls.Validate(ctx, link)
	if err != nil {
		return nil, err
	}

	var authorizationRequestMessage *protocol.AuthorizationRequestMessage
	var raw []byte
	if err := json.Unmarshal(link.AuthorizationRequestMessage.Bytes, &authorizationRequestMessage); err != nil {
		log.Error(ctx, "cannot unmarshal the authorization", "err", err)
		return nil, err
	}
	raw = link.AuthorizationRequestMessage.Bytes
	return &ports.CreateQRCodeResponse{
		DeepLink:      qrlink.NewDeepLink(serverURL, linkID, &issuerDID),
		UniversalLink: qrlink.NewUniversal(ls.cfg.BaseUrl, serverURL, link.ID, &issuerDID),
		QrID:          link.ID,
		Link:          link,
		QrCodeRaw:     string(raw),
	}, nil
}

// IssueOrFetchClaim - Create a new claim
func (ls *Link) IssueOrFetchClaim(ctx context.Context, issuerDID w3c.DID, userDID w3c.DID, linkID uuid.UUID, hostURL string) (*protocol.CredentialsOfferMessage, error) {
	link, err := ls.linkRepository.GetByID(ctx, issuerDID, linkID)
	if err != nil {
		log.Error(ctx, "cannot fetch the link", "err", err)
		return nil, err
	}

	issuedByUser, err := ls.claimRepository.GetClaimsIssuedForUser(ctx, ls.storage.Pgx, issuerDID, userDID, linkID)
	if err != nil {
		log.Error(ctx, "cannot fetch the claims issued for the user", "err", err, "issuerDID", issuerDID, "userDID", userDID)
		return nil, err
	}

	if err := ls.Validate(ctx, link); err != nil {
		log.Error(ctx, "cannot Validate the link", "err", err)
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
		identity, err := ls.identityService.GetByDID(ctx, issuerDID)
		if err != nil {
			log.Error(ctx, "cannot fetch the identity", "err", err)
			return nil, err
		}
		credentialStatusType := verifiable.CredentialStatusType(identity.AuthCoreClaimRevocationStatus.Type)
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
			credentialStatusType,
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
		credOffer, err := notifications.NewOfferMsg(fmt.Sprintf(ports.AgentUrl, hostURL), credentialIssued)
		return credOffer, err
	} else {
		if credentialIssued.MTPProof.Bytes != nil {
			credOffer, err := notifications.NewOfferMsg(fmt.Sprintf(ports.AgentUrl, hostURL), credentialIssued)
			return credOffer, err
		}
		log.Info(ctx, "credential issued without MTP proof. Publishing state have to be done", "credential", credentialIssued.ID.String())
		return nil, nil
	}
}

// ProcessCallBack - process the callback.
func (ls *Link) ProcessCallBack(ctx context.Context, issuerID w3c.DID, message string, linkID uuid.UUID, hostURL string) (*protocol.CredentialsOfferMessage, error) {
	link, err := ls.linkRepository.GetByID(ctx, issuerID, linkID)
	if err != nil {
		log.Error(ctx, "error fetching the link from the database", "err", err)
		return nil, err
	}

	var authenticationRequest protocol.AuthorizationRequestMessage
	if err := json.Unmarshal(link.AuthorizationRequestMessage.Bytes, &authenticationRequest); err != nil {
		log.Error(ctx, "error unmarshaling the authorization request", "err", err)
		return nil, err
	}

	arm, err := ls.identityService.AuthenticateWithRequest(ctx, nil, authenticationRequest, message, hostURL)
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

	offer, err := ls.IssueOrFetchClaim(ctx, *issuerDID, *userDID, linkID, hostURL)
	if err != nil {
		log.Error(ctx, "error issuing claim", "err", err)
		return nil, err
	}
	return offer, nil
}

// Validate - validate the link
// It checks if the link is active, not expired and has not exceeded the maximum number of claims
func (ls *Link) Validate(ctx context.Context, link *domain.Link) error {
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
	return json_schema.ValidateCredentialSubject(ctx, ls.loader, schemaDB.URL, schemaDB.Type, cSubject)
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
