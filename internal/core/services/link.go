package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/polygonid/sh-id-platform/internal/db"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/iden3comm/packers"
	"github.com/iden3/iden3comm/protocol"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/jsonschema"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	linkState "github.com/polygonid/sh-id-platform/pkg/link"
)

var (
	// ErrLinkAlreadyActive link is already active
	ErrLinkAlreadyActive = errors.New("link is already active")
	// ErrLinkAlreadyInactive link is already inactive
	ErrLinkAlreadyInactive = errors.New("link is already inactive")
	// ErrLinkAlreadyExpired - link already expired
	ErrLinkAlreadyExpired = errors.New("can not issue a credential for an expired link")
)

// Link - represents a link in the issuer node
type Link struct {
	storage          *db.Storage
	claimsService    ports.ClaimsService
	linkRepository   ports.LinkRepository
	schemaRepository ports.SchemaRepository
	loaderFactory    loader.Factory
	sessionManager   ports.SessionRepository
}

// NewLinkService - constructor
func NewLinkService(storage *db.Storage, claimsService ports.ClaimsService, linkRepository ports.LinkRepository, schemaRepository ports.SchemaRepository, loaderFactory loader.Factory, sessionManager ports.SessionRepository) ports.LinkService {
	return &Link{
		storage:          storage,
		claimsService:    claimsService,
		linkRepository:   linkRepository,
		schemaRepository: schemaRepository,
		loaderFactory:    loaderFactory,
		sessionManager:   sessionManager,
	}
}

// Save - save a new credential
func (ls *Link) Save(ctx context.Context, did core.DID, maxIssuance *int, validUntil *time.Time, schemaID uuid.UUID, credentialExpiration *time.Time, credentialSignatureProof bool, credentialMTPProof bool, credentialAttributes []domain.CredentialAttributes) (*domain.Link, error) {
	schema, err := ls.schemaRepository.GetByID(ctx, schemaID)
	if err != nil {
		return nil, err
	}

	remoteSchema, err := jsonschema.Load(ctx, ls.loaderFactory(schema.URL))
	if err != nil {
		return nil, ErrLoadingSchema
	}

	credentialAttributes, err = remoteSchema.ValidateAndConvert(credentialAttributes)
	if err != nil {
		return nil, err
	}

	link := domain.NewLink(did, maxIssuance, validUntil, schemaID, credentialExpiration, credentialSignatureProof, credentialMTPProof, credentialAttributes)
	_, err = ls.linkRepository.Save(ctx, ls.storage.Pgx, link)
	if err != nil {
		return nil, err
	}
	return link, nil
}

// Activate - activates or deactivates a credential link
func (ls *Link) Activate(ctx context.Context, linkID uuid.UUID, active bool) error {
	link, err := ls.linkRepository.GetByID(ctx, linkID)
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

// Delete - delete a link by id
func (ls *Link) Delete(ctx context.Context, id uuid.UUID, did core.DID) error {
	return ls.linkRepository.Delete(ctx, id, did)
}

// CreateQRCode - generates a qr code for a link
func (ls *Link) CreateQRCode(ctx context.Context, issuerDID core.DID, linkID uuid.UUID, serverURL string) (*protocol.AuthorizationRequestMessage, string, error) {
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
		return nil, "", err
	}

	err = ls.sessionManager.SetLink(ctx, linkState.CredentialStateCacheKey(linkID.String(), sessionID), linkState.NewStatePending().String())
	if err != nil {
		return nil, "", err
	}

	if err != nil {
		return nil, "", err
	}
	return qrCode, sessionID, err
}

// IssueClaim - Create a new claim
func (ls *Link) IssueClaim(ctx context.Context, sessionID string, issuerDID core.DID, userDID core.DID, linkID uuid.UUID, hostURL string) error {
	link, err := ls.linkRepository.GetByID(ctx, linkID)
	if err != nil {
		log.Error(ctx, "can not fetch the link", err)
		return err
	}

	if err := ls.validate(ctx, sessionID, link, linkID); err != nil {
		return err
	}

	schema, err := ls.schemaRepository.GetByID(ctx, link.SchemaID)
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
		common.ToPointer(link.CredentialExpiration.Unix()),
		schema.Type,
		nil, nil, nil,
		common.ToPointer(link.CredentialSignatureProof),
		common.ToPointer(link.CredentialMTPProof))

	claimIssued, err := ls.claimsService.CreateClaim(ctx, claimReq)
	if err != nil {
		log.Error(ctx, "Can not create the claim", err.Error())
		return err
	}

	r := &ports.LinkQRCodeMessage{
		ID:       uuid.NewString(),
		Typ:      "application/iden3comm-plain-json",
		Type:     ports.CredentialOfferMessageType,
		ThreadID: uuid.NewString(),
		Body: ports.CredentialsLinkMessageBody{
			URL: fmt.Sprintf("%s/api/v1/agent", hostURL),
			Credentials: []ports.CredentialLink{{
				ID:          claimIssued.ID.String(),
				Description: schema.Type,
			}},
		},
		From: issuerDID.String(),
		To:   userDID.String(),
	}

	err = ls.sessionManager.SetLink(ctx, linkState.CredentialStateCacheKey(linkID.String(), sessionID), linkState.NewStateDone(r))
	if err != nil {
		log.Error(ctx, "can not set the sate", err)
		return err
	}

	return nil
}

func (ls *Link) validate(ctx context.Context, sessionID string, link *domain.Link, linkID uuid.UUID) error {
	if link.ValidUntil != nil && time.Now().UTC().After(*link.ValidUntil) {
		log.Debug(ctx, "can not issue a credential for an expired link")
		err := ls.sessionManager.SetLink(ctx, linkState.CredentialStateCacheKey(linkID.String(), sessionID), linkState.NewStateError(ErrLinkAlreadyExpired).String())
		if err != nil {
			log.Error(ctx, "can not set the sate", err)
			return err
		}

		return ErrLinkAlreadyExpired
	}
	return nil
}
