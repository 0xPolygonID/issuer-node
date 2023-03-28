package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

// Link - represents a link in the issuer node
type Link struct {
	linkRepository ports.LinkRepository
}

// NewLinkService - constructor
func NewLinkService(linkRepository ports.LinkRepository) ports.LinkService {
	return &Link{linkRepository: linkRepository}
}

// Save - save a new credential
func (ls *Link) Save(ctx context.Context, did core.DID, maxIssuance *int, validUntil *time.Time, schemaID uuid.UUID, credentialExpiration *time.Time, credentialSignatureProof bool, credentialMTPProof bool, credentialAttributes []domain.CredentialAttributes) (*domain.Link, error) {
	link := domain.NewLink(did, maxIssuance, validUntil, schemaID, credentialExpiration, credentialSignatureProof, credentialMTPProof, credentialAttributes)

	_, err := ls.linkRepository.Save(ctx, link)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (ls *Link) GetByID(ctx context.Context, issuerID core.DID, id uuid.UUID) (*domain.Link, error) {
	return ls.linkRepository.GetByID(ctx, id)
}

// Delete - delete a link by id
func (ls *Link) Delete(ctx context.Context, id uuid.UUID, did core.DID) error {
	return ls.linkRepository.Delete(ctx, id, did)
}
