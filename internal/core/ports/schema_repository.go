package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// SchemaRepository interface that define repo methods for schemas
type SchemaRepository interface {
	Save(ctx context.Context, schema *domain.Schema) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Schema, error)
}
