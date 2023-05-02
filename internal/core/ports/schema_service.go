package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// SchemaService defines the methods that Schema manager will expose.
type SchemaService interface {
	ImportSchema(ctx context.Context, issuerDID core.DID, url string, sType string) (*domain.Schema, error)
	GetByID(ctx context.Context, issuerDID core.DID, id uuid.UUID) (*domain.Schema, error)
	GetAll(ctx context.Context, issuerDID core.DID, query *string) ([]domain.Schema, error)
}
