package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/iden3comm/packers"
	"github.com/iden3/iden3comm/protocol"
	"github.com/jackc/pgx/v4"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	linkState "github.com/polygonid/sh-id-platform/pkg/link"
)

var (
	// ErrLinkAlreadyActive link is already active
	ErrLinkAlreadyActive = errors.New("link is already active")
	// ErrLinkAlreadyInactive link is already inactive
	ErrLinkAlreadyInactive = errors.New("link is already inactive")
	// ErrLinkAlreadyExpired - link already expired
	ErrLinkAlreadyExpired = errors.New("can not issue a credential for an expired link")
	// ErrLinkMaxExceeded - link max exceeded
	ErrLinkMaxExceeded = errors.New("can not issue a credential for an expired link")
	// ErrClaimAlreadyIssued - claim already issued
	ErrClaimAlreadyIssued = errors.New("the claim was already issued for the user")
)

// Link - represents a link in the issuer node
type Link struct {
	storage          *db.Storage
	claimsService    ports.ClaimsService
	claimRepository  ports.ClaimsRepository
	linkRepository   ports.LinkRepository
	schemaRepository ports.SchemaRepository
	loaderFactory    loader.Factory
	sessionManager   ports.SessionRepository
}

// NewLinkService - constructor
func NewLinkService(storage *db.Storage, claimsService ports.ClaimsService, claimRepository ports.ClaimsRepository, linkRepository ports.LinkRepository, schemaRepository ports.SchemaRepository, loaderFactory loader.Factory, sessionManager ports.SessionRepository) ports.LinkService {
	return &Link{
		storage:          storage,
		claimsService:    claimsService,
		claimRepository:  claimRepository,
		linkRepository:   linkRepository,
		schemaRepository: schemaRepository,
		loaderFactory:    loaderFactory,
		sessionManager:   sessionManager,
	}
}

// Save - save a new credential
func (ls *Link) Save(
	ctx context.Context,
	did core.DID,
	maxIssuance *int,
	validUntil *time.Time,
	schemaID uuid.UUID,
	credentialExpiration *time.Time,
	credentialSignatureProof bool,
	credentialMTPProof bool,
	credentialAttributes []domain.CredentialAttrsRequest,
) (*domain.Link, error) {
	schema, err := ls.schemaRepository.GetByID(ctx, did, schemaID)
	if err != nil {
		return nil, err
	}
	link := domain.NewLink(did, maxIssuance, validUntil, schemaID, credentialExpiration, credentialSignatureProof, credentialMTPProof)
	if err := link.ProcessAttributes(ctx, ls.loaderFactory(schema.URL), credentialAttributes); err != nil {
		return nil, err
	}
	_, err = ls.linkRepository.Save(ctx, ls.storage.Pgx, link)
	if err != nil {
		return nil, err
	}
	link.Schema = schema
	return link, nil
}

// Activate - activates or deactivates a credential link
func (ls *Link) Activate(ctx context.Context, issuerID core.DID, linkID uuid.UUID, active bool) error {
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
func (ls *Link) GetByID(ctx context.Context, issuerID core.DID, id uuid.UUID) (*domain.Link, error) {
	link, err := ls.linkRepository.GetByID(ctx, issuerID, id)
	if errors.Is(err, repositories.ErrLinkDoesNotExist) {
		return nil, ErrLinkNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := link.LoadAttributeTypes(ctx, ls.loaderFactory(link.Schema.URL)); err != nil {
		return nil, err
	}
	return link, nil
}

// GetAll returns all links from issueDID of type lType filtered by query string
func (ls *Link) GetAll(ctx context.Context, issuerDID core.DID, status ports.LinkStatus, query *string) ([]domain.Link, error) {
	return ls.linkRepository.GetAll(ctx, issuerDID, status, query)
}

// Delete - delete a link by id
func (ls *Link) Delete(ctx context.Context, id uuid.UUID, did core.DID) error {
	return ls.linkRepository.Delete(ctx, id, did)
}

// CreateQRCode - generates a qr code for a link
func (ls *Link) CreateQRCode(ctx context.Context, issuerDID core.DID, linkID uuid.UUID, serverURL string) (*ports.CreateQRCodeResponse, error) {
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
		},
	}

	err := ls.sessionManager.Set(ctx, sessionID, *qrCode)
	if err != nil {
		return nil, err
	}

	err = ls.sessionManager.SetLink(ctx, linkState.CredentialStateCacheKey(linkID.String(), sessionID), *linkState.NewStatePending())
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	link, err := ls.GetByID(ctx, issuerDID, linkID)
	if err != nil {
		return nil, err
	}
	return &ports.CreateQRCodeResponse{
		SessionID: sessionID,
		QrCode:    qrCode,
		Link:      link,
	}, nil
}

