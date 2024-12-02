package kms

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"regexp"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/utils"

	"github.com/polygonid/sh-id-platform/internal/log"
)

type localBJJKeyProvider struct {
	keyType          KeyType
	reIdenKeyPathHex *regexp.Regexp // RE of key path bound to identity
	reAnonKeyPathHex *regexp.Regexp // RE of key path not bound to identity
	storageManager   StorageManager
	temporaryKeys    map[string]map[string]string
}

// NewLocalBJJKeyProvider - creates new key provider for BabyJubJub keys stored in local storage
func NewLocalBJJKeyProvider(keyType KeyType, storageManager StorageManager) KeyProvider {
	keyTypeRE := regexp.QuoteMeta(string(keyType))
	reIdenKeyPathHex := regexp.MustCompile("^(?i).*/" + keyTypeRE + ":([a-f0-9]{64})$")
	reAnonKeyPathHex := regexp.MustCompile("^(?i)" + keyTypeRE + ":([a-f0-9]{64})$")
	return &localBJJKeyProvider{
		keyType:          keyType,
		storageManager:   storageManager,
		reIdenKeyPathHex: reIdenKeyPathHex,
		reAnonKeyPathHex: reAnonKeyPathHex,
		temporaryKeys:    make(map[string]map[string]string),
	}
}

// New generates random a KeyID.
func (ls *localBJJKeyProvider) New(identity *w3c.DID) (KeyID, error) {
	bjjPrivateKey := babyjub.NewRandPrivKey()
	keyID := KeyID{
		Type: ls.keyType,
		ID:   keyPathForLocalProvider(identity, ls.keyType, bjjPrivateKey.Public().String()),
	}
	keyMaterial := map[string]string{
		jsonKeyType: string(ls.keyType),
		jsonKeyData: hex.EncodeToString(bjjPrivateKey[:]),
	}
	if identity == nil {
		ls.temporaryKeys[keyID.ID] = keyMaterial
	} else {
		if err := ls.storageManager.SaveKeyMaterial(context.Background(), keyMaterial, keyID.ID); err != nil {
			return KeyID{}, err
		}
		delete(ls.temporaryKeys, keyID.ID)
	}
	return keyID, nil
}

// PublicKey returns bytes representation for public key for specified key ID
func (ls *localBJJKeyProvider) PublicKey(keyID KeyID) ([]byte, error) {
	ctx := context.Background()
	if keyID.Type != ls.keyType {
		return nil, ErrIncorrectKeyType
	}

	_, err := ls.storageManager.getKeyMaterial(ctx, keyID)
	if err != nil {
		log.Error(ctx, "cannot get public key", "err", err, "keyID", keyID)
		return nil, err
	}

	ss := ls.reAnonKeyPathHex.FindStringSubmatch(keyID.ID)
	if ss == nil {
		ss = ls.reIdenKeyPathHex.FindStringSubmatch(keyID.ID)
	}
	if len(ss) != partsNumber {
		return nil, errors.New("unable to get public key from key ID")
	}

	val, err := hex.DecodeString(ss[1])
	return val, err
}

// Sign signs digest with private key
func (ls *localBJJKeyProvider) Sign(ctx context.Context, keyID KeyID, data []byte) ([]byte, error) {
	if len(data) > defaultLength {
		return nil, errors.New("data to sign is too large")
	}

	i := new(big.Int).SetBytes(utils.SwapEndianness(data))
	if !utils.CheckBigIntInField(i) {
		return nil, errors.New("data to sign is too large")
	}

	privKeyData, err := ls.privateKey(ctx, keyID)
	if err != nil {
		return nil, err
	}

	privKey, err := decodeBJJPrivateKey(privKeyData)
	if err != nil {
		return nil, err
	}

	sig := privKey.SignPoseidon(i).Compress()
	return sig[:], nil
}

// ListByIdentity lists keys by identity
func (ls *localBJJKeyProvider) ListByIdentity(ctx context.Context, identity w3c.DID) ([]KeyID, error) {
	return ls.storageManager.searchByIdentity(ctx, identity, ls.keyType)
}

// LinkToIdentity links key to identity
func (ls *localBJJKeyProvider) LinkToIdentity(ctx context.Context, keyID KeyID, identity w3c.DID) (KeyID, error) {
	if keyID.Type != ls.keyType {
		return keyID, ErrIncorrectKeyType
	}

	keyMaterial, ok := ls.temporaryKeys[keyID.ID]
	if !ok {
		return keyID, errors.New("key not found")
	}
	delete(ls.temporaryKeys, keyID.ID)
	newKey := getKeyID(&identity, ls.keyType, keyID.ID)
	if err := ls.storageManager.SaveKeyMaterial(ctx, keyMaterial, newKey); err != nil {
		return KeyID{}, err
	}
	keyID.ID = identity.String()
	return keyID, nil
}

func (ls *localBJJKeyProvider) Delete(ctx context.Context, keyID KeyID) error {
	return ls.storageManager.deleteKeyMaterial(ctx, keyID)
}

func (ls *localBJJKeyProvider) privateKey(ctx context.Context, keyID KeyID) ([]byte, error) {
	if keyID.Type != ls.keyType {
		return nil, ErrIncorrectKeyType
	}

	if !ls.reAnonKeyPathHex.MatchString(keyID.ID) &&
		!ls.reIdenKeyPathHex.MatchString(keyID.ID) {
		log.Error(ctx, "incorrect key ID", "keyID", keyID)
		return nil, errors.New("incorrect key ID")
	}

	privateKey, err := ls.storageManager.searchPrivateKey(ctx, keyID)
	if err != nil {
		log.Error(ctx, "cannot get private key", "err", err, "keyID", keyID)
		return nil, err
	}

	val, err := hex.DecodeString(privateKey)
	if err != nil {
		log.Error(ctx, "cannot decode private key", "err", err, "keyID", keyID)
		return nil, err
	}

	if len(val) != defaultLength {
		log.Error(ctx, "incorrect private key", "keyID", keyID)
		return nil, errors.New("incorrect private key")
	}

	return val, nil
}
