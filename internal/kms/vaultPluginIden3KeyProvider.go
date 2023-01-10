package kms

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/vault/api"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-iden3-crypto/utils"
)

const (
	keyDest      = "dest"
	keyData      = "data"
	keySignature = "signature"
	keyPublicKey = "public_key"
)

type pluginIden3KeyTp string

const (
	pluginIden3KeyTpUndefined pluginIden3KeyTp = ""
	pluginIden3KeyTpBJJ       pluginIden3KeyTp = "babyjubjub"
	pluginIden3KeyTpETH       pluginIden3KeyTp = "ethereum"
)

type vaultPluginIden3KeyProvider struct {
	keyType        KeyType
	vaultCli       *api.Client
	keysMountPath  string
	keysPathPrefix string
	keyNameRE      *regexp.Regexp
}

func (v *vaultPluginIden3KeyProvider) keyPathFromID(keyID KeyID) keyPathT {
	return keyPathT{
		keyID:     keyID.ID,
		mountPath: v.keysMountPath,
	}
}

func (v *vaultPluginIden3KeyProvider) LinkToIdentity(_ context.Context, keyID KeyID, identity core.DID) (KeyID, error) {
	if keyID.Type != v.keyType {
		return keyID, ErrIncorrectKeyType
	}

	keyPath := v.keyPathFromID(keyID)
	pubKeyStr, err := publicKey(v.vaultCli, keyPath)
	if err != nil {
		return KeyID{}, err
	}

	newKeyPath := v.keyPathFromPublic(&identity, pubKeyStr)
	if newKeyPath == keyPath {
		return keyID, nil
	}

	err = moveKey(v.vaultCli, keyPath, newKeyPath)
	if err != nil {
		return KeyID{}, err
	}

	keyID.ID = newKeyPath.keyID
	return keyID, nil
}

// Sign signs *big.Int using poseidon algorithm.
// data should be a little-endian bytes representation of *big.Int.
func (v *vaultPluginIden3KeyProvider) Sign(_ context.Context, keyID KeyID, dataToSign []byte) ([]byte, error) {
	switch keyID.Type {
	case KeyTypeBabyJubJub:
		if len(dataToSign) > 32 {
			return nil, errors.New("data to sign is too large")
		}

		i := new(big.Int).SetBytes(utils.SwapEndianness(dataToSign))
		if !utils.CheckBigIntInField(i) {
			return nil, errors.New("data to sign is too large")
		}
	case KeyTypeEthereum:
		if len(dataToSign) != common.HashLength {
			return nil, fmt.Errorf("data to sign should be %v bytes length",
				common.HashLength)
		}
	default:
		return nil, errors.New("unsupported key type")
	}

	return signData(v.vaultCli, v.keyPathFromID(keyID), dataToSign)
}

func (v *vaultPluginIden3KeyProvider) ListByIdentity(_ context.Context, identity core.DID) ([]KeyID, error) {
	identityKeysPath := keyPathT{
		keyID:     v.identityPath(identity),
		mountPath: v.keysMountPath,
	}
	entries, err := listKVv1Entries(v.vaultCli, identityKeysPath.keys())
	if err != nil {
		return nil, err
	}

	result := make([]KeyID, len(entries))
	ln := 0
	for i, k := range entries {
		if !v.keyNameRE.MatchString(k) {
			// ignore unknown keys
			continue
		}

		result[i].Type = v.keyType
		result[i].ID = path.Join(identityKeysPath.keyID, k)
		ln++
	}

	return result[:ln], nil
}

func (v *vaultPluginIden3KeyProvider) PublicKey(keyID KeyID) ([]byte, error) {
	if keyID.Type != v.keyType {
		return nil, errors.New("incorrect key type")
	}

	publicKeyStr, err := publicKey(v.vaultCli, v.keyPathFromID(keyID))
	if err != nil {
		return nil, err
	}

	val, err := hex.DecodeString(publicKeyStr)
	return val, err
}

func (v *vaultPluginIden3KeyProvider) New(identity *core.DID) (KeyID, error) {
	randomKeyPath, err := v.randomKeyPath()
	if err != nil {
		return KeyID{}, err
	}

	err = newRandomKey(v.vaultCli, randomKeyPath, v.keyType)
	if err != nil {
		return KeyID{}, err
	}

	pubKeyStr, err := publicKey(v.vaultCli, randomKeyPath)
	if err != nil {
		return KeyID{}, err
	}

	keyPath := v.keyPathFromPublic(identity, pubKeyStr)
	keyID := KeyID{Type: v.keyType, ID: keyPath.keyID}
	err = moveKey(v.vaultCli, randomKeyPath, keyPath)
	if err != nil {
		return KeyID{}, err
	}

	return keyID, nil
}

func (v *vaultPluginIden3KeyProvider) randomKeyPath() (keyPathT, error) {
	var rnd [16]byte
	_, err := rand.Read(rnd[:])
	if err != nil {
		return keyPathT{}, err
	}

	keyName := v.keyFileName(hex.EncodeToString(rnd[:]))
	return keyPathT{
		keyID:     path.Join(v.keysPathPrefix, "random", keyName),
		mountPath: v.keysMountPath,
	}, nil
}

func (v *vaultPluginIden3KeyProvider) keyPathFromPublic(identity *core.DID, publicKey string) keyPathT {
	basePath := v.keysPathPrefix
	if identity != nil {
		basePath = v.identityPath(*identity)
	}
	return keyPathT{
		keyID:     path.Join(basePath, v.keyFileName(publicKey)),
		mountPath: v.keysMountPath,
	}
}

