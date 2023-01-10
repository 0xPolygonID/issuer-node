package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type ClaimsRepository interface {
	Save(ctx context.Context, conn db.Querier, claim *domain.Claim) (uuid.UUID, error)
	Revoke(ctx context.Context, conn db.Querier, revocation *domain.Revocation) error
	GetByRevocationNonce(ctx context.Context, conn db.Querier, identifier *core.DID, revocationNonce domain.RevNonceUint64) (*domain.Claim, error)
}
