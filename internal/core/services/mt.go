package services

import (
	"context"
	"crypto/rand"
	goerr "errors"

	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
	"github.com/polygonid/sh-id-platform/internal/core/mt"

	sql "github.com/iden3/go-merkletree-sql/db/pgx/v2"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

const randomLength = 27

const (
	// MerkleTreeTypeClaims is merkle tree type for claims tree
	MerkleTreeTypeClaims = 0
	// MerkleTreeTypeRevocations is merkle tree type for revocations tree
	MerkleTreeTypeRevocations = 1
	// MerkleTreeTypeRoots is merkle tree type for roots tree
	MerkleTreeTypeRoots = 2
)

const mtTypesCount = 3

// TODO: move to config
const mtDepth = 40

var errorMsgNotCreated = goerr.New("identity merkle trees were not created")

var mtTypes = []uint16{MerkleTreeTypeClaims, MerkleTreeTypeRevocations, MerkleTreeTypeRoots}

type mtService struct {
	imtRepo ports.IdentityMerkleTreeRepository
}

func NewIdentityMerkleTrees(imtRepo ports.IdentityMerkleTreeRepository) ports.MtService {
	return &mtService{
		imtRepo: imtRepo,
	}
}

func (mts *mtService) CreateIdentityMerkleTrees(ctx context.Context, conn db.Querier) (*mt.IdentityMerkleTrees, error) {
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
			return nil, errors.WithStack(err)
		}
		trees[mtType] = tree
	}

	imts := &mt.IdentityMerkleTrees{
		Trees:     trees,
		ImtModels: imtModels,
	}
	return imts, nil
}
