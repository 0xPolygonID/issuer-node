package services

import (
	"context"
	b64 "encoding/base64"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
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
func (ks *Key) CreateKey(ctx context.Context, did *w3c.DID, keyType kms.KeyType) (kms.KeyID, error) {
	var keyID kms.KeyID
	var err error
	if keyType == kms.KeyTypeBabyJubJub {
		keyID, err = ks.kms.CreateKey(keyType, did)
		if err != nil {
			log.Error(ctx, "failed to create key", "err", err)
			return kms.KeyID{}, err
		}
	}

	if keyType == kms.KeyTypeEthereum {
		keyID, err = ks.kms.CreateKey(keyType, nil)
		if err != nil {
			log.Error(ctx, "failed to create key", "err", err)
			return kms.KeyID{}, err
		}
		keyID, err = ks.kms.LinkToIdentity(ctx, keyID, *did)
		if err != nil {
			log.Error(ctx, "failed to link key to identity", "err", err)
			return kms.KeyID{}, err
		}
	}

	encodedKeyID := b64.StdEncoding.EncodeToString([]byte(keyID.ID))
	log.Info(ctx, "key created successfully", "keyID", encodedKeyID)
	keyID.ID = encodedKeyID
	return keyID, nil
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

	hasAssociatedAuthCredential := false
	switch keyType {
	case kms.KeyTypeBabyJubJub:
		hasAssociatedAuthCredential, _, err = ks.hasAssociatedAuthCredential(ctx, did, publicKey)
		if err != nil {
			log.Error(ctx, "failed to check if key has associated auth credential", "err", err)
			return nil, err
		}
	case kms.KeyTypeEthereum:
		hasAssociatedAuthCredential, err = ks.isAssociatedWithIdentity(ctx, did, publicKey)
		if err != nil {
			log.Error(ctx, "failed to check if key has associated auth credential", "err", err)
			return nil, err
		}
	default:
		return nil, ports.ErrInvalidKeyType
	}

	return &ports.KMSKey{
		KeyID:                       keyID,
		KeyType:                     keyType,
		PublicKey:                   hexutil.Encode(publicKey),
		HasAssociatedAuthCredential: hasAssociatedAuthCredential,
	}, nil
}

// GetAll returns all the keys for the given DID
func (ks *Key) GetAll(ctx context.Context, did *w3c.DID, filter ports.KeyFilter) ([]*ports.KMSKey, uint, error) {
	keyIDs, err := ks.kms.KeysByIdentity(ctx, *did)
	if err != nil {
		log.Error(ctx, "failed to get keys", "err", err)
		return nil, 0, err
	}

	total := uint(len(keyIDs))
	start := (int(filter.Page) - 1) * int(filter.MaxResults)
	end := start + int(filter.MaxResults)

	if start >= len(keyIDs) {
		return []*ports.KMSKey{}, 0, nil
	}

	if end > len(keyIDs) {
		end = len(keyIDs)
	}

	sort.Slice(keyIDs, func(i, j int) bool {
		return keyIDs[i].ID < keyIDs[j].ID
	})

	keyIDs = keyIDs[start:end]
	keys := make([]*ports.KMSKey, len(keyIDs))
	for i, keyID := range keyIDs {
		key, err := ks.Get(ctx, did, keyID.ID)
		if err != nil {
			log.Error(ctx, "failed to get key", "err", err)
			return nil, 0, err
		}
		keys[i] = key
	}
	return keys, total, nil
}

// Delete deletes the key with the given keyID
func (ks *Key) Delete(ctx context.Context, did *w3c.DID, keyID string) error {
	keyType, err := getKeyType(keyID)
	if err != nil {
		log.Error(ctx, "failed to get key type", "err", err)
		return err
	}

	kmsKeyID := kms.KeyID{
		ID:   keyID,
		Type: keyType,
	}

	exists, err := ks.kms.Exists(ctx, kmsKeyID)
	if err != nil {
		log.Error(ctx, "failed to check if key exists", "err", err)
		return err
	}

	if !exists {
		return ports.ErrKeyNotFound
	}

	publicKey, err := ks.getPublicKey(ctx, keyID)
	if err != nil {
		log.Error(ctx, "failed to get public key", "err", err)
		return err
	}

	hasAssociatedAuthCoreCredential := false
	var authCredential *domain.Claim
	switch keyType {
	case kms.KeyTypeBabyJubJub:
		hasAssociatedAuthCoreCredential, authCredential, err = ks.hasAssociatedAuthCredential(ctx, did, publicKey)
		if err != nil {
			log.Error(ctx, "failed to check if key has associated auth credential", "err", err)
			return err
		}

		if hasAssociatedAuthCoreCredential {
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
	case kms.KeyTypeEthereum:
		hasAssociatedAuthCoreCredential, err = ks.isAssociatedWithIdentity(ctx, did, publicKey)
		if err != nil {
			log.Error(ctx, "failed to check if key has associated auth credential", "err", err)
			return err
		}
		if hasAssociatedAuthCoreCredential {
			log.Info(ctx, "can not be deleted because it is associated with the identity")
			return ports.ErrKeyAssociatedWithIdentity
		}
	default:
		return ports.ErrInvalidKeyType
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

// hasAssociatedAuthCredential checks if the bbj key has an associated auth credential
func (ks *Key) hasAssociatedAuthCredential(ctx context.Context, did *w3c.DID, publicKey []byte) (bool, *domain.Claim, error) {
	hasAssociatedAuthCredential := false
	authCredential, err := ks.claimService.GetAuthCredentialWithPublicKey(ctx, did, publicKey)
	if err != nil {
		log.Error(ctx, "failed to check if key has associated auth credential", "err", err)
		return false, nil, err
	}
	hasAssociatedAuthCredential = authCredential != nil
	return hasAssociatedAuthCredential, authCredential, nil
}

// isAssociatedWithIdentity checks if the eth key is associated with the identity
func (ks *Key) isAssociatedWithIdentity(ctx context.Context, did *w3c.DID, publicKey []byte) (bool, error) {
	hasAssociatedAuthCredential := false
	pubKey, err := crypto.DecompressPubkey(publicKey)
	if err != nil {
		log.Error(ctx, "failed to decompress public key", "err", err)
		return false, err
	}

	keyETHAddress := crypto.PubkeyToAddress(*pubKey)
	isEthAddress, identityAddress, err := common.CheckEthIdentityByDID(did)
	if err != nil {
		log.Error(ctx, "failed to check if DID is ETH identity", "err", err)
		return false, err
	}

	identityAddressToBeChecked := strings.ToUpper("0x" + identityAddress)
	if isEthAddress {
		hasAssociatedAuthCredential = identityAddressToBeChecked == strings.ToUpper(keyETHAddress.Hex())
	}

	return hasAssociatedAuthCredential, nil
}
