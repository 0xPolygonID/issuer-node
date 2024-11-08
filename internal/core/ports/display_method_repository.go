package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// DisplayMethodRepository is the interface implemented by the display method repository
type DisplayMethodRepository interface {
	Save(ctx context.Context, displayMethod domain.DisplayMethod) (*uuid.UUID, error)
	GetByID(ctx context.Context, identityDID w3c.DID, id uuid.UUID) (*domain.DisplayMethod, error)
	GetAll(ctx context.Context, identityDID w3c.DID) ([]domain.DisplayMethod, error)
	Delete(ctx context.Context, identityDID w3c.DID, id uuid.UUID) error
	GetDefault(ctx context.Context, identityDID w3c.DID) (*domain.DisplayMethod, error)
}
