package ports

import (
	"context"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/kms"
)

// KMSKey is the struct that represents a key
type KMSKey struct {
	KeyID                       string
	KeyType                     kms.KeyType
	PublicKey                   string
	HasAssociatedAuthCredential bool
	Name                        string
}

// KeyFilter is the filter to use when getting keys
type KeyFilter struct {
	MaxResults uint // Max number of results to return on each call.
	Page       uint // Page number to return. First is 1.
}

// KeyService is the service that manages keys
type KeyService interface {
	Create(ctx context.Context, did *w3c.DID, keyType kms.KeyType, name string) (kms.KeyID, error)
	Update(ctx context.Context, did *w3c.DID, keyID string, name string) error
	Get(ctx context.Context, did *w3c.DID, keyID string) (*KMSKey, error)
	GetAll(ctx context.Context, did *w3c.DID, filter KeyFilter) ([]*KMSKey, uint, error)
	Delete(ctx context.Context, did *w3c.DID, keyID string) error
}
