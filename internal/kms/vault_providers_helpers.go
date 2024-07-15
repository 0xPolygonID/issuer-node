package kms

import (
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/vault/api"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/pkg/errors"
)

const keysPathPrefix = "keys/"

const kvStoragePath = "secret"

const (
	jsonKeyType   = "key_type"
	jsonKeyData   = "key_data"
	defaultLength = 32
	partsNumber   = 2
	// LocalStorageFileName is the name of the file where the keys are stored
	LocalStorageFileName = "kms_localstorage_keys.json"
)

func saveKeyMaterial(vaultCli *api.Client, path string, jsonObj map[string]string) error {
	secret := map[string]interface{}{"data": jsonObj}
	vaultPath := absVaultSecretPath(path)
	_, err := vaultCli.Logical().Write(vaultPath, secret)
	return errors.WithStack(err)
}

func listDirectoryEntries(vaultCli *api.Client, path string) ([]string, error) {
	path = strings.TrimPrefix(path, "/")
	path = kvStoragePath + "/metadata/" + path
	se, err := vaultCli.Logical().List(path)
	if err != nil {
		return nil, errors.WithStack(err)
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
		return nil, errors.Errorf("unexpected keys section format: %T", keys)
	}
	result := make([]string, 0, len(keysList))
	for _, k := range keysList {
		str, ok := k.(string)
		if !ok {
			return nil, errors.Errorf("unexpected key type: %T", k)
		}
		result = append(result, str)
	}
	return result, nil
}

func identityPath(identity *w3c.DID) string {
	return fmt.Sprintf("%v%v", keysPathPrefix, identity.String())
}

func keyPath(identity *w3c.DID, keyType KeyType, keyID string) string {
	basePath := ""
	if identity != nil {
		basePath = identityPath(identity) + "/"
	}
	return basePath + string(keyType) + ":" + keyID
}

func absVaultSecretPath(path string) string {
	return kvStoragePath + "/data/" + strings.TrimPrefix(path, "/")
}

// extract data map from Secret for kv v2 storage (secret.Data["data"])
func getKVv2SecretData(secret *api.Secret) (map[string]interface{}, error) {
	if secret == nil {
		return nil, errors.New("secret is nil")
	}

	if secret.Data == nil {
		return nil, errors.New("secret data is nil")
	}

	secDataI, ok := secret.Data["data"]
	if !ok {
		return nil, errors.New("secret data not found")
	}

	secData, ok := secDataI.(map[string]interface{})
	if !ok {
		return nil, errors.New("secret data has unexpected format")
	}

	return secData, nil
}

func moveSecretData(vaultCli *api.Client, oldPath, newPath string) error {
	cli := vaultCli.Logical()
	secret, err := cli.Read(absVaultSecretPath(oldPath))
	if err != nil {
		return errors.WithStack(err)
	}

	secData, err := getKVv2SecretData(secret)
	if err != nil {
		return err
	}

	_, err = cli.Write(absVaultSecretPath(newPath),
		map[string]interface{}{"data": secData})
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = cli.Delete(absVaultSecretPath(oldPath))
	return errors.WithStack(err)
}

// decodeETHPrivateKey is a helper method to convert byte representation of
// private key to *ecdsa.PrivateKey
func decodeETHPrivateKey(key []byte) (*ecdsa.PrivateKey, error) {
	privKey, err := crypto.ToECDSA(key)
	return privKey, errors.WithStack(err)
}
