package mt

import (
	"context"
	"errors"
	"fmt"

	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-merkletree-sql/v2"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

const (
	// MerkleTreeTypeClaims is merkle tree type for claims tree
	MerkleTreeTypeClaims = 0
	// MerkleTreeTypeRevocations is merkle tree type for revocations tree
	MerkleTreeTypeRevocations = 1
	// MerkleTreeTypeRoots is merkle tree type for roots tree
	MerkleTreeTypeRoots = 2
	mtTypesCount        = 3
)

type IdentityMerkleTrees struct {
	identifier *core.DID
	Trees      []*merkletree.MerkleTree
	ImtModels  []*domain.IdentityMerkleTree
}

var (
	errorMsgNotCreated = errors.New("identity merkle trees were not created")
	mtTypes            = []uint16{MerkleTreeTypeClaims, MerkleTreeTypeRevocations, MerkleTreeTypeRoots}
)

// AddEntry adds claim to claims merkle tree
func (imts *IdentityMerkleTrees) AddEntry(ctx context.Context, entry *merkletree.Entry) error {
	if len(imts.Trees) < mtTypesCount {
		return errorMsgNotCreated
	}

	var entryData string
	marshalEntry := func() string {
		if entryData == "" {
			data, err := entry.MarshalText()
			if err != nil {
				return "<cannot marshal entry>"
			}
			entryData = string(data)
		}
		return entryData
	}

	index, value, err := entry.HiHv()
	if err != nil {
		fmt.Printf("cannot get Index and Value from entry: %v", marshalEntry())
		return fmt.Errorf("cannot get Index and Value from entry: %w", err)
	}

	err = imts.Trees[MerkleTreeTypeClaims].Add(ctx, index.BigInt(), value.BigInt())
	if err != nil {
		fmt.Printf("cannot add entry to claims merkle tree: %v", marshalEntry())
		return fmt.Errorf("cannot add entry to claims merkle tree: %w", err)
	}
	return nil
}

// ClaimsTree returns claims merkle tree
func (imts *IdentityMerkleTrees) ClaimsTree() (*merkletree.MerkleTree, error) {
	if len(imts.Trees) < mtTypesCount {
		return nil, errorMsgNotCreated
	}
	return imts.Trees[MerkleTreeTypeClaims], nil
}

// BindToIdentifier swaps temporary identifier for real one in IdentityMerkleTree models
func (imts *IdentityMerkleTrees) BindToIdentifier(conn db.Querier, identifier *core.DID) error {
	if imts.identifier != nil {
		return errors.New("can't change not empty identifier")
	}
	if len(imts.ImtModels) < mtTypesCount {
		return errorMsgNotCreated
	}
	imts.identifier = identifier
	for _, mtType := range mtTypes {
		imts.ImtModels[mtType].Identifier = identifier.String()
	}
	return nil
}

func (imts *IdentityMerkleTrees) GetMtModels() []*domain.IdentityMerkleTree {
	result := make([]*domain.IdentityMerkleTree, 0)
	for _, mtType := range mtTypes {
		result = append(result, imts.ImtModels[mtType])
	}
	return result
}
