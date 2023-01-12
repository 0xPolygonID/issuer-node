package ports

import (
	"context"

	"github.com/polygonid/sh-id-platform/internal/core/domain"

	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/db"
)

type MtService interface {
	CreateIdentityMerkleTrees(ctx context.Context, conn db.Querier) (*domain.IdentityMerkleTrees, error)
	GetIdentityMerkleTrees(ctx context.Context, conn db.Querier, identifier *core.DID) (*domain.IdentityMerkleTrees, error)
}
