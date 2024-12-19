package services

import (
	"context"
	b64 "encoding/base64"
	"errors"
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
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

var (

	// ErrAuthCredentialNotRevoked is returned when the associated auth core claim is not revoked
	ErrAuthCredentialNotRevoked = errors.New("associated auth core claim not revoked")
	// ErrKeyAssociatedWithIdentity is returned when the key is associated with an identity
	ErrKeyAssociatedWithIdentity = errors.New("key is associated with an identity")
	// ErrDuplicateKeyName is returned when the key name already exists
	ErrDuplicateKeyName = errors.New("duplicate key name")
)

// Key is the service that manages keys
type Key struct {
	kms           *kms.KMS
	claimService  ports.ClaimService
	keyRepository ports.KeyRepository
}

// NewKey creates a new Key
func NewKey(kms *kms.KMS, claimService ports.ClaimService, keyRepository ports.KeyRepository) ports.KeyService {
	return &Key{
		kms:           kms,
		claimService:  claimService,
		keyRepository: keyRepository,
	}
}

// Create creates a new key for the given DID
func (ks *Key) Create(ctx context.Context, did *w3c.DID, keyType kms.KeyType, name string) (kms.KeyID, error) {
	var keyID kms.KeyID
	var err error

	keyWithName, err := ks.keyRepository.GetByName(ctx, *did, name)
	if keyWithName != nil {
		return kms.KeyID{}, ErrDuplicateKeyName
	}
	if err != nil {
		if !errors.Is(err, repositories.ErrKeyNotFound) {
			return kms.KeyID{}, err
		}
	}

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

	publicKeyAsBytes, err := ks.kms.PublicKey(keyID)
	if err != nil {
		log.Error(ctx, "failed to get public key", "err", err)
		return kms.KeyID{}, err
	}

	publicKey := hexutil.Encode(publicKeyAsBytes)
	keyToSave := domain.NewKey(*did, publicKey, name)
	_, err = ks.keyRepository.Save(ctx, nil, keyToSave)
	if err != nil {
		log.Error(ctx, "failed to save key", "err", err)
		return kms.KeyID{}, err
	}

	encodedKeyID := b64.StdEncoding.EncodeToString([]byte(keyID.ID))
	log.Info(ctx, "key created successfully", "keyID", encodedKeyID)
	keyID.ID = encodedKeyID
	return keyID, nil
}

// Update updates the key with the given keyID
func (ks *Key) Update(ctx context.Context, did *w3c.DID, keyID string, name string) error {
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
		return ErrKeyNotFound
	}

	publicKey, err := ks.getPublicKey(ctx, keyID)
	if err != nil {
		log.Error(ctx, "failed to get public key", "err", err)
		return ErrKeyNotFound
	}

	keyInfo, err := ks.keyRepository.GetByPublicKey(ctx, *did, hexutil.Encode(publicKey))
	if err != nil {
		if !errors.Is(err, repositories.ErrKeyNotFound) {
			return err
		}
	}
	if keyInfo == nil {
		keyInfo = domain.NewKey(*did, hexutil.Encode(publicKey), name)
	}
	keyInfo.Name = name
	_, err = ks.keyRepository.Save(ctx, nil, keyInfo)
	return err
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
		return nil, ErrKeyNotFound
	}

	publicKey, err := ks.getPublicKey(ctx, keyID)
	if err != nil {
		log.Error(ctx, "failed to get public key", "err", err)
		return nil, ErrKeyNotFound
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
		return nil, ErrInvalidKeyType
	}

	keyInfo, err := ks.keyRepository.GetByPublicKey(ctx, *did, hexutil.Encode(publicKey))
	if err != nil {
		if !errors.Is(err, repositories.ErrKeyNotFound) {
			return nil, err
		}
		keyInfo = &domain.Key{
			Name: hexutil.Encode(publicKey),
		}
	}

	return &ports.KMSKey{
		KeyID:                       keyID,
		KeyType:                     keyType,
		PublicKey:                   hexutil.Encode(publicKey),
		HasAssociatedAuthCredential: hasAssociatedAuthCredential,
		Name:                        keyInfo.Name,
	}, nil
}

// GetAll returns all the keys for the given DID
func (ks *Key) GetAll(ctx context.Context, did *w3c.DID, filter ports.KeyFilter) ([]*ports.KMSKey, uint, error) {
	keyIDs, err := ks.kms.KeysByIdentity(ctx, *did)
	if err != nil {
		log.Error(ctx, "failed to get keys", "err", err)
		return nil, 0, err
	}

	if filter.KeyType != nil {
		filteredKeyIDs := make([]kms.KeyID, 0)
		for _, keyID := range keyIDs {
			if keyID.Type == *filter.KeyType {
				filteredKeyIDs = append(filteredKeyIDs, keyID)
			}
		}
		keyIDs = filteredKeyIDs
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

	keys := make([]*ports.KMSKey, len(keyIDs))
	for i, keyID := range keyIDs {
		key, err := ks.Get(ctx, did, keyID.ID)
		if err != nil {
			log.Error(ctx, "failed to get key", "err", err)
			return nil, 0, err
		}
		keys[i] = key
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Name < keys[j].Name
	})

	keys = keys[start:end]
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
		return ErrKeyNotFound
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
				return ErrAuthCredentialNotRevoked
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
			return ErrKeyAssociatedWithIdentity
		}
	default:
		return ErrInvalidKeyType
	}

	if err := ks.keyRepository.Delete(ctx, *did, hexutil.Encode(publicKey)); err != nil {
		log.Error(ctx, "failed to delete key", "err", err)
		return err
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
		return keyType, ErrInvalidKeyType
	}

	return keyType, nil
}

// hasAssociatedAuthCredential checks if the bbj key has an associated auth credential
func (ks *Key) hasAssociatedAuthCredential(ctx context.Context, did *w3c.DID, publicKey []byte) (bool, *domain.Claim, error) {
	hasAssociatedAuthCredential := false
	authCredential, err := ks.claimService.GetAuthCredentialByPublicKey(ctx, did, publicKey)
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
