package kms

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"regexp"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/vault/api"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/pkg/errors"

	"github.com/polygonid/sh-id-platform/internal/log"
)

type vaultETHKeyProvider struct {
	keyType          KeyType
	vaultCli         *api.Client
	reIdenKeyPathHex *regexp.Regexp
}

func (v *vaultETHKeyProvider) LinkToIdentity(ctx context.Context, keyID KeyID, identity w3c.DID) (KeyID, error) {
	return keyID, errors.New("Ethereum keys does not support binding")
}

// nolint
func (v *vaultETHKeyProvider) privateKey(keyID KeyID) ([]byte, error) {
	if keyID.Type != v.keyType {
		return nil, errors.WithStack(ErrIncorrectKeyType)
	}

	if keyID.ID == "" {
		return nil, errors.New("key ID is empty")
	}

	path := absVaultSecretPath(keyID.ID)
	secret, err := v.vaultCli.Logical().Read(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	secData, err := getKVv2SecretData(secret)
	if err != nil {
		return nil, err
	}

	var keyHexI interface{}

	ss := v.reIdenKeyPathHex.FindStringSubmatch(keyID.ID)
	if len(ss) == 2 {
		// key stored for identity in format
		//   key_type: type
		//   key_data: private_key
		keyTypeI, ok := secData[jsonKeyType]
		if !ok {
			return nil, errors.New("key type not found")
		}
		keyType, ok := keyTypeI.(string)
		if !ok {
			return nil, errors.New("unexpected format of key type")
		}
		if KeyType(keyType) != v.keyType {
			return nil, errors.WithStack(ErrIncorrectKeyType)
		}
		keyHexI, ok = secData[jsonKeyData]
		if !ok {
			return nil, errors.New("key data not found")
		}
	} else {
		// key is stored in root directory
		var ok bool
		keyHexI, ok = secData[keyID.ID]
		if !ok {
			return nil, errors.New("private key not found")
		}
	}

	keyHex, ok := keyHexI.(string)
	if !ok {
		return nil, errors.New("unexpected format for private key")
	}
	val, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if len(val) != 32 {
		return nil, errors.New("incorrect private key")
	}

	return val, nil
}

func (v *vaultETHKeyProvider) Sign(_ context.Context, keyID KeyID, data []byte) ([]byte, error) {
	privKeyData, err := v.privateKey(keyID)
	if err != nil {
		return nil, err
	}

	privKey, err := decodeETHPrivateKey(privKeyData)
	if err != nil {
		return nil, err
	}

	sig, err := crypto.Sign(data, privKey)
	return sig, errors.WithStack(err)
}

func (v *vaultETHKeyProvider) ListByIdentity(_ context.Context, identity w3c.DID) ([]KeyID, error) {
	path := identityPath(&identity)
	entries, err := listDirectoryEntries(v.vaultCli, path)
	if err != nil {
		return nil, err
	}

	reVaultKeyHex, err := regexp.Compile("^(?i)" +
		regexp.QuoteMeta(string(v.keyType)) + ":([a-f0-9]{66})$")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var result []KeyID //nolint:prealloc // result may be empty
	for _, k := range entries {
		if !reVaultKeyHex.MatchString(k) {
			// ignore unknown keys
			continue
		}

		result = append(result, KeyID{
			Type: v.keyType,
			ID:   path + "/" + k,
		})
	}
	return result, nil
}

// nolint
func (v *vaultETHKeyProvider) PublicKey(keyID KeyID) ([]byte, error) {
	if keyID.Type != v.keyType {
		return nil, errors.New("incorrect key type")
	}

	ss := v.reIdenKeyPathHex.FindStringSubmatch(keyID.ID)
	if len(ss) != 2 {
		// if not found. try get public key from private key.
		pkBytes, err := v.privateKey(keyID)
		if err != nil {
			return nil, errors.New("unable to get private key for build public key")
		}
		pk, err := decodeETHPrivateKey(pkBytes)
		if err != nil {
			return nil, err
		}
		switch v := pk.Public().(type) {
		case *ecdsa.PublicKey:
			return crypto.FromECDSAPub(v), nil
		default:
			return nil, errors.New("unable to get public key from key ID")
		}
	}

	val, err := hex.DecodeString(ss[1])
	return val, errors.WithStack(err)
}

func (v *vaultETHKeyProvider) New(identity *w3c.DID) (KeyID, error) {
	keyID := KeyID{Type: v.keyType}

	if identity == nil {
		return keyID, errors.New(
			"Ethereum keys can be created only for non-nil identities")
	}

	ethPrivKey, err := crypto.GenerateKey()
	if err != nil {
		return keyID, errors.WithStack(err)
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
	keyID.ID = keyPath(identity, v.keyType, pubKeyHex)

	return keyID, saveKeyMaterial(v.vaultCli, keyID.ID, keyMaterial)
}

// NewVaultEthProvider creates new provider for Ethereum keys stored in vault
func NewVaultEthProvider(valutCli *api.Client, keyType KeyType) KeyProvider {
	reIdenKeyPathHex := regexp.MustCompile("^(?i).*/" +
		regexp.QuoteMeta(string(keyType)) + ":([a-f0-9]{66})$")
	return &vaultETHKeyProvider{keyType, valutCli, reIdenKeyPathHex}
}

// DecodeETHPubKey is a helper method to convert byte representation of public
// key to *ecdsa.PublicKey
func DecodeETHPubKey(key []byte) (*ecdsa.PublicKey, error) {
	pubKey, err := crypto.DecompressPubkey(key)
	return pubKey, errors.WithStack(err)
}

// EthPubKey returns the ethereum public key from the key manager service.
// the public key is either uncompressed or compressed, so we need to handle both cases.
// TODO: Move out of this package
func EthPubKey(ctx context.Context, keyMS KMSType, keyID KeyID) (*ecdsa.PublicKey, error) {
	const (
		uncompressedKeyLength = 65
		awsKeyLength          = 88
		defaultKeyLength      = 33
	)

	if keyID.Type != KeyTypeEthereum {
		return nil, errors.New("key type is not ethereum")
	}

	keyBytes, err := keyMS.PublicKey(keyID)
	if err != nil {
		log.Error(ctx, "can't get bytes from public key", "err", err)
		return nil, err
	}

	// public key is uncompressed. It's 65 bytes long.
	if len(keyBytes) == uncompressedKeyLength {
		return crypto.UnmarshalPubkey(keyBytes)
	}

	// public key is AWS format. It's 88 bytes long.
	if len(keyBytes) == awsKeyLength {
		return DecodeAWSETHPubKey(ctx, keyBytes)
	}

	// public key is compressed. It's 33 bytes long.
	if len(keyBytes) == defaultKeyLength {
		return DecodeETHPubKey(keyBytes)
	}

	return nil, errors.New("unsupported public key format")
}
