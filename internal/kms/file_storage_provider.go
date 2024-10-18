package kms

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/log"
)

const (
	partsNumber3 = 3
	babyjubjub   = "babyjubjub"
	ethereum     = "ethereum"
)

type localStorageProviderFileContent struct {
	KeyType    string `json:"key_type"`
	KeyPath    string `json:"key_path"`
	PrivateKey string `json:"private_key"`
}

type localStorageFileProvider struct {
	file string
}

// NewLocalStorageFileProvider - creates new local storage file manager
func NewLocalStorageFileProvider(file string) StorageManager {
	return &localStorageFileProvider{file}
}

func (ls *localStorageFileProvider) SaveKeyMaterial(ctx context.Context, keyMaterial map[string]string, id string) error {
	localStorageFileContent, err := readContentFile(ctx, ls.file)
	if err != nil {
		return err
	}
	localStorageFileContent = append(localStorageFileContent, localStorageProviderFileContent{
		KeyPath:    id,
		KeyType:    keyMaterial[jsonKeyType],
		PrivateKey: keyMaterial[jsonKeyData],
	})

	newFileContent, err := json.Marshal(localStorageFileContent)
	if err != nil {
		log.Error(ctx, "cannot marshal file content", "err", err)
		return err
	}
	// nolint: all
	if err := os.WriteFile(ls.file, newFileContent, 0644); err != nil {
		log.Error(ctx, "cannot write file", "err", err)
		return err
	}
	return nil
}

func (ls *localStorageFileProvider) searchByIdentity(ctx context.Context, identity w3c.DID, keyType KeyType) ([]KeyID, error) {
	localStorageFileContent, err := readContentFile(ctx, ls.file)
	if err != nil {
		return nil, err
	}
	keyIDs := make([]KeyID, 0)
	keyTypeToRead := string(keyType)
	for _, keyMaterial := range localStorageFileContent {
		keyParts := strings.Split(keyMaterial.KeyPath, "/")
		if len(keyParts) != partsNumber && len(keyParts) != partsNumber3 {
			continue
		}
		if keyParts[0] == identity.String() || keyParts[1] == identity.String() {
			if keyMaterial.KeyType == keyTypeToRead {
				keyIDs = append(keyIDs, KeyID{
					Type: KeyType(keyMaterial.KeyType),
					ID:   keyMaterial.KeyPath,
				})
			}
		}
	}
	return keyIDs, nil
}

func (ls *localStorageFileProvider) searchPrivateKey(ctx context.Context, keyID KeyID) (string, error) {
	localStorageFileContent, err := readContentFile(ctx, ls.file)
	if err != nil {
		return "", err
	}
	for _, keyMaterial := range localStorageFileContent {
		if keyMaterial.KeyPath == keyID.ID {
			return keyMaterial.PrivateKey, nil
		}
	}
	return "", errors.New("key not found")
}

func readContentFile(ctx context.Context, file string) ([]localStorageProviderFileContent, error) {
	fileContent, err := os.ReadFile(file)
	if err != nil {
		log.Error(ctx, "cannot read file", "err", err, "file", file)
		return nil, err
	}

	var localStorageFileContent []localStorageProviderFileContent
	if err := json.Unmarshal(fileContent, &localStorageFileContent); err != nil {
		log.Error(ctx, "cannot unmarshal file content", "err", err)
		return nil, err
	}

	return localStorageFileContent, nil
}
