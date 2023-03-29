package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/jsonschema"
	"github.com/polygonid/sh-id-platform/internal/loader"
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
	_, err = ls.linkRepository.Save(ctx, link)
	if err != nil {
		return nil, err
	}
	return link, nil
}

// Delete - delete a link by id
func (ls *Link) Delete(ctx context.Context, id uuid.UUID) error {
	return ls.linkRepository.Delete(ctx, id)
}
