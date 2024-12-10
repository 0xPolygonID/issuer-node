package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// KeyRepository key repository interface
type KeyRepository interface {
	Save(ctx context.Context, conn db.Querier, key *domain.Key) (uuid.UUID, error)
	GetByPublicKey(ctx context.Context, issuerDID w3c.DID, publicKey string) (*domain.Key, error)
	Delete(ctx context.Context, issuerDID w3c.DID, publicKey string) error
}
