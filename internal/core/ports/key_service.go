package ports

import (
	"context"
	"errors"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/kms"
)

var (
	// ErrInvalidKeyType is returned when the key type is invalid
	ErrInvalidKeyType = errors.New("invalid key type")
	// ErrAuthCredentialNotRevoked is returned when the associated auth core claim is not revoked
	ErrAuthCredentialNotRevoked = errors.New("associated auth core claim not revoked")
	// ErrKeyNotFound is returned when the key is not found
	ErrKeyNotFound = errors.New("key not found")
	// ErrKeyAssociatedWithIdentity is returned when the key is associated with an identity
	ErrKeyAssociatedWithIdentity = errors.New("key is associated with an identity")
)

// KMSKey is the struct that represents a key
type KMSKey struct {
	KeyID                       string
	KeyType                     kms.KeyType
	PublicKey                   string
	HasAssociatedAuthCredential bool
}

// KeyFilter is the filter to use when getting keys
type KeyFilter struct {
	MaxResults uint // Max number of results to return on each call.
	Page       uint // Page number to return. First is 1.
}

// KeyService is the service that manages keys
type KeyService interface {
	CreateKey(ctx context.Context, did *w3c.DID, keyType kms.KeyType) (kms.KeyID, error)
	Get(ctx context.Context, did *w3c.DID, keyID string) (*KMSKey, error)
	GetAll(ctx context.Context, did *w3c.DID, filter KeyFilter) ([]*KMSKey, uint, error)
	Delete(ctx context.Context, did *w3c.DID, keyID string) error
}