func (v *vaultPluginIden3KeyProvider) identityPath(identity core.DID) string {
	return path.Join(v.keysPathPrefix, identity.String())
}

func (v *vaultPluginIden3KeyProvider) keyFileName(keyName string) string {
	return string(v.keyType) + ":" + keyName
}

// NewVaultPluginIden3KeyProvider creates new key provider for BabyJubJub and
// Ethereum keys stored in vault
func NewVaultPluginIden3KeyProvider(vaultCli *api.Client, keysPath string, keyType KeyType) (KeyProvider, error) {
	var pubKeyLn uint64
	switch keyType {
	case KeyTypeBabyJubJub:
		pubKeyLn = 32
	case KeyTypeEthereum:
		pubKeyLn = 33
	default:
		return nil, errors.New("unsupported key type")
	}

	keysPath = strings.Trim(keysPath, "/")
	if keysPath == "" {
		return nil, errors.New("keys path cannot be empty")
	}
	var keysPathPrefix string
	parts := strings.SplitN(keysPath, "/", 2)
	if len(parts) > 1 {
		keysPathPrefix = parts[1]
	}

	keyNameRE, err := regexp.Compile(fmt.Sprintf("^(?i)%v:([a-f0-9]{%v})$",
		regexp.QuoteMeta(string(keyType)),
		regexp.QuoteMeta(strconv.FormatUint(pubKeyLn*2, 10))))
	if err != nil {
		return nil, err
	}

	return &vaultPluginIden3KeyProvider{
			keyType:        keyType,
			vaultCli:       vaultCli,
			keysMountPath:  parts[0],
			keysPathPrefix: keysPathPrefix,
			keyNameRE:      keyNameRE},
		nil
}

// create random key in vault
func newRandomKey(vaultCli *api.Client, keyPath keyPathT, keyType KeyType) error {
	pluginKeyType, err := toPluginKeyType(keyType)
	if err != nil {
		return err
	}

	_, err = vaultCli.Logical().Write(keyPath.new(),
		map[string]interface{}{jsonKeyType: pluginKeyType})
	return err
}

// move the key under new path
func moveKey(vaultCli *api.Client, oldPath, newPath keyPathT) error {
	data := map[string]interface{}{keyDest: newPath.keys()}
	_, err := vaultCli.Logical().Write(oldPath.move(), data)
	return err
}

// get string representation of public key
func publicKey(vaultCli *api.Client, keyPath keyPathT) (string, error) {
	secret, err := vaultCli.Logical().Read(keyPath.keys())
	if err != nil {
		return "", err
	}

	data, err := getSecretData(secret)
	if err != nil {
		return "", err
	}

	pubKeyStr, ok := data[keyPublicKey].(string)
	if !ok {
		return "", errors.New("unable to get public key from secret")
	}

	return pubKeyStr, nil
}

// return non-nil secret.Data or return error if something in chain is nil.
func getSecretData(secret *api.Secret) (map[string]interface{}, error) {
	if secret == nil {
		return nil, errors.New("secret is nil")
	}

	if secret.Data == nil {
		return nil, errors.New("secret data is nil")
	}

	return secret.Data, nil
}

func signData(vaultCli *api.Client, keyPath keyPathT, dataToSign []byte) ([]byte, error) {
	dataStr := hex.EncodeToString(dataToSign)
	data := map[string][]string{keyData: {dataStr}}
	secret, err := vaultCli.Logical().ReadWithData(keyPath.sign(), data)
	if err != nil {
		return nil, err
	}
	data2, err := getSecretData(secret)
	if err != nil {
		return nil, err
	}
	sigStr, ok := data2[keySignature].(string)
	if !ok {
		return nil, errors.New("unable to get signature from secret")
	}
	sig, err := hex.DecodeString(sigStr)
	return sig, err
}

func listKVv1Entries(vaultCli *api.Client, dirPath string) ([]string, error) {
	se, err := vaultCli.Logical().List(dirPath)
	if err != nil {
		return nil, err
	}
	if se == nil {
		return nil, nil
	}
	keys, ok := se.Data["keys"]
	if !ok {
		return nil, errors.New("keys section is empty")
	}
	keysList, ok := keys.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected keys section format: %T", keys)
	}

	result := make([]string, 0, len(keysList))
	for _, k := range keysList {
		str, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected key type: %T", k)
		}
		result = append(result, str)
	}
	return result, nil
}

type keyPathT struct {
	keyID     string
	mountPath string
}

func (p keyPathT) join(verb string) string {
	return path.Join(p.mountPath, verb, p.keyID)
}

func (p keyPathT) keys() string {
	return p.join("keys")
}

func (p keyPathT) move() string {
	return p.join("move")
}

func (p keyPathT) sign() string {
	return p.join("sign")
}

func (p keyPathT) new() string {
	return p.join("new")
}

func toPluginKeyType(keyType KeyType) (pluginIden3KeyTp, error) {
	switch keyType {
	case KeyTypeBabyJubJub:
		return pluginIden3KeyTpBJJ, nil
	case KeyTypeEthereum:
		return pluginIden3KeyTpETH, nil
	default:
		return pluginIden3KeyTpUndefined, errors.New("unsupported key type")
	}
}