// IssueClaim - Create a new claim
func (ls *Link) IssueClaim(ctx context.Context, sessionID string, issuerDID core.DID, userDID core.DID, linkID uuid.UUID, hostURL string) error {
	link, err := ls.linkRepository.GetByID(ctx, issuerDID, linkID)
	if err != nil {
		log.Error(ctx, "cannot fetch the link", "error", err)
		return err
	}

	issuedByUser, err := ls.claimRepository.GetClaimsIssuedForUser(ctx, ls.storage.Pgx, issuerDID, userDID, linkID)
	if err != nil {
		log.Error(ctx, "cannot fetch the claims issued for the user", "err", err, "issuerDID", issuerDID, "userDID", userDID)
		return err
	}

	if len(issuedByUser) > 0 {
		log.Info(ctx, "the claim was already issued for the user", "user DID", userDID.String())
		return ErrClaimAlreadyIssued
	}

	if err := ls.validate(ctx, sessionID, link, linkID); err != nil {
		return err
	}

	schema, err := ls.schemaRepository.GetByID(ctx, issuerDID, link.SchemaID)
	if err != nil {
		log.Error(ctx, "can not fetch the schema", err)
		return err
	}

	credentialSubject := make(map[string]any)

	credentialSubject["id"] = userDID.String()
	for _, credAtt := range link.CredentialAttributes {
		credentialSubject[credAtt.Name] = credAtt.Value
	}

	claimReq := ports.NewCreateClaimRequest(&issuerDID,
		schema.URL,
		credentialSubject,
		link.CredentialExpiration,
		schema.Type,
		nil, nil, nil,
		common.ToPointer(link.CredentialSignatureProof),
		common.ToPointer(link.CredentialMTPProof),
		&linkID,
	)

	credentialIssued, err := ls.claimsService.CreateCredential(ctx, claimReq)
	if err != nil {
		log.Error(ctx, "Can not create the claim", err.Error())
		return err
	}

	var credentialIssuedID uuid.UUID
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
		return err
	}
	credentialIssued.ID = credentialIssuedID

	r := &linkState.QRCodeMessage{
		ID:       uuid.NewString(),
		Typ:      "application/iden3comm-plain-json",
		Type:     linkState.CredentialOfferMessageType,
		ThreadID: uuid.NewString(),
		Body: linkState.CredentialsLinkMessageBody{
			URL: fmt.Sprintf("%s/v1/agent", hostURL),
			Credentials: []linkState.CredentialLink{{
				ID:          credentialIssued.ID.String(),
				Description: schema.Type,
			}},
		},
		From: issuerDID.String(),
		To:   userDID.String(),
	}

	if link.CredentialSignatureProof {
		err = ls.sessionManager.SetLink(ctx, linkState.CredentialStateCacheKey(linkID.String(), sessionID), *linkState.NewStateDone(r))
	} else {
		err = ls.sessionManager.SetLink(ctx, linkState.CredentialStateCacheKey(linkID.String(), sessionID), *linkState.NewStatePendingPublish())
	}

	if err != nil {
		log.Error(ctx, "can not set the sate", err)
		return err
	}

	return nil
}

// GetQRCode - return the link qr code.
func (ls *Link) GetQRCode(ctx context.Context, sessionID uuid.UUID, issuerID core.DID, linkID uuid.UUID) (*ports.GetQRCodeResponse, error) {
	link, err := ls.GetByID(ctx, issuerID, linkID)
	if err != nil {
		log.Error(ctx, "error fetching the link from the database", "error", err)
		return nil, err
	}

	linkStateInCache, err := ls.sessionManager.GetLink(ctx, linkState.CredentialStateCacheKey(linkID.String(), sessionID.String()))
	if err != nil {
		log.Error(ctx, "error fetching the link state from the cache", "error", err)
		return nil, err
	}
	return &ports.GetQRCodeResponse{
		State: &linkStateInCache,
		Link:  link,
	}, nil
}

func (ls *Link) validate(ctx context.Context, sessionID string, link *domain.Link, linkID uuid.UUID) error {
	if link.ValidUntil != nil && time.Now().UTC().After(*link.ValidUntil) {
		log.Debug(ctx, "can not issue a credential for an expired link")
		err := ls.sessionManager.SetLink(ctx, linkState.CredentialStateCacheKey(linkID.String(), sessionID), *linkState.NewStateError(ErrLinkAlreadyExpired))
		if err != nil {
			log.Error(ctx, "can not set the sate", err)
			return err
		}

		return ErrLinkAlreadyExpired
	}

	if link.MaxIssuance != nil && *link.MaxIssuance <= link.IssuedClaims {
		log.Debug(ctx, "can not dispatch more claims for this link")
		err := ls.sessionManager.SetLink(ctx, linkState.CredentialStateCacheKey(linkID.String(), sessionID), *linkState.NewStateError(ErrLinkMaxExceeded))
		if err != nil {
			log.Error(ctx, "can not set the sate", err)
			return err
		}

		return ErrLinkMaxExceeded
	}

	return nil
}
