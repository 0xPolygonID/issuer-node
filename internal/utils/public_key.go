package utils

import (
	"bytes"

	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// PublicKey - defines the interface for public keys
type PublicKey interface {
	Equal([]byte) bool
	String() string
}

type bjjPublicKey struct {
	publicKey babyjub.PublicKey
}

// newBJJPublicKey creates a new PublicKey from a Claim
func newBJJPublicKey(claim domain.Claim) PublicKey {
	entry := claim.CoreClaim.Get()
	bjjClaim := entry.RawSlotsAsInts()
	var authCoreClaimPublicKey babyjub.PublicKey
	authCoreClaimPublicKey.X = bjjClaim[2]
	authCoreClaimPublicKey.Y = bjjClaim[3]
	return &bjjPublicKey{publicKey: authCoreClaimPublicKey}
}

func (b *bjjPublicKey) Equal(pubKey []byte) bool {
	compPubKey := b.publicKey.Compress()
	return bytes.Equal(pubKey, compPubKey[:])
}

func (b *bjjPublicKey) String() string {
	return "0x" + b.publicKey.String()
}

type unSupportedPublicKeyType struct{}

func (u *unSupportedPublicKeyType) Equal([]byte) bool {
	return false
}

func (u *unSupportedPublicKeyType) String() string {
	return ""
}

// GetPublicKeyFromClaim returns the public key of the claim
// If the schema is not supported, it returns an unSupportedPublicKeyType
func GetPublicKeyFromClaim(c *domain.Claim) PublicKey {
	if c.SchemaURL == verifiable.JSONSchemaIden3AuthBJJCredential {
		return newBJJPublicKey(*c)
	}
	return &unSupportedPublicKeyType{}
}
