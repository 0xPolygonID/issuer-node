package kms

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"regexp"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/log"
)

type localEthKeyProvider struct {
	keyType          KeyType
	reIdenKeyPathHex *regexp.Regexp // RE of key path bounded to identity
	storageManager   StorageManager
	temporaryKeys    map[string]map[string]string
}

// NewLocalEthKeyProvider - creates new key provider for Ethereum keys stored in local storage
func NewLocalEthKeyProvider(keyType KeyType, storageManager StorageManager) KeyProvider {
	keyTypeRE := regexp.QuoteMeta(string(keyType))
	reIdenKeyPathHex := regexp.MustCompile("^(?i).*/" + keyTypeRE + ":([a-f0-9]{64})$")
	return &localEthKeyProvider{
		keyType:          keyType,
		storageManager:   storageManager,
		reIdenKeyPathHex: reIdenKeyPathHex,
		temporaryKeys:    make(map[string]map[string]string),
	}
}

func (ls *localEthKeyProvider) New(identity *w3c.DID) (KeyID, error) {
	keyID := KeyID{Type: ls.keyType}
	ethPrivKey, err := crypto.GenerateKey()
	if err != nil {
		return keyID, err
	}

	keyMaterial := map[string]string{
		jsonKeyType: string(KeyTypeEthereum),
		jsonKeyData: hex.EncodeToString(crypto.FromECDSA(ethPrivKey)),
	}

	pubKey, ok := ethPrivKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return keyID, errors.New("unexpected public key type")
	}

	pubKeyBytes := crypto.CompressPubkey(pubKey)
	pubKeyHex := hex.EncodeToString(pubKeyBytes)
	keyID.ID = getKeyID(identity, ls.keyType, pubKeyHex)

	ls.temporaryKeys[keyID.ID] = keyMaterial
	return keyID, nil
}

func (ls *localEthKeyProvider) PublicKey(keyID KeyID) ([]byte, error) {
	ctx := context.Background()
	if keyID.Type != ls.keyType {
		return nil, ErrIncorrectKeyType
	}

	ss := ls.reIdenKeyPathHex.FindStringSubmatch(keyID.ID)
	if len(ss) != partsNumber {
		pkBytes, err := ls.privateKey(ctx, keyID)
		if err != nil {
			return nil, errors.New("unable to get private key for build public key")
		}
		pk, err := decodeETHPrivateKey(pkBytes)
		if err != nil {
			return nil, err
		}
		switch v := pk.Public().(type) {
		case *ecdsa.PublicKey:
			return crypto.CompressPubkey(v), nil
		default:
			return nil, errors.New("unable to get public key from key ID")
		}
	}

	val, err := hex.DecodeString(ss[1])
	return val, err
}

func (ls *localEthKeyProvider) Sign(ctx context.Context, keyID KeyID, data []byte) ([]byte, error) {
	privKeyData, err := ls.privateKey(ctx, keyID)
	if err != nil {
		return nil, err
	}

	privKey, err := decodeETHPrivateKey(privKeyData)
	if err != nil {
		return nil, err
	}

	sig, err := crypto.Sign(data, privKey)
	return sig, err
}

func (ls *localEthKeyProvider) LinkToIdentity(ctx context.Context, keyID KeyID, identity w3c.DID) (KeyID, error) {
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

	keyID.ID = identity.String()
	return keyID, nil
}

// ListByIdentity lists keys by identity
func (ls *localEthKeyProvider) ListByIdentity(ctx context.Context, identity w3c.DID) ([]KeyID, error) {
	return ls.storageManager.searchByIdentity(ctx, identity, ls.keyType)
}

func (ls *localEthKeyProvider) Delete(ctx context.Context, keyID KeyID) error {
	return errors.New("not implemented")
}

func (ls *localEthKeyProvider) Exists(ctx context.Context, keyID KeyID) (bool, error) {
	_, err := ls.storageManager.getKeyMaterial(ctx, keyID)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return false, nil
		}
	}
	return true, nil
}

// nolint
func (ls *localEthKeyProvider) privateKey(ctx context.Context, keyID KeyID) ([]byte, error) {
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
