package ports

import (
	"context"
	"errors"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/kms"
)

// ErrInvalidKeyType is returned when the key type is invalid
var ErrInvalidKeyType = errors.New("invalid key type")

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
}
