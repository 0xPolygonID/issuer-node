package ports

import (
	"context"
	"errors"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/kms"
)

// ErrInvalidKeyType is returned when the key type is invalid
var (
	// ErrInvalidKeyType is returned when the key type is invalid
	ErrInvalidKeyType = errors.New("invalid key type")
	// ErrAuthCoreClaimNotRevoked is returned when the associated auth core claim is not revoked
	ErrAuthCoreClaimNotRevoked = errors.New("associated auth core claim not revoked")
	// ErrKeyNotFound is returned when the key is not found
	ErrKeyNotFound = errors.New("key not found")
)

// KMSKey is the struct that represents a key
type KMSKey struct {
	KeyID                      string
	KeyType                    kms.KeyType
	PublicKey                  string
	HasAssociatedAuthCoreClaim bool
}

// KeyService is the service that manages keys
type KeyService interface {
	CreateKey(ctx context.Context, did *w3c.DID, keyType kms.KeyType) (kms.KeyID, error)
	Get(ctx context.Context, did *w3c.DID, keyID string) (*KMSKey, error)
	GetAll(ctx context.Context, did *w3c.DID) ([]*KMSKey, error)
	Delete(ctx context.Context, did *w3c.DID, keyID string) error
}
