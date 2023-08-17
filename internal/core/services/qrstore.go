package services

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/polygonid/sh-id-platform/pkg/cache"
)

// defaultTTL is the default time to live for a QR code.
const defQRUrlShortenerTTL = 30 * 24 * time.Hour

var ErrQRCodeLinkNotFound = errors.New("qr code link not found")

type QrStoreService struct {
	mx    sync.Mutex
	ttl   time.Duration
	store cache.Cache
}

type payload struct {
	QrCode string `json:"qr_code"`
}

func NewQrStoreService(store cache.Cache) *QrStoreService {
	return &QrStoreService{
		store: store,
	}
}

func (s *QrStoreService) Find(ctx context.Context, id uuid.UUID) ([]byte, error) {
	var raw payload
	if found := s.store.Get(ctx, s.key(id), &raw); !found {
		return nil, ErrQRCodeLinkNotFound
	}
	return []byte(raw.QrCode), nil
}

func (s *QrStoreService) Store(ctx context.Context, qrCode []byte, ttl time.Duration) (uuid.UUID, error) {
	id := s.newID(ctx)
	return id, s.store.Set(ctx, s.key(id), &payload{QrCode: string(qrCode)}, ttl)
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
