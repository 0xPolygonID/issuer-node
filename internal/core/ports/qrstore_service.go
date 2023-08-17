package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type QrStoreService interface {
	Find(ctx context.Context, id uuid.UUID) ([]byte, error)
	Store(ctx context.Context, qrCode []byte, ttl time.Duration) (uuid.UUID, error)
}
