package ports

import (
	"context"

	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// MtService is the interface that defines the MT Methods
type MtService interface {
	CreateIdentityMerkleTrees(ctx context.Context, conn db.Querier) (*domain.IdentityMerkleTrees, error)
	GetIdentityMerkleTrees(ctx context.Context, conn db.Querier, identifier *core.DID) (*domain.IdentityMerkleTrees, error)
}
