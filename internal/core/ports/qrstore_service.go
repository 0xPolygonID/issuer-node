package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	// QRStoreUrl is the URL to the QR store service
	QRStoreUrl = "iden3comm://?request_uri=%s/v2/qr-store?id=%s"
	// UniversalLinkURL - is the URL to the Universal Link
	UniversalLinkURL = "%s#request_uri=%s/v2/qr-store?id=%s"
)

// QrStoreService is the interface that provides methods to store and retrieve the body of QR codes and to provide support
// to the QR url shortener functionality.
type QrStoreService interface {
	Find(ctx context.Context, id uuid.UUID) ([]byte, error)
	Store(ctx context.Context, qrCode []byte, ttl time.Duration) (uuid.UUID, error)
	ToDeepLink(hostURL string, id uuid.UUID) string
	ToUniversalLink(ULinkBaseUrl string, hostURL string, id uuid.UUID) string
}
