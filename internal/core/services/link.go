package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

var (
	// ErrLinkAlreadyActive link is already active
	ErrLinkAlreadyActive = errors.New("link is already active")
	// ErrLinkAlreadyInactive link is already inactive
	ErrLinkAlreadyInactive = errors.New("link is already inactive")
)

// Link - represents a link in the issuer node
type Link struct {
	linkRepository   ports.LinkRepository
	schemaRepository ports.SchemaRepository
	loaderFactory    loader.Factory
}

// NewLinkService - constructor
func NewLinkService(linkRepository ports.LinkRepository, schemaRepository ports.SchemaRepository, loaderFactory loader.Factory) ports.LinkService {
	return &Link{
		linkRepository:   linkRepository,
		schemaRepository: schemaRepository,
		loaderFactory:    loaderFactory,
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
	schema, err := ls.schemaRepository.GetByID(ctx, schemaID)
	if err != nil {
		return nil, err
	}
	link := domain.NewLink(did, maxIssuance, validUntil, schemaID, credentialExpiration, credentialSignatureProof, credentialMTPProof)

	if err := link.ProcessAttributes(ctx, ls.loaderFactory(schema.URL), credentialAttributes); err != nil {
		return nil, err
	}
	_, err = ls.linkRepository.Save(ctx, link)
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
	_, err = ls.linkRepository.Save(ctx, link)
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

// Delete - delete a link by id
func (ls *Link) Delete(ctx context.Context, id uuid.UUID, did core.DID) error {
	return ls.linkRepository.Delete(ctx, id, did)
}
