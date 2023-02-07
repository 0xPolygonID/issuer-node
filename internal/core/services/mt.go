package services

import (
	"context"
	"crypto/rand"
	"fmt"

	core "github.com/iden3/go-iden3-core"
	sql "github.com/iden3/go-merkletree-sql/db/pgx/v2"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

const (
	randomLength = 27
	// MerkleTreeTypeClaims is merkle tree type for claims tree
	MerkleTreeTypeClaims = 0
	// MerkleTreeTypeRevocations is merkle tree type for revocations tree
	MerkleTreeTypeRevocations = 1
	// MerkleTreeTypeRoots is merkle tree type for roots tree
	MerkleTreeTypeRoots = 2
	mtTypesCount        = 3

	mtDepth = 40
)

var (
	errNotFound = errors.New("not found")
	mtTypes     = []uint16{MerkleTreeTypeClaims, MerkleTreeTypeRevocations, MerkleTreeTypeRoots}
)

type mtService struct {
	imtRepo ports.IdentityMerkleTreeRepository
}

// NewIdentityMerkleTrees generates a new merkle tree service
func NewIdentityMerkleTrees(imtRepo ports.IdentityMerkleTreeRepository) ports.MtService {
	return &mtService{
		imtRepo: imtRepo,
	}
}

func (mts *mtService) CreateIdentityMerkleTrees(ctx context.Context, conn db.Querier) (*domain.IdentityMerkleTrees, error) {
	trees := make([]*merkletree.MerkleTree, mtTypesCount)
	imtModels := make([]*domain.IdentityMerkleTree, mtTypesCount)

	var buf [randomLength]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		return nil, err
	}

	tmpIdentifier := "tmp-" + base58.Encode(buf[:])

	for _, mtType := range mtTypes {
		imt, err := mts.imtRepo.Save(ctx, conn, tmpIdentifier, mtType)
		if err != nil {
			return nil, err
		}
		imtModels[mtType] = imt
		treeStorage := sql.NewSqlStorage(conn, imt.ID)
		var tree *merkletree.MerkleTree
		tree, err = merkletree.NewMerkleTree(ctx, treeStorage, mtDepth)
		if err != nil {
			return nil, err
		}
		trees[mtType] = tree
	}

	imts := &domain.IdentityMerkleTrees{
		Trees:     trees,
		ImtModels: imtModels,
	}
	return imts, nil
}

func (mts *mtService) GetIdentityMerkleTrees(ctx context.Context, conn db.Querier, identifier *core.DID) (*domain.IdentityMerkleTrees, error) {
	trees := make([]*merkletree.MerkleTree, mtTypesCount)
	imtModels := make([]*domain.IdentityMerkleTree, mtTypesCount)
	imts, err := mts.imtRepo.GetByIdentifierAndTypes(ctx, conn, identifier, mtTypes)
	if err != nil {
		return nil, fmt.Errorf("error getting merkle tree: %w", err)
	}

	for _, mtType := range mtTypes {
		imt := findByType(imts, mtType)
		if imt == nil {
			return nil, errNotFound
		}
		imtModels[mtType] = imt
		treeStorage := sql.NewSqlStorage(conn, imt.ID)
		tree, err := merkletree.NewMerkleTree(ctx, treeStorage, mtDepth)
		if err != nil {
			return nil, err
		}
		trees[mtType] = tree
	}

	imTrees := &domain.IdentityMerkleTrees{
		Identifier: identifier,
		Trees:      trees,
		ImtModels:  imtModels,
	}
	return imTrees, nil
}

func findByType(mts []domain.IdentityMerkleTree, tp uint16) *domain.IdentityMerkleTree {
	for i := range mts {
		if mts[i].Type == tp {
			return &mts[i]
		}
	}
	return nil
}
