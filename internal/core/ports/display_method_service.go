package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/sqltools"
)

const (
	// DisplayMethodCreatedAtFilterField is the field used to filter display methods by creation date
	DisplayMethodCreatedAtFilterField sqltools.SQLFieldName = "created_at"
	// DisplayMethodNameFilterField is the field used to filter display methods by update date
	DisplayMethodNameFilterField sqltools.SQLFieldName = "name"
	// DisplayMethodTypeFilterField is the field used to filter display methods by type
	DisplayMethodTypeFilterField = "type"
)

// DisplayMethodFilter is the filter for display methods
type DisplayMethodFilter struct {
	MaxResults uint
	Page       uint
	OrderBy    sqltools.OrderByFilters
}

// DisplayMethodService is the interface implemented by the display method service
type DisplayMethodService interface {
	Save(ctx context.Context, issuerDID w3c.DID, name, url string, dtype *string) (*uuid.UUID, error)
	Update(ctx context.Context, identityDID w3c.DID, id uuid.UUID, name, url, dtype *string) (*uuid.UUID, error)
	GetByID(ctx context.Context, identityDID w3c.DID, id uuid.UUID) (*domain.DisplayMethod, error)
	GetAll(ctx context.Context, identityDID w3c.DID, filter DisplayMethodFilter) ([]domain.DisplayMethod, uint, error)
	Delete(ctx context.Context, identityDID w3c.DID, id uuid.UUID) error
}
