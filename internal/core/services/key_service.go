package services

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
)

// KeyService is the service that manages keys
type KeyService struct {
	kms          *kms.KMS
	claimService ports.ClaimService
}

// NewKeyService creates a new KeyService
func NewKeyService(kms *kms.KMS, claimService ports.ClaimService) ports.KeyService {
	return &KeyService{
		kms:          kms,
		claimService: claimService,
	}
}

// CreateKey creates a new key for the given DID
func (ks *KeyService) CreateKey(_ context.Context, did *w3c.DID, keyType kms.KeyType) (kms.KeyID, error) {
	return ks.kms.CreateKey(keyType, did)
}

// Get returns the public key for the given keyID
func (ks *KeyService) Get(ctx context.Context, did *w3c.DID, keyID string) (*ports.KMSKey, error) {
	var keyType kms.KeyType
	if strings.Contains(keyID, "BJJ") {
		keyType = kms.KeyTypeBabyJubJub
	} else if strings.Contains(keyID, "ETH") {
		keyType = kms.KeyTypeEthereum
	} else {
		return nil, ports.ErrInvalidKeyType
	}

	kmsKeyID := kms.KeyID{
		ID:   keyID,
		Type: keyType,
	}

	publicKey, err := ks.kms.PublicKey(kmsKeyID)
	if err != nil {
		log.Error(ctx, "failed to get public key", "err", err)
		return nil, err
	}

	authCoreClaims, err := ks.claimService.GetAuthCoreClaims(ctx, did)
	if err != nil {
		log.Error(ctx, "failed to get auth core claims", "err", err)
		return nil, err
	}

	ok := false
	for _, authCoreClaim := range authCoreClaims {
		if authCoreClaim.CoreClaim.HasPublicKey(publicKey) {
			ok = true
			break
		}
	}

	return &ports.KMSKey{
		KeyID:                      keyID,
		KeyType:                    keyType,
		PublicKey:                  hexutil.Encode(publicKey),
		HasAssociatedAuthCoreClaim: ok,
	}, nil
}

// GetAll returns all the keys for the given DID
func (ks *KeyService) GetAll(ctx context.Context, did *w3c.DID) ([]*ports.KMSKey, error) {
	keyIDs, err := ks.kms.KeysByIdentity(ctx, *did)
	if err != nil {
		log.Error(ctx, "failed to get keys", "err", err)
		return nil, err
	}

	keys := make([]*ports.KMSKey, len(keyIDs))
	for i, keyID := range keyIDs {
		key, err := ks.Get(ctx, did, keyID.ID)
		if err != nil {
			log.Error(ctx, "failed to get key", "err", err)
			return nil, err
		}
		keys[i] = key
	}
	return keys, nil
}
