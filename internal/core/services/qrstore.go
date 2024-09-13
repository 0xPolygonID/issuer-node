package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/cache"
)

// DefaultQRBodyTTL is the default time to live for a QRcode body
const DefaultQRBodyTTL = 30 * 24 * time.Hour

// ErrQRCodeLinkNotFound is the error returned when a QR code link is not found in the QR storage
var ErrQRCodeLinkNotFound = errors.New("qr code link not found")

// QrStoreService implements the ports.QrStoreService interface.
// It provides methods to store and retrieve the body of QR codes and to provide support
// to the QR url shortener functionality
type QrStoreService struct {
	mx    sync.Mutex
	store cache.Cache
}

type payload struct {
	QrCode string `json:"qr_code"`
}

// NewQrStoreService creates a new QrStoreService instance.
func NewQrStoreService(store cache.Cache) *QrStoreService {
	return &QrStoreService{
		store: store,
	}
}

// Find retrieves the body of a QR code. Not finding an item is considered an error
func (s *QrStoreService) Find(ctx context.Context, id uuid.UUID) ([]byte, error) {
	var raw payload
	if found := s.store.Get(ctx, s.key(id), &raw); !found {
		log.Error(ctx, "qr code body not found. Tip: Recreate the Qr code again", "id", id.String())
		return nil, ErrQRCodeLinkNotFound
	}
	return []byte(raw.QrCode), nil
}

// Store stores the body of a QR code, creating a new unique ID for it and returning it.
func (s *QrStoreService) Store(ctx context.Context, qrCode []byte, ttl time.Duration) (uuid.UUID, error) {
	id := s.newID(ctx)
	if err := s.store.Set(ctx, s.key(id), payload{QrCode: string(qrCode)}, ttl); err != nil {
		log.Error(ctx, "error storing qr code body", "id", id.String(), "error", err, "qrCode", string(qrCode))
		return uuid.Nil, err
	}
	return id, nil
}

// ToDeepLink constructs a deeplink that will be used to get the body of a QR code.
func (s *QrStoreService) ToDeepLink(hostURL string, id uuid.UUID, issuerDID *w3c.DID) string {
	if issuerDID != nil {
		return fmt.Sprintf(ports.QRStoreUrlWithDID, hostURL, id.String(), issuerDID.String())
	}

	return fmt.Sprintf(ports.QRStoreUrl, hostURL, id.String())
}

// ToUniversalLink constructs a universal link
func (s *QrStoreService) ToUniversalLink(uLinkBaseUrl string, hostURL string, id uuid.UUID, issuerDID *w3c.DID) string {
	if issuerDID != nil {
		return fmt.Sprintf(ports.UniversalLinkURLWithDID, uLinkBaseUrl, hostURL, id.String(), issuerDID.String())
	}
	return fmt.Sprintf(ports.UniversalLinkURL, uLinkBaseUrl, hostURL, id.String())
}

func (s *QrStoreService) key(id uuid.UUID) string {
	return "issuer-node:qr-code:" + id.String()
}

// newID generates a new unique ID for a QR code.
func (s *QrStoreService) newID(ctx context.Context) uuid.UUID {
	s.mx.Lock()
	defer s.mx.Unlock()
	for {
		id := uuid.New()
		if !s.store.Exists(ctx, s.key(id)) {
			return id
		}
	}
}
