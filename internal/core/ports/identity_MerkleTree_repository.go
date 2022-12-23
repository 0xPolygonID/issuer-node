package ports

import (
	"context"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type IdentityMerkleTreeRepository interface {
	Save(ctx context.Context, conn db.Querier, identifier string, mtType uint16) (*domain.IdentityMerkleTree, error)
	UpdateByID(ctx context.Context, conn db.Querier, imt *domain.IdentityMerkleTree) error
}
