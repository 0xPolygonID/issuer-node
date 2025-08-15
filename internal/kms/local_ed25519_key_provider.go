package kms

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"

	"github.com/gagliardetto/solana-go"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/log"
)

type localEd25519KeyProvider struct {
	keyType          KeyType
	reIdenKeyPathHex *regexp.Regexp // RE of key path bounded to identity
	storageManager   StorageManager
	temporaryKeys    map[string]map[string]string
}

// NewLocalEd25519KeyProvider - creates new key provider for Ed25519 keys stored in local storage
func NewLocalEd25519KeyProvider(keyType KeyType, storageManager StorageManager) KeyProvider {
	keyTypeRE := regexp.QuoteMeta(string(keyType))
	reIdenKeyPathHex := regexp.MustCompile("^(?i).*/" + keyTypeRE + ":([a-f0-9]{64})$")
	return &localEd25519KeyProvider{
		keyType:          keyType,
		storageManager:   storageManager,
		reIdenKeyPathHex: reIdenKeyPathHex,
		temporaryKeys:    make(map[string]map[string]string),
	}
}

func (ls *localEd25519KeyProvider) New(identity *w3c.DID) (KeyID, error) {
	keyID := KeyID{Type: ls.keyType}
	ed25519PubKey, ed25519PrivKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return keyID, err
	}

	keyMaterial := map[string]string{
		jsonKeyType: string(KeyTypeEd25519),
		jsonKeyData: hex.EncodeToString(ed25519PrivKey.Seed()),
	}

	pubKey := solana.PublicKeyFromBytes(ed25519PubKey)
	keyID.ID = getKeyID(identity, ls.keyType, pubKey.String())

	ls.temporaryKeys[keyID.ID] = keyMaterial
	return keyID, nil
}

func (ls *localEd25519KeyProvider) PublicKey(keyID KeyID) ([]byte, error) {
	ctx := context.Background()
	if keyID.Type != ls.keyType {
		return nil, ErrIncorrectKeyType
	}

	ss := ls.reIdenKeyPathHex.FindStringSubmatch(keyID.ID)
	if len(ss) != partsNumber {
		pk, err := ls.privateKey(ctx, keyID)
		if err != nil {
			return nil, errors.New("unable to get private key for build public key")
		}
		switch v := pk.Public().(type) {
		case ed25519.PublicKey:
			return v, nil
		default:
			return nil, errors.New("unable to get public key from key ID")
		}
	}

	val, err := hex.DecodeString(ss[1])
	return val, err
}

func (ls *localEd25519KeyProvider) Sign(ctx context.Context, keyID KeyID, data []byte) ([]byte, error) {
	privKey, err := ls.privateKey(ctx, keyID)
	if err != nil {
		return nil, err
	}

	sig := ed25519.Sign(privKey, data)
	return sig, nil
}

func (ls *localEd25519KeyProvider) LinkToIdentity(ctx context.Context, keyID KeyID, identity w3c.DID) (KeyID, error) {
	if keyID.Type != ls.keyType {
		return keyID, ErrIncorrectKeyType
	}

	keyMaterial, ok := ls.temporaryKeys[keyID.ID]
	delete(ls.temporaryKeys, keyID.ID)
	if !ok {
		return keyID, errors.New("key not found")
	}

	newKey := getKeyID(&identity, ls.keyType, keyID.ID)
	if err := ls.storageManager.SaveKeyMaterial(ctx, keyMaterial, newKey); err != nil {
		return KeyID{}, err
	}

	keyID.ID = identity.String() + "/" + keyID.ID
	return keyID, nil
}

// ListByIdentity lists keys by identity
func (ls *localEd25519KeyProvider) ListByIdentity(ctx context.Context, identity w3c.DID) ([]KeyID, error) {
	return ls.storageManager.searchByIdentity(ctx, identity, ls.keyType)
}

func (ls *localEd25519KeyProvider) Delete(ctx context.Context, keyID KeyID) error {
	return ls.storageManager.deleteKeyMaterial(ctx, keyID)
}

func (ls *localEd25519KeyProvider) Exists(ctx context.Context, keyID KeyID) (bool, error) {
	_, err := ls.storageManager.getKeyMaterial(ctx, keyID)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return false, nil
		}
	}
	return true, nil
}

// nolint
func (ls *localEd25519KeyProvider) privateKey(ctx context.Context, keyID KeyID) (ed25519.PrivateKey, error) {
	if keyID.Type != ls.keyType {
		return nil, ErrIncorrectKeyType
	}

	if keyID.ID == "" {
		return nil, errors.New("key ID is empty")
	}

	privateKey := ""
	var err error
	keyMaterial, ok := ls.temporaryKeys[keyID.ID]
	if ok {
		privateKey = keyMaterial["key_data"]
	}

	if privateKey == "" {
		privateKey, err = ls.storageManager.searchPrivateKey(context.Background(), keyID)
		if err != nil {
			log.Error(ctx, "cannot get private key", "err", err, "keyID", keyID)
			return nil, err
		}
	}

	decodedSeed, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}
	val := ed25519.NewKeyFromSeed(decodedSeed)
	return val, nil
}
