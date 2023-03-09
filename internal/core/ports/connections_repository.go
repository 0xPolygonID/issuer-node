package ports

import (
	"context"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ConnectionsRepository defines the available methods for connections repository
type ConnectionsRepository interface {
	Save(ctx context.Context, conn db.Querier, connection *domain.Connection) error
}
