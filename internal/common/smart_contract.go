package common

import (
	"math/big"

	"github.com/iden3/go-merkletree-sql/v2"
)

// SmartContractProof is a proof returned by smart contract
type SmartContractProof struct {
	Root         *big.Int
	Existence    bool
	Siblings     []*big.Int
	Index        *big.Int
	Value        *big.Int
	AuxExistence bool
	AuxIndex     *big.Int
	AuxValue     *big.Int
}

// SmartContractProofToMtProofAdapter converts SmartContractProof to merkletree.Proof
func SmartContractProofToMtProofAdapter(smtProof SmartContractProof) (*merkletree.Proof, error) {
	var (
		existence bool
		nodeAux   *merkletree.NodeAux
		err       error
	)

	if smtProof.Existence {
		existence = true
	} else {
		existence = false
		if smtProof.AuxExistence {
			nodeAux = &merkletree.NodeAux{}
			nodeAux.Key, err = merkletree.NewHashFromBigInt(smtProof.AuxIndex)
			if err != nil {
				return nil, err
			}
			nodeAux.Value, err = merkletree.NewHashFromBigInt(smtProof.AuxValue)
			if err != nil {
				return nil, err
			}
		}
	}

	allSiblings := make([]*merkletree.Hash, len(smtProof.Siblings))
	for i, s := range smtProof.Siblings {
		sh, err := merkletree.NewHashFromBigInt(s)
		if err != nil {
			return nil, err
		}
		allSiblings[i] = sh
	}

	proof, err := merkletree.NewProofFromData(existence, allSiblings, nodeAux)
	if err != nil {
		return nil, err
	}

	return proof, nil
}
