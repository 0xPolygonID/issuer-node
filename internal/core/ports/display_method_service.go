package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// DisplayMethodService is the interface implemented by the display method service
type DisplayMethodService interface {
	Save(ctx context.Context, issuerDID w3c.DID, name, url string, isDefault bool) (*uuid.UUID, error)
	Update(ctx context.Context, identityDID w3c.DID, id uuid.UUID, name, url *string, isDefault *bool) (*uuid.UUID, error)
	GetByID(ctx context.Context, identityDID w3c.DID, id uuid.UUID) (*domain.DisplayMethod, error)
	GetAll(ctx context.Context, identityDID w3c.DID) ([]domain.DisplayMethod, error)
	Delete(ctx context.Context, identityDID w3c.DID, id uuid.UUID) error
	GetDefault(ctx context.Context, identityDID w3c.DID) (*domain.DisplayMethod, error)
}
