package services

import (
	"context"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/kms"
)

type KeyManagementService struct {
	kms kms.KeyProvider
}

func NewKeyManagementService(kms kms.KeyProvider) *KeyManagementService {
	return &KeyManagementService{
		kms: kms,
	}
}

// CreateKey creates a new key for the given DID
func (kms *KeyManagementService) CreateKey(ctx context.Context, did *w3c.DID, keyType kms.KeyType) (kms.KeyID, error) {
	return kms.kms.New(did)
}
