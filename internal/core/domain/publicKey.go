package domain

import (
	"bytes"

	"github.com/iden3/go-iden3-crypto/babyjub"
)

const bjjAuthSchemaURL = "https://schema.iden3.io/core/json/auth.json"

// PublicKey - defines the interface for public keys
type PublicKey interface {
	Equal([]byte) bool
}

type bjjPublicKey struct {
	publicKey babyjub.PublicKey
}

// newBJJPublicKey creates a new PublicKey from a Claim
func newBJJPublicKey(claim Claim) PublicKey {
	entry := claim.CoreClaim.Get()
	bjjClaim := entry.RawSlotsAsInts()
	var authCoreClaimPublicKey babyjub.PublicKey
	authCoreClaimPublicKey.X = bjjClaim[2]
	authCoreClaimPublicKey.Y = bjjClaim[3]
	return bjjPublicKey{publicKey: authCoreClaimPublicKey}
}

func (b bjjPublicKey) Equal(pubKey []byte) bool {
	compPubKey := b.publicKey.Compress()
	return bytes.Equal(pubKey, compPubKey[:])
}

type unSupportedPublicKeyType struct{}

func (u unSupportedPublicKeyType) Equal([]byte) bool {
	return false
}
