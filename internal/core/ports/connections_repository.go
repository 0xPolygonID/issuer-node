package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ConnectionsRepository defines the available methods for connections repository
type ConnectionsRepository interface {
	Save(ctx context.Context, conn db.Querier, connection *domain.Connection) (uuid.UUID, error)
	Delete(ctx context.Context, conn db.Querier, id uuid.UUID) error
	GetByIDAndIssuerID(ctx context.Context, conn db.Querier, id uuid.UUID, issuerID core.DID) (*domain.Connection, error)
}
