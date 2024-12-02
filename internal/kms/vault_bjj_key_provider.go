package kms

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"regexp"

	"github.com/hashicorp/vault/api"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/utils"
)

type vaultBJJKeyProvider struct {
	keyType          KeyType
	vaultCli         *api.Client
	reIdenKeyPathHex *regexp.Regexp // RE of key path bounded to identity
	reAnonKeyPathHex *regexp.Regexp // RE of key path not bounded to identity
}

// NewVaultBJJKeyProvider creates new key provider for BabyJubJub keys stored
// in vault
func NewVaultBJJKeyProvider(vaultCli *api.Client, keyType KeyType) KeyProvider {
	keyTypeRE := regexp.QuoteMeta(string(keyType))
	reIdenKeyPathHex := regexp.MustCompile("^(?i).*/" + keyTypeRE +
		":([a-f0-9]{64})$")
	reAnonKeyPathHex := regexp.MustCompile("^(?i)" + keyTypeRE +
		":([a-f0-9]{64})$")
	return &vaultBJJKeyProvider{keyType, vaultCli, reIdenKeyPathHex, reAnonKeyPathHex}
}

func (v *vaultBJJKeyProvider) New(identity *w3c.DID) (KeyID, error) {
	bjjPrivKey := babyjub.NewRandPrivKey()
	keyID := KeyID{
		Type: v.keyType,
		ID:   keyPath(identity, v.keyType, bjjPrivKey.Public().String()),
	}
	keyMaterial := map[string]string{
		jsonKeyType: string(keyID.Type),
		jsonKeyData: hex.EncodeToString(bjjPrivKey[:]),
	}
	return keyID, saveKeyMaterial(v.vaultCli, keyID.ID, keyMaterial)
}

func (v *vaultBJJKeyProvider) LinkToIdentity(_ context.Context, keyID KeyID, identity w3c.DID) (KeyID, error) {
	if keyID.Type != v.keyType {
		return keyID, ErrIncorrectKeyType
	}

	ss := v.reAnonKeyPathHex.FindStringSubmatch(keyID.ID)
	if len(ss) != partsNumber {
		return keyID, errors.New("key ID does not looks like unbound")
	}

	newKeyID := KeyID{
		Type: keyID.Type,
		ID:   keyPath(&identity, v.keyType, ss[1]),
	}

	return newKeyID, moveSecretData(v.vaultCli, keyID.ID, newKeyID.ID)
}

// Sign signs *big.Int using poseidon algorithm.
// data should be a little-endian bytes representation of *big.Int.
func (v *vaultBJJKeyProvider) Sign(_ context.Context, keyID KeyID, data []byte) ([]byte, error) {
	if len(data) > defaultLength {
		return nil, errors.New("data to sign is too large")
	}

	i := new(big.Int).SetBytes(utils.SwapEndianness(data))
	if !utils.CheckBigIntInField(i) {
		return nil, errors.New("data to sign is too large")
	}

	privKeyData, err := v.privateKey(keyID)
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

func (v *vaultBJJKeyProvider) ListByIdentity(_ context.Context, identity w3c.DID) ([]KeyID, error) {
	path := identityPath(&identity)
	entries, err := listDirectoryEntries(v.vaultCli, path)
	if err != nil {
		return nil, err
	}

	reVaultKeyHex, err := regexp.Compile("^(?i)" +
		regexp.QuoteMeta(string(v.keyType)) + ":([a-f0-9]{64})$")
	if err != nil {
		return nil, err
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

func (v *vaultBJJKeyProvider) PublicKey(keyID KeyID) ([]byte, error) {
	if keyID.Type != v.keyType {
		return nil, errors.New("incorrect key type")
	}

	ss := v.reAnonKeyPathHex.FindStringSubmatch(keyID.ID)
	if ss == nil {
		ss = v.reIdenKeyPathHex.FindStringSubmatch(keyID.ID)
	}
	if len(ss) != partsNumber {
		return nil, errors.New("unable to get public key from key ID")
	}

	val, err := hex.DecodeString(ss[1])
	return val, err
}

func (v *vaultBJJKeyProvider) Delete(ctx context.Context, keyID KeyID) error {
	return nil
}

func (v *vaultBJJKeyProvider) privateKey(keyID KeyID) ([]byte, error) {
	if keyID.Type != v.keyType {
		return nil, ErrIncorrectKeyType
	}

	if !v.reAnonKeyPathHex.MatchString(keyID.ID) &&
		!v.reIdenKeyPathHex.MatchString(keyID.ID) {
		return nil, errors.New("incorrect key ID")
	}

	path := absVaultSecretPath(keyID.ID)
	secret, err := v.vaultCli.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	secData, err := getKVv2SecretData(secret)
	if err != nil {
		return nil, err
	}

	// check key type stored in vault is correct
	keyTypeI, ok := secData[jsonKeyType]
	if !ok {
		return nil, errors.New("key type not found")
	}
	keyType, ok := keyTypeI.(string)
	if !ok {
		return nil, errors.New("unexpected format of key type")
	}
	if KeyType(keyType) != v.keyType {
		return nil, ErrIncorrectKeyType
	}

	keyHexI, ok := secData[jsonKeyData]
	if !ok {
		return nil, errors.New("key data not found")
	}
	keyHex, ok := keyHexI.(string)
	if !ok {
		return nil, errors.New("unexpected format for private key")
	}
	val, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, err
	}
	if len(val) != defaultLength {
		return nil, errors.New("incorrect private key")
	}

	return val, nil
}

// DecodeBJJPubKey is a helper method to convert byte representation of public
// key to *babyjub.PublicKey
func DecodeBJJPubKey(key []byte) (*babyjub.PublicKey, error) {
	var compPubKey babyjub.PublicKeyComp
	copy(compPubKey[:], key)

	pk, err := compPubKey.Decompress()
	return pk, err
}

// decodeBJJPrivateKey is a helper method to convert byte representation of
// private key to babyjub.PrivateKey
func decodeBJJPrivateKey(key []byte) (babyjub.PrivateKey, error) {
	var privKey babyjub.PrivateKey
	n := copy(privKey[:], key)
	if n != len(privKey) {
		return privKey, errors.New("unexpected length of private key")
	}

	return privKey, nil
}

// BJJDigest creates []byte digest for signing from *big.Int by marshaling
// *big.Int to little-endian byte array.
func BJJDigest(i *big.Int) []byte {
	return utils.SwapEndianness(i.Bytes())
}

// DecodeBJJSignature converts byte array representation of BJJ signature
// returned by Sign method to *babyjub.Signature.
func DecodeBJJSignature(sigBytes []byte) (*babyjub.Signature, error) {
	var sigComp babyjub.SignatureComp
	if len(sigBytes) != len(sigComp) {
		return nil, fmt.Errorf(
			"unexpected signature length, got %v bytes, want %v",
			len(sigBytes), len(sigComp))
	}
	copy(sigComp[:], sigBytes)
	sig, err := sigComp.Decompress()
	return sig, err
}
