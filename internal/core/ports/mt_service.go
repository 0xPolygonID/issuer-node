package ports

import (
	"context"

	"github.com/polygonid/sh-id-platform/internal/core/mt"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type MtService interface {
	CreateIdentityMerkleTrees(ctx context.Context, conn db.Querier) (*mt.IdentityMerkleTrees, error)
}
