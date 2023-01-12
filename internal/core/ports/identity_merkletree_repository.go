package ports

import (
	"context"

	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// IdentityMerkleTreeRepository is the interface that defines the available methods
type IdentityMerkleTreeRepository interface {
	Save(ctx context.Context, conn db.Querier, identifier string, mtType uint16) (*domain.IdentityMerkleTree, error)
	UpdateByID(ctx context.Context, conn db.Querier, imt *domain.IdentityMerkleTree) error
	GetByID(ctx context.Context, conn db.Querier, mtID uint64) (*domain.IdentityMerkleTree, error)
	GetByIdentifierAndTypes(ctx context.Context, conn db.Querier, identifier *core.DID, mtTypes []uint16) ([]domain.IdentityMerkleTree, error)
}
