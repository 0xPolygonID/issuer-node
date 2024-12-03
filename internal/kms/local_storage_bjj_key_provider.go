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

type localStorageBJJKeyProviderFileContent struct {
	KeyType    string `json:"key_type"`
	KeyPath    string `json:"key_path"`
	PrivateKey string `json:"private_key"`
}

type localStorageBJJKeyProvider struct {
	keyType                 KeyType
	reIdenKeyPathHex        *regexp.Regexp // RE of key path bounded to identity
	reAnonKeyPathHex        *regexp.Regexp // RE of key path not bounded to identity
	localStorageFileManager LocalStorageFileManager
}

// NewLocalStorageBJJKeyProvider - creates new key provider for BabyJubJub keys stored in local storage
func NewLocalStorageBJJKeyProvider(keyType KeyType, localStorageFileManager LocalStorageFileManager) KeyProvider {
	keyTypeRE := regexp.QuoteMeta(string(keyType))
	reIdenKeyPathHex := regexp.MustCompile("^(?i).*/" + keyTypeRE + ":([a-f0-9]{64})$")
	reAnonKeyPathHex := regexp.MustCompile("^(?i)" + keyTypeRE + ":([a-f0-9]{64})$")
	return &localStorageBJJKeyProvider{
		keyType:                 keyType,
		localStorageFileManager: localStorageFileManager,
		reIdenKeyPathHex:        reIdenKeyPathHex,
		reAnonKeyPathHex:        reAnonKeyPathHex,
	}
}

// New generates random a KeyID.
func (ls *localStorageBJJKeyProvider) New(identity *w3c.DID) (KeyID, error) {
	ctx := context.Background()
	bjjPrivateKey := babyjub.NewRandPrivKey()
	keyID := KeyID{
		Type: ls.keyType,
		ID:   keyPath(identity, ls.keyType, bjjPrivateKey.Public().String()),
	}
	keyMaterial := map[string]string{
		jsonKeyType: string(keyID.Type),
		jsonKeyData: hex.EncodeToString(bjjPrivateKey[:]),
	}
	if err := ls.localStorageFileManager.saveKeyMaterialToFile(ctx, keyMaterial, keyID.ID); err != nil {
		return KeyID{}, err
	}
	return keyID, nil
}

// PublicKey returns bytes representation for public key for specified key ID
func (ls *localStorageBJJKeyProvider) PublicKey(keyID KeyID) ([]byte, error) {
	if keyID.Type != ls.keyType {
		return nil, ErrIncorrectKeyType
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
func (ls *localStorageBJJKeyProvider) Sign(ctx context.Context, keyID KeyID, data []byte) ([]byte, error) {
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
func (ls *localStorageBJJKeyProvider) ListByIdentity(ctx context.Context, identity w3c.DID) ([]KeyID, error) {
	return ls.localStorageFileManager.searchByIdentityInFile(ctx, identity, ls.keyType)
}

// LinkToIdentity links key to identity
func (ls *localStorageBJJKeyProvider) LinkToIdentity(ctx context.Context, keyID KeyID, identity w3c.DID) (KeyID, error) {
	if keyID.Type != ls.keyType {
		return keyID, ErrIncorrectKeyType
	}

	err := ls.localStorageFileManager.searchKeyMaterialInFileAndReplace(ctx, keyID.ID, identity)
	if err != nil {
		return keyID, err
	}

	keyID.ID = identity.String()
	return keyID, nil
}

func (ls *localStorageBJJKeyProvider) privateKey(ctx context.Context, keyID KeyID) ([]byte, error) {
	if keyID.Type != ls.keyType {
		return nil, ErrIncorrectKeyType
	}

	if !ls.reAnonKeyPathHex.MatchString(keyID.ID) &&
		!ls.reIdenKeyPathHex.MatchString(keyID.ID) {
		log.Error(ctx, "incorrect key ID", "keyID", keyID)
		return nil, errors.New("incorrect key ID")
	}

	privateKey, err := ls.localStorageFileManager.searchPrivateKeyInFile(ctx, keyID)
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
