package domain

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-merkletree-sql/v2"

	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/log"
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

// IdentityMerkleTrees defines the merkle tree structure
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
		log.Error(ctx, "not enough merkle trees", "err", errorMsgNotCreated, "count", len(imts.Trees))
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
		log.Error(ctx, "cannot get HiHv values", "err", err, "entry", marshalEntry())
		return fmt.Errorf("cannot get HiHv index, values: %w", err)
	}

	err = imts.Trees[MerkleTreeTypeClaims].Add(ctx, index.BigInt(), value.BigInt())
	if err != nil {
		log.Error(ctx, "adding entry to claims merkle tree", "err", err, "entry", marshalEntry())
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

// GetMtModels returns a list of IdentityMerkleTree
func (imts *IdentityMerkleTrees) GetMtModels() []*IdentityMerkleTree {
	result := make([]*IdentityMerkleTree, 0)
	for _, mtType := range mtTypes {
		result = append(result, imts.ImtModels[mtType])
	}
	return result
}

// RevokeClaim - revoke a claim per a given nonce.
func (imts *IdentityMerkleTrees) RevokeClaim(ctx context.Context, revNonce *big.Int) error {
	// Now it is hardcoded version 0, but later on, it could be changed when
	// we introduce more cases with versioning
	err := imts.Trees[MerkleTreeTypeRevocations].Add(ctx, revNonce, big.NewInt(0))
	if err != nil {
		return fmt.Errorf("cannot add revocation nonce: %d to revocation merkle tree: %w", revNonce, err)
	}
	return nil
}

// GenerateRevocationProof generates the proof of existence (or non-existence) of an nonce in RevocationTree
func (imts *IdentityMerkleTrees) GenerateRevocationProof(ctx context.Context, nonce *big.Int, root *merkletree.Hash) (*merkletree.Proof, error) {
	proof, _, err := imts.Trees[MerkleTreeTypeRevocations].GenerateProof(ctx, nonce, root)
	if err != nil {
		return nil, fmt.Errorf("cannot generate revocation proof: %w", err)
	}
	return proof, nil
}

// AddClaim adds a Claim into the MerkleTree
func (imts *IdentityMerkleTrees) AddClaim(ctx context.Context, c *Claim) error {
	if len(imts.Trees) < mtTypesCount {
		return errorMsgNotCreated
	}

	coreClaim := c.CoreClaim.Get()
	hi, hv, err := coreClaim.HiHv()
	if err != nil {
		return fmt.Errorf("error getting index and value: %w", err)
	}

	err = imts.Trees[MerkleTreeTypeClaims].Add(ctx, hi, hv)
	if err != nil {
		return fmt.Errorf("cannot add entry to claims merkle tree: %w", err)
	}

	return nil
}

// RevsTree returns revocations merkle tree
func (imts *IdentityMerkleTrees) RevsTree() (*merkletree.MerkleTree, error) {
	if len(imts.Trees) < mtTypesCount {
		return nil, errorMsgNotCreated
	}
	return imts.Trees[MerkleTreeTypeRevocations], nil
}

// RootsTree returns roots merkle tree
func (imts *IdentityMerkleTrees) RootsTree() (*merkletree.MerkleTree, error) {
	if len(imts.Trees) < mtTypesCount {
		return nil, errorMsgNotCreated
	}
	return imts.Trees[MerkleTreeTypeRoots], nil
}
