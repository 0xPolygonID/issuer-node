package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ConnectionsRepository defines the available methods for connections repository
type ConnectionsRepository interface {
	Save(ctx context.Context, conn db.Querier, connection *domain.Connection) (uuid.UUID, error)
	Delete(ctx context.Context, conn db.Querier, id uuid.UUID) error
}
