package primitive

import (
	"context"
	stderr "errors"
	"math/big"

	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/utils"
	"github.com/pkg/errors"

	"github.com/polygonid/sh-id-platform/internal/kms"
)

var (
	errorNotInitialized         = stderr.New("signer is not initialized")
	errorInvalidSignatureLength = stderr.New("incorrect signature length")
	errorInvalidSignature       = stderr.New("invalid signature")
	errorDecompress             = stderr.New("can't decompress bjj signature")
)

// BJJSinger represents signer with BJJ key
type BJJSinger struct {
	kms   kms.KMSType
	keyID kms.KeyID
}

// NewBJJSigner creates new instance oj BJJ signer
func NewBJJSigner(keyMS kms.KMSType, keyID kms.KeyID) (*BJJSinger, error) {
	if keyID.Type != kms.KeyTypeBabyJubJub {
		return nil, errors.New("wrong key type")
	}
	if keyID.ID == "" {
		return nil, errors.New("empty key ID")
	}
	if keyMS == nil {
		return nil, errors.New("KMS is nil")
	}
	return &BJJSinger{keyMS, keyID}, nil
}

// Sign signs prepared data ( value in field Q)
func (s *BJJSinger) Sign(ctx context.Context, data []byte) ([]byte, error) {
	if s.kms == nil {
		return nil, errors.WithStack(errorNotInitialized)
	}
	return s.kms.Sign(ctx, s.keyID, data)
}

// BJJVerifier represents verifier with BJJ key
type BJJVerifier struct{}

// Verify verifies BJJ signature on data
func (s *BJJVerifier) Verify(publicKey, data, signature []byte) error {
	var sigComp babyjub.SignatureComp
	if len(signature) != len(sigComp) {
		return errors.WithStack(errorInvalidSignatureLength)
	}

	copy(sigComp[:], signature)
	sig, err := sigComp.Decompress()
	if err != nil {
		return errors.WithStack(errorDecompress)
	}

	message := new(big.Int).SetBytes(utils.SwapEndianness(data))

	pub := babyjub.PublicKey{}
	err = pub.UnmarshalText(publicKey)
	if err != nil {
		return errors.WithStack(err)
	}
	valid := pub.VerifyPoseidon(message, sig)

	if !valid {
		return errors.WithStack(errorInvalidSignature)
	}
	return nil
}
