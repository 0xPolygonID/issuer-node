package signer

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/pkg/errors"
)

// SignatureSuite encapsulates signature suite methods required for signing information to circuit.
type SignatureSuite interface {
	// Accept registers this signature suite with the given signature type
	Accept(signatureType string) bool
	// Sign will sign data and return signature
	Sign(ctx context.Context, data []byte) ([]byte, error)

	// GetDigest returns
	GetDigest(data []byte) ([]byte, error)
}

// CircuitSigner implements signing of JSONLD documents.
type CircuitSigner struct {
	signatureSuites []SignatureSuite
}

// Context can be used for proof creation
type Context struct {
	SignatureType       string                  `json:"signature_type"`
	SignatureHex        string                  `json:"signature_hex"`
	Creator             *core.DID               `json:"creator,omitempty"`
	Created             int64                   `json:"created,omitempty"`
	Domain              string                  `json:"domain,omitempty"`
	Nonce               []byte                  `json:"nonce,omitempty"`
	VerificationMethod  string                  `json:"verification_method"`
	Challenge           string                  `json:"challenge,omitempty"`
	Purpose             verifiable.ProofPurpose `json:"purpose,omitempty"`
	IssuerMTP           *merkletree.Proof       `json:"issuer_mtp"`
	IssuerAuthClaim     *core.Claim             `json:"issuer_auth_claim,omitempty"`
	RevocationStatusURL string                  `json:"revocation_status_url,omitempty"`
}

// New returns new instance of circuit signer.
func New(signatureSuites ...SignatureSuite) *CircuitSigner {
	return &CircuitSigner{signatureSuites: signatureSuites}
}

// Sign returns SignatureProof for circuit verification
func (signer *CircuitSigner) Sign(sigType string, claim *core.Claim) ([]byte, error) {

	hashIndex, hashValue, err := claim.HiHv()
	if err != nil {
		return nil, err
	}

	commonHash, err := poseidon.Hash([]*big.Int{hashIndex, hashValue})
	if err != nil {
		return nil, err
	}

	suite, err := signer.getSignatureSuite(sigType)
	if err != nil {
		return nil, err
	}

	return suite.Sign(context.Background(), merkletree.SwapEndianness(commonHash.Bytes()))
}

// getSignatureSuite returns signature suite based on signature type.
func (signer *CircuitSigner) getSignatureSuite(
	signatureType string) (SignatureSuite, error) {
	for _, s := range signer.signatureSuites {
		if s.Accept(signatureType) {
			return s, nil
		}
	}

	return nil, fmt.Errorf("signature type %s not supported", signatureType)
}

// BJJSignatureFromHexString converts hex to  babyjub.Signature
func BJJSignatureFromHexString(sigHex string) (*babyjub.Signature, error) {
	signatureBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var sig [64]byte
	copy(sig[:], signatureBytes)
	bjjSig, err := new(babyjub.Signature).Decompress(sig)
	return bjjSig, errors.WithStack(err)
}
