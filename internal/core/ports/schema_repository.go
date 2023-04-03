package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// SchemaRepository interface that define repo methods for schemas
type SchemaRepository interface {
	Save(ctx context.Context, schema *domain.Schema) error
	GetByID(ctx context.Context, issuerDID core.DID, id uuid.UUID) (*domain.Schema, error)
	GetAll(ctx context.Context, issuerDID core.DID, query *string) ([]domain.Schema, error)
}
