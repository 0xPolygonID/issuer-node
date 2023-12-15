package ports

import (
	"context"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/issuer-node/internal/core/domain"
	"github.com/polygonid/issuer-node/internal/db"
)

// MtService is the interface that defines the MT Methods
type MtService interface {
	CreateIdentityMerkleTrees(ctx context.Context, conn db.Querier) (*domain.IdentityMerkleTrees, error)
	GetIdentityMerkleTrees(ctx context.Context, conn db.Querier, identifier *w3c.DID) (*domain.IdentityMerkleTrees, error)
}
