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

type fileStorageManager struct {
	file string
}

// NewFileStorageManager - creates new local storage file manager
func NewFileStorageManager(file string) *fileStorageManager {
	return &fileStorageManager{file}
}

func (ls *fileStorageManager) SaveKeyMaterial(ctx context.Context, keyMaterial map[string]string, id string) error {
	localStorageFileContent, err := readContentFile(ctx, ls.file)
	if err != nil {
		return err
	}
	localStorageFileContent = append(localStorageFileContent, localStorageProviderFileContent{
		KeyPath:    id,
		KeyType:    convertFromKeyType(KeyType(keyMaterial[jsonKeyType])),
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

func (ls *fileStorageManager) searchByIdentity(ctx context.Context, identity w3c.DID, keyType KeyType) ([]KeyID, error) {
	keyTypeToRead := convertFromKeyType(keyType)
	localStorageFileContent, err := readContentFile(ctx, ls.file)
	if err != nil {
		return nil, err
	}
	keyIDs := make([]KeyID, 0)
	for _, keyMaterial := range localStorageFileContent {
		keyParts := strings.Split(keyMaterial.KeyPath, "/")
		if len(keyParts) != partsNumber && len(keyParts) != partsNumber3 {
			continue
		}
		if keyParts[0] == identity.String() || keyParts[1] == identity.String() {
			if keyMaterial.KeyType == keyTypeToRead {
				keyIDs = append(keyIDs, KeyID{
					Type: convertToKeyType(keyTypeToRead),
					ID:   keyMaterial.KeyPath,
				})
			}
		}
	}
	return keyIDs, nil
}

func (ls *fileStorageManager) searchPrivateKey(ctx context.Context, keyID KeyID) (string, error) {
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

func (ls *fileStorageManager) deleteKeyMaterial(ctx context.Context, keyID KeyID) error {
	localStorageFileContent, err := readContentFile(ctx, ls.file)
	if err != nil {
		return err
	}
	for i, keyMaterial := range localStorageFileContent {
		if keyMaterial.KeyPath == keyID.ID {
			localStorageFileContent = append(localStorageFileContent[:i], localStorageFileContent[i+1:]...)
			break
		}
	}

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

func (ls *fileStorageManager) getKeyMaterial(ctx context.Context, keyID KeyID) (map[string]string, error) {
	localStorageFileContent, err := readContentFile(ctx, ls.file)
	if err != nil {
		return nil, err
	}
	for _, keyMaterial := range localStorageFileContent {
		if keyMaterial.KeyPath == keyID.ID {
			return map[string]string{
				jsonKeyType: keyMaterial.KeyType,
				jsonKeyData: keyMaterial.PrivateKey,
			}, nil
		}
	}
	return nil, ErrKeyNotFound
}
