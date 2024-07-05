package protocol

import (
	"context"
	"crypto"
	"io"

	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/utils"

	"github.com/polygonid/sh-id-platform/internal/kms"
)

// KMSBJJJWSAdapter is adapter to sign with BJJ key and KMS for JWS token creation
type KMSBJJJWSAdapter struct {
	kms   *kms.KMS
	keyID kms.KeyID
}

// Public returns public bjj  key as crypto key
func (a KMSBJJJWSAdapter) Public() crypto.PublicKey {
	pb, _ := a.kms.PublicKey(a.keyID)
	return pb
}

// Sign signs prepared digest (swap, because there is a swap inside)
func (a KMSBJJJWSAdapter) Sign(_ io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	sig, err := a.kms.Sign(context.Background(), a.keyID, utils.SwapEndianness(digest))
	if err != nil {
		return nil, err
	}
	var comp babyjub.SignatureComp
	copy(comp[:], sig)

	return comp.MarshalText()
}
