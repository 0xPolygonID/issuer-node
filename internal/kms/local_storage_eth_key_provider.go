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

type localStorageEthKeyProvider struct {
	keyType                 KeyType
	reIdenKeyPathHex        *regexp.Regexp // RE of key path bounded to identity
	localStorageFileManager LocalStorageFileManager
}

// NewLocalStorageEthKeyProvider - creates new key provider for Ethereum keys stored in local storage
func NewLocalStorageEthKeyProvider(keyType KeyType, localStorageFileManager LocalStorageFileManager) KeyProvider {
	keyTypeRE := regexp.QuoteMeta(string(keyType))
	reIdenKeyPathHex := regexp.MustCompile("^(?i).*/" + keyTypeRE + ":([a-f0-9]{64})$")
	return &localStorageEthKeyProvider{
		keyType:                 keyType,
		localStorageFileManager: localStorageFileManager,
		reIdenKeyPathHex:        reIdenKeyPathHex,
	}
}

func (ls *localStorageEthKeyProvider) New(identity *w3c.DID) (KeyID, error) {
	ctx := context.Background()
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
	keyID.ID = keyPath(identity, ls.keyType, pubKeyHex)

	if err := ls.localStorageFileManager.saveKeyMaterialToFile(ctx, keyMaterial, keyID.ID); err != nil {
		return KeyID{}, err
	}
	return keyID, nil
}

func (ls *localStorageEthKeyProvider) PublicKey(keyID KeyID) ([]byte, error) {
	ctx := context.Background()
	if keyID.Type != ls.keyType {
		return nil, ErrIncorrectKeyType
	}

	ss := ls.reIdenKeyPathHex.FindStringSubmatch(keyID.ID)
	if len(ss) != partsNumber {
		// if not found. try get public key from private key.
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

func (ls *localStorageEthKeyProvider) Sign(ctx context.Context, keyID KeyID, data []byte) ([]byte, error) {
	privKeyData, err := ls.privateKey(ctx, keyID)
	if err != nil {
		return nil, err
	}

	privKey, err := decodeETHPrivateKey(privKeyData)
	if err != nil {
		return nil, err
	}

	sig, err := crypto.Sign(data, privKey)
	if err != nil {
		return nil, err
	}
	if len(sig) > 65 { // nolint:mnd  // Checking for safe signature length
		sig[64] += 27
	}
	return sig, nil
}

func (ls *localStorageEthKeyProvider) LinkToIdentity(ctx context.Context, keyID KeyID, identity w3c.DID) (KeyID, error) {
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

// ListByIdentity lists keys by identity
func (ls *localStorageEthKeyProvider) ListByIdentity(ctx context.Context, identity w3c.DID) ([]KeyID, error) {
	return ls.localStorageFileManager.searchByIdentityInFile(ctx, identity, ls.keyType)
}

// nolint
func (ls *localStorageEthKeyProvider) privateKey(ctx context.Context, keyID KeyID) ([]byte, error) {
	if keyID.Type != ls.keyType {
		return nil, ErrIncorrectKeyType
	}

	if keyID.ID == "" {
		return nil, errors.New("key ID is empty")
	}

	privateKey, err := ls.localStorageFileManager.searchPrivateKeyInFile(context.Background(), keyID)
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
