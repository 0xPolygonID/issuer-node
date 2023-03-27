package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

var (
	// ErrLinkAlreadyActive link is already active
	ErrLinkAlreadyActive = errors.New("link is already active")
	// ErrLinkAlreadyInactive link is already inactive
	ErrLinkAlreadyInactive = errors.New("link is already inactive")
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
	_, err = ls.linkRepository.Save(ctx, link)
	return err
}
