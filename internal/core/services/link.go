package services

import (
	"context"

	"github.com/google/uuid"

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
func (ls *Link) Save(ctx context.Context, link domain.Link) (*uuid.UUID, error) {
	// TODO: make attributes type validations.
	return ls.linkRepository.Save(ctx, &link)
}
