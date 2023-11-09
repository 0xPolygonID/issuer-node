package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// QrStoreService is the interface that provides methods to store and retrieve the body of QR codes and to provide support
// to the QR url shortener functionality.
type QrStoreService interface {
	Find(ctx context.Context, id uuid.UUID) ([]byte, error)
	Store(ctx context.Context, qrCode []byte, ttl time.Duration) (uuid.UUID, error)
	ToURL(hostURL string, id uuid.UUID) string
}
