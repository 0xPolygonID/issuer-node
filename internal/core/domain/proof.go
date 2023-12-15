package domain

import (
	"math/big"

	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/issuer-node/internal/common"
)

const AnyProofType verifiable.ProofType = "AnyProof" // AnyProofType defines any proof type

// ZKProof is structure that represents SnarkJS library result of proof generation
type ZKProof struct {
	A        []string   `json:"pi_a"`
	B        [][]string `json:"pi_b"`
	C        []string   `json:"pi_c"`
	Protocol string     `json:"protocol"`
}

// FullProof is ZKP proof with public signals
type FullProof struct {
	Proof      *ZKProof `json:"proof"`
	PubSignals []string `json:"pub_signals"`
}

// ProofToBigInts transforms a zkp (that uses `*bn256.G1` and `*bn256.G2`) into
// `*big.Int` format, to be used for example in snarkjs solidity verifiers.
func (p *ZKProof) ProofToBigInts() (a []*big.Int, b [][]*big.Int, c []*big.Int, err error) {
	a, err = common.ArrayStringToBigInt(p.A)
	if err != nil {
		return nil, nil, nil, err
	}
	b = make([][]*big.Int, len(p.B))
	for i, v := range p.B {
		b[i], err = common.ArrayStringToBigInt(v)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	c, err = common.ArrayStringToBigInt(p.C)
	if err != nil {
		return nil, nil, nil, err
	}

	return a, b, c, nil
}
