package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// SchemaRepository interface that define repo methods for schemas
type SchemaRepository interface {
	Save(ctx context.Context, schema *domain.Schema) error
	GetByID(ctx context.Context, issuerDID w3c.DID, id uuid.UUID) (*domain.Schema, error)
	GetAll(ctx context.Context, issuerDID w3c.DID, filter SchemasFilter) ([]domain.Schema, uint, error)
}
