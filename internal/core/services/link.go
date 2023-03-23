package services

import (
	"context"
	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"time"

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
func (ls *Link) Save(ctx context.Context, ID uuid.UUID, did core.DID, maxIssuance *int, validUntil *time.Time, schemaID uuid.UUID,
	credentialExpiration *time.Time, credentialSignatureProof bool, credentialMTPProof bool, credentialAttributes []domain.CredentialAttributes) (*domain.Link, error) {
	link := domain.NewLink(ID, did, maxIssuance, validUntil, schemaID, credentialExpiration, credentialSignatureProof, credentialMTPProof, credentialAttributes)

	_, err := ls.linkRepository.Save(ctx, link)
	if err != nil {
		return nil, err
	}
	return link, nil
}
