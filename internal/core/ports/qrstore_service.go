package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
)

const (
	// QRStoreUrl is the URL to the QR store service
	QRStoreUrl = "iden3comm://?request_uri=%s/v2/qr-store?id=%s"

	// QRStoreUrlWithDID is the URL to the QR store service with the issuer DID
	QRStoreUrlWithDID = "iden3comm://?request_uri=%s/v2/qr-store?id=%s&issuer=%s"

	// UniversalLinkURL - is the URL to the Universal Link
	UniversalLinkURL = "%s#request_uri=%s/v2/qr-store?id=%s"

	// UniversalLinkURLWithDID - is the URL to the Universal Link with the issuer DID
	UniversalLinkURLWithDID = "%s#request_uri=%s/v2/qr-store?id=%s&issuer=%s"
)

// QrStoreService is the interface that provides methods to store and retrieve the body of QR codes and to provide support
// to the QR url shortener functionality.
type QrStoreService interface {
	Find(ctx context.Context, id uuid.UUID) ([]byte, error)
	Store(ctx context.Context, qrCode []byte, ttl time.Duration) (uuid.UUID, error)
	ToDeepLink(hostURL string, id uuid.UUID, issuerDID *w3c.DID) string
	ToUniversalLink(ULinkBaseUrl string, hostURL string, id uuid.UUID, issuerDID *w3c.DID) string
}
