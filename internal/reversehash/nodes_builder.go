package reversehash

import (
	"context"
	"errors"
	"math/big"

	"github.com/iden3/go-merkletree-sql/v2"
	proof "github.com/iden3/merkletree-proof"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// append nodes to list without duplication
type nodesBuilder struct {
	nodes []proof.Node
	seen  map[merkletree.Hash]struct{}
}

// newNodesBuilder - constructor
func newNodesBuilder() *nodesBuilder {
	return &nodesBuilder{
		nodes: make([]proof.Node, 0),
		seen:  make(map[merkletree.Hash]struct{}),
	}
}

func (nb *nodesBuilder) numberOfNodes() int {
	return len(nb.nodes)
}

func (nb *nodesBuilder) addRevocationNodes(ctx context.Context, trees *domain.IdentityMerkleTrees, revocations []*domain.Revocation) error {
	if len(revocations) == 0 {
		return nil
	}

	revTree, err := trees.RevsTree()
	if err != nil {
		return err
	}

	// Prepare revocation nodes to push
	for _, r := range revocations {
		err = nb.addKey(ctx, revTree, new(big.Int).SetUint64(uint64(r.Nonce)))
		if err != nil {
			return err
		}
	}

	return nil
}

func (nb *nodesBuilder) addRoRNode(ctx context.Context, trees *domain.IdentityMerkleTrees) error {
	currentRootsTree, err := trees.RootsTree()
	if err != nil {
		return err
	}

	claimsTree, err := trees.ClaimsTree()
	if err != nil {
		return err
	}

	return nb.addKey(ctx, currentRootsTree, claimsTree.Root().BigInt())
}

func (nb *nodesBuilder) addProofNode(node proof.Node) {
	if _, ok := nb.seen[*node.Hash]; !ok {
		nb.nodes = append(nb.nodes, node)
		nb.seen[*node.Hash] = struct{}{}
	}
}

func (nb *nodesBuilder) addKey(ctx context.Context, tree *merkletree.MerkleTree, nodeKey *big.Int) error {
	_, nodeValue, siblings, err := tree.Get(ctx, nodeKey)
	if err != nil {
		return err
	}

	nodeKeyHash, err := merkletree.NewHashFromBigInt(nodeKey)
	if err != nil {
		return err
	}

	nodeValueHash, err := merkletree.NewHashFromBigInt(nodeValue)
	if err != nil {
		return err
	}

	node := merkletree.NewNodeLeaf(nodeKeyHash, nodeValueHash)
	newNodes, err := buildNodesUp(siblings, node)
	if err != nil {
		return err
	}

	for _, n := range newNodes {
		if _, ok := nb.seen[*n.Hash]; !ok {
			nb.nodes = append(nb.nodes, n)
			nb.seen[*n.Hash] = struct{}{}
		}
	}

	return nil
}

func buildNodesUp(siblings []*merkletree.Hash, node *merkletree.Node) ([]proof.Node, error) {
	if node.Type != merkletree.NodeTypeLeaf {
		return nil, errors.New("node is not a leaf")
	}

	prevHash, err := node.Key()
	if err != nil {
		return nil, err
	}
	nodes := make([]proof.Node, len(siblings)+1)
	nodes[len(siblings)].Hash = prevHash
	hashOfOne, err := merkletree.NewHashFromBigInt(big.NewInt(1))
	if err != nil {
		return nil, err
	}
	nodes[len(siblings)].Children = []*merkletree.Hash{
		node.Entry[0], node.Entry[1], hashOfOne,
	}

	pathKey := node.Entry[0][:]
	hashArraySize := 2
	for i := len(siblings) - 1; i >= 0; i-- {
		isRight := merkletree.TestBit(pathKey, uint(i))
		var err error
		nodes[i].Children = make([]*merkletree.Hash, hashArraySize)
		if isRight {
			nodes[i].Children[0] = siblings[i]
			nodes[i].Children[1] = prevHash
		} else {
			nodes[i].Children[0] = prevHash
			nodes[i].Children[1] = siblings[i]
		}
		nodes[i].Hash, err = merkletree.HashElems(
			nodes[i].Children[0].BigInt(),
			nodes[i].Children[1].BigInt())
		if err != nil {
			return nil, err
		}
		prevHash = nodes[i].Hash
	}

	return nodes, nil
}
