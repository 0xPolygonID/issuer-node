package babyjubjub

import (
	"math/big"

	cryptoUtils "github.com/iden3/go-iden3-crypto/utils"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/pkg/errors"

	"github.com/polygonid/issuer-node/pkg/credentials/signature/suite"
)

const (
	// SignatureType defines type of signature proof for claims
	SignatureType = "BJJSignature2021"
)

var errSpongeHashingIsNotImplemented = errors.New("data is not in the field. Sponge hashing is not supported")

// Suite is an instance of crypto suite for signatures and verifications
type Suite struct {
	suite.Suite
}

// New an instance of babyjubjub signature suite.
func New(opts ...suite.Opt) *Suite {
	s := &Suite{}
	suite.InitSuiteOptions(&s.Suite, opts...)
	return s
}

// Accept will accept only bjj data signature
func (s *Suite) Accept(t string) bool {
	return t == SignatureType
}

// GetDigest returns digest on provided data for signing
func (s *Suite) GetDigest(data []byte) ([]byte, error) {
	// check if data size more than 32 byte we need to do the hashing
	bi := new(big.Int).SetBytes(merkletree.SwapEndianness(data))

	if !cryptoUtils.CheckBigIntArrayInField([]*big.Int{bi}) {
		return nil, errSpongeHashingIsNotImplemented
	}
	hash, err := merkletree.HashElems(bi)
	if err != nil {
		return nil, err
	}

	return hash[:], nil
}
