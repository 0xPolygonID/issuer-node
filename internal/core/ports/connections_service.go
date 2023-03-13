package ports

import (
	"context"

	"github.com/google/uuid"
)

// ConnectionsService  is the interface implemented by the Connections service
type ConnectionsService interface {
	Delete(ctx context.Context, id uuid.UUID) error
}
