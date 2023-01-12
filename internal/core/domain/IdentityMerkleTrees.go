package domain

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-merkletree-sql/v2"

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
	Identifier *core.DID
	Trees      []*merkletree.MerkleTree
	ImtModels  []*IdentityMerkleTree
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
		fmt.Println(fmt.Sprintf("cannot get Index and Value from entry: %v", marshalEntry()))
		return fmt.Errorf("cannot get Index and Value from entry: %w", err)
	}

	err = imts.Trees[MerkleTreeTypeClaims].Add(ctx, index.BigInt(), value.BigInt())
	if err != nil {
		fmt.Println(fmt.Printf("cannot add entry to claims merkle tree: %v", marshalEntry()))
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

// BindToIdentifier swaps temporary Identifier for real one in IdentityMerkleTree models
func (imts *IdentityMerkleTrees) BindToIdentifier(conn db.Querier, identifier *core.DID) error {
	if imts.Identifier != nil {
		return errors.New("can't change not empty Identifier")
	}
	if len(imts.ImtModels) < mtTypesCount {
		return errorMsgNotCreated
	}
	imts.Identifier = identifier
	for _, mtType := range mtTypes {
		imts.ImtModels[mtType].Identifier = identifier.String()
	}
	return nil
}

func (imts *IdentityMerkleTrees) GetMtModels() []*IdentityMerkleTree {
	result := make([]*IdentityMerkleTree, 0)
	for _, mtType := range mtTypes {
		result = append(result, imts.ImtModels[mtType])
	}
	return result
}

func (imts *IdentityMerkleTrees) RevokeClaim(ctx context.Context, revNonce *big.Int) error {
	// Now it is hardcoded version 0, but later on, it could be changed when
	// we introduce more cases with versioning
	err := imts.Trees[MerkleTreeTypeRevocations].Add(ctx, revNonce, big.NewInt(0))
	if err != nil {
		return fmt.Errorf("cannot add revocation nonce: %d to revocation merkle tree: %w", revNonce, err)
	}
	return nil
}
