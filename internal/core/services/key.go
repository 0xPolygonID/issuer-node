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

// Key is the service that manages keys
type Key struct {
	kms          *kms.KMS
	claimService ports.ClaimService
}

// NewKey creates a new Key
func NewKey(kms *kms.KMS, claimService ports.ClaimService) ports.KeyService {
	return &Key{
		kms:          kms,
		claimService: claimService,
	}
}

// CreateKey creates a new key for the given DID
func (ks *Key) CreateKey(_ context.Context, did *w3c.DID, keyType kms.KeyType) (kms.KeyID, error) {
	return ks.kms.CreateKey(keyType, did)
}

// Get returns the public key for the given keyID
func (ks *Key) Get(ctx context.Context, did *w3c.DID, keyID string) (*ports.KMSKey, error) {
	keyType, err := getKeyType(keyID)
	if err != nil {
		log.Error(ctx, "failed to get key type", "err", err)
		return nil, err
	}

	kmsKeyID := kms.KeyID{
		ID:   keyID,
		Type: keyType,
	}

	exists, err := ks.kms.Exists(ctx, kmsKeyID)
	if err != nil {
		log.Error(ctx, "failed to check if key exists", "err", err)
		return nil, err
	}

	if !exists {
		return nil, ports.ErrKeyNotFound
	}

	publicKey, err := ks.getPublicKey(ctx, keyID)
	if err != nil {
		log.Error(ctx, "failed to get public key", "err", err)
		return nil, ports.ErrKeyNotFound
	}

	authCoreClaim, err := ks.claimService.GetAuthCredentialWithPublicKey(ctx, did, publicKey)
	if err != nil {
		log.Error(ctx, "failed to check if key has associated auth credential", "err", err)
		return nil, err
	}
	return &ports.KMSKey{
		KeyID:                      keyID,
		KeyType:                    keyType,
		PublicKey:                  hexutil.Encode(publicKey),
		HasAssociatedAuthCoreClaim: authCoreClaim != nil,
	}, nil
}

// GetAll returns all the keys for the given DID
func (ks *Key) GetAll(ctx context.Context, did *w3c.DID) ([]*ports.KMSKey, error) {
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

// Delete deletes the key with the given keyID
func (ks *Key) Delete(ctx context.Context, did *w3c.DID, keyID string) error {
	publicKey, err := ks.getPublicKey(ctx, keyID)
	if err != nil {
		log.Error(ctx, "failed to get public key", "err", err)
		return err
	}
	authCredential, err := ks.claimService.GetAuthCredentialWithPublicKey(ctx, did, publicKey)
	if err != nil {
		log.Error(ctx, "failed to check if key has associated auth credential", "err", err)
		return err
	}

	if authCredential != nil {
		log.Info(ctx, "can not be deleted because it has an associated auth credential. Have to check revocation status")
		revStatus, err := ks.claimService.GetRevocationStatus(ctx, *did, uint64(authCredential.RevNonce))
		if err != nil {
			log.Error(ctx, "failed to get revocation status", "err", err)
			return err
		}

		if revStatus != nil && !revStatus.MTP.Existence {
			log.Info(ctx, "auth credential is non revoked. Can not be deleted")
			return ports.ErrAuthCredentialNotRevoked
		}
	}
	keyType, err := getKeyType(keyID)
	if err != nil {
		log.Error(ctx, "failed to get key type", "err", err)
	}
	kmsKeyID := kms.KeyID{
		ID:   keyID,
		Type: keyType,
	}
	return ks.kms.Delete(ctx, kmsKeyID)
}

// getPublicKey returns the public key for the given keyID
func (ks *Key) getPublicKey(ctx context.Context, keyID string) ([]byte, error) {
	keyType, err := getKeyType(keyID)
	if err != nil {
		log.Error(ctx, "failed to get key type", "err", err)
		return nil, err
	}
	kmsKeyID := kms.KeyID{
		ID:   keyID,
		Type: keyType,
	}

	return ks.kms.PublicKey(kmsKeyID)
}

// getKeyType returns the key type for the given keyID
func getKeyType(keyID string) (kms.KeyType, error) {
	var keyType kms.KeyType
	if strings.Contains(keyID, "BJJ") {
		keyType = kms.KeyTypeBabyJubJub
	} else if strings.Contains(keyID, "ETH") {
		keyType = kms.KeyTypeEthereum
	} else {
		return keyType, ports.ErrInvalidKeyType
	}

	return keyType, nil
}
