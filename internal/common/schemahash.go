package common

import (
	core "github.com/iden3/go-iden3-core/v2"
	"golang.org/x/crypto/sha3"
)

// keccak256 calculates the Keccak256 hash of the input data.
func keccak256(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// CreateSchemaHash computes schema hash from schemaID
func CreateSchemaHash(schemaID []byte) core.SchemaHash {
	var sHash core.SchemaHash
	h := keccak256(schemaID)
	copy(sHash[:], h[len(h)-16:])
	return sHash
}
