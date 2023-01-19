package reverse_hash

import (
	"context"
	"errors"

	"github.com/iden3/go-merkletree-sql/v2"
	proof "github.com/iden3/merkletree-proof"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/log"
)

// stateHashes - handle hashes states.
type stateHashes struct {
	State  merkletree.Hash
	Claims merkletree.Hash
	Rev    merkletree.Hash
	Roots  merkletree.Hash
}

// RhsPublisher defines reverse hash publisher functions.
type RhsPublisher interface {
	PushHashesToRHS(ctx context.Context, newState, prevState *domain.IdentityState, revocations []*domain.Revocation, trees *domain.IdentityMerkleTrees) error
}

type rhsPublisher struct {
	rhsCli          *proof.HTTPReverseHashCli
	ignoreRHSErrors bool
}

// NewRhsPublisher - constructor
func NewRhsPublisher(rhsCli *proof.HTTPReverseHashCli, ignoreRHSErrors bool) *rhsPublisher {
	return &rhsPublisher{
		rhsCli:          rhsCli,
		ignoreRHSErrors: ignoreRHSErrors,
	}
}

// PushHashesToRHS pushes following changes to reverse hash service:
//   - all revocations with their parents up to revocations tree root;
//   - new state node hash with children trees' roots.
//   - if claim's tree root is changed, also send new claim's tree root with
//     its parents up to RoR tree root.
func (rhsp *rhsPublisher) PushHashesToRHS(ctx context.Context, newState, prevState *domain.IdentityState, revocations []*domain.Revocation, trees *domain.IdentityMerkleTrees) error {
	// TODO - check rhs cli
	// if Reverse-Hash-Service is not configure, do nothing.
	//if i.rhsCli == nil {
	//	return nil
	//}
	nb := newNodesBuilder()

	// add revocation nodes
	err := nb.addRevocationNodes(ctx, trees, revocations)
	if err != nil {
		return err
	}

	prevStateHashes, err := newStateHashesFromModel(prevState)
	if err != nil {
		return err
	}

	newStateHashes, err := newStateHashesFromModel(newState)
	if err != nil {
		return err
	}

	// if claims tree root is changed, add its RoR tree subtree to RHS
	if prevStateHashes.Claims != newStateHashes.Claims {
		err = nb.addRoRNode(ctx, trees)
		if err != nil {
			return err
		}
	}

	// add new state node
	if newStateHashes.State != merkletree.HashZero {
		nb.addProofNode(proof.Node{
			Hash: &newStateHashes.State,
			Children: []*merkletree.Hash{
				&newStateHashes.Claims,
				&newStateHashes.Rev,
				&newStateHashes.Roots,
			},
		})
	}

	if nb.numberOfNodes() > 0 {
		// todo: call rhs
		// err = i.rhsCli.SaveNodes(ctx, nb.nodes)
		log.Info(ctx, "new state nodes", nb.nodes)
		err = nil
	}
	return err
}

func newStateHashesFromModel(inState *domain.IdentityState) (stateHashes, error) {
	if inState == nil {
		return stateHashes{}, errors.New("nil state")
	}

	var err error
	var outState stateHashes
	if inState.State != nil {
		outState.State, err = HashFromString(inState.State)
		if err != nil {
			return stateHashes{}, err
		}
	}
	if inState.ClaimsTreeRoot != nil {
		outState.Claims, err = HashFromString(inState.ClaimsTreeRoot)
		if err != nil {
			return stateHashes{}, err
		}
	}

	if inState.RevocationTreeRoot != nil {
		outState.Rev, err = HashFromString(inState.RevocationTreeRoot)
		if err != nil {
			return stateHashes{}, err
		}
	}

	if inState.RootOfRoots != nil {
		outState.Roots, err = HashFromString(inState.RootOfRoots)
		if err != nil {
			return stateHashes{}, err
		}
	}

	expectedState, err := merkletree.HashElems(outState.Claims.BigInt(), outState.Rev.BigInt(), outState.Roots.BigInt())
	if err != nil {
		return stateHashes{}, err
	}
	if *expectedState != outState.State {
		return stateHashes{}, errors.New("state hash is incorrect")
	}

	return outState, nil
}

// HashFromString - crearte a merkle trees hash from a string.
func HashFromString(s *string) (merkletree.Hash, error) {
	var hash merkletree.Hash
	if s == nil {
		return hash, nil
	}

	h, err := merkletree.NewHashFromHex(*s)
	if err != nil {
		return hash, err
	}

	return *h, nil
}
