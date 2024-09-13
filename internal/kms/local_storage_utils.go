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

// LocalStorageFileManager - interface for managing local storage
type LocalStorageFileManager interface {
	saveKeyMaterialToFile(ctx context.Context, keyMaterial map[string]string, id string) error
	searchByIdentityInFile(ctx context.Context, identity w3c.DID, keyType KeyType) ([]KeyID, error)
	searchKeyMaterialInFileAndReplace(ctx context.Context, id string, identity w3c.DID) error
	searchPrivateKeyInFile(ctx context.Context, keyID KeyID) (string, error)
}

const (
	partsNumber3 = 3
	babyjubjub   = "babyjubjub"
	ethereum     = "ethereum"
)

type localStorageFileManager struct {
	file string
}

// NewLocalStorageFileManager - creates new local storage file manager
func NewLocalStorageFileManager(file string) LocalStorageFileManager {
	return &localStorageFileManager{file}
}

func (ls *localStorageFileManager) saveKeyMaterialToFile(ctx context.Context, keyMaterial map[string]string, id string) error {
	localStorageFileContent, err := readContentFile(ctx, ls.file)
	if err != nil {
		return err
	}

	keyTypeToSave := ""
	if keyMaterial[jsonKeyType] == string(KeyTypeBabyJubJub) {
		keyTypeToSave = babyjubjub
	} else if keyMaterial[jsonKeyType] == string(KeyTypeEthereum) {
		keyTypeToSave = ethereum
	} else {
		return errors.New("unknown key type")
	}

	localStorageFileContent = append(localStorageFileContent, localStorageBJJKeyProviderFileContent{
		KeyPath:    id,
		KeyType:    keyTypeToSave,
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

func (ls *localStorageFileManager) searchByIdentityInFile(ctx context.Context, identity w3c.DID, keyType KeyType) ([]KeyID, error) {
	localStorageFileContent, err := readContentFile(ctx, ls.file)
	if err != nil {
		return nil, err
	}
	keyIDs := make([]KeyID, 0)

	keyTypeToRead := ""
	if keyType == KeyTypeBabyJubJub {
		keyTypeToRead = babyjubjub
	} else if keyType == KeyTypeEthereum {
		keyTypeToRead = ethereum
	} else {
		return nil, errors.New("unknown key type")
	}

	for _, keyMaterial := range localStorageFileContent {
		keyParts := strings.Split(keyMaterial.KeyPath, "/")
		if len(keyParts) != partsNumber && len(keyParts) != partsNumber3 {
			continue
		}
		if keyParts[0] == identity.String() || keyParts[1] == identity.String() {
			if keyMaterial.KeyType == keyTypeToRead {
				keyTypeLoaded := ""
				if keyMaterial.KeyType == babyjubjub {
					keyTypeLoaded = string(KeyTypeBabyJubJub)
				} else if keyMaterial.KeyType == ethereum {
					keyTypeLoaded = string(KeyTypeEthereum)
				} else {
					continue
				}
				keyIDs = append(keyIDs, KeyID{
					Type: KeyType(keyTypeLoaded),
					ID:   keyMaterial.KeyPath,
				})
			}
		}
	}
	return keyIDs, nil
}

func (ls *localStorageFileManager) searchKeyMaterialInFileAndReplace(ctx context.Context, id string, identity w3c.DID) error {
	localStorageFileContent, err := readContentFile(ctx, ls.file)
	if err != nil {
		return err
	}

	for i, keyMaterial := range localStorageFileContent {
		if keyMaterial.KeyPath == id {
			keyMaterial.KeyPath = identity.String() + "/" + keyMaterial.KeyPath
			localStorageFileContent[i] = keyMaterial
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
	}

	return errors.New("key not found")
}

func (ls *localStorageFileManager) searchPrivateKeyInFile(ctx context.Context, keyID KeyID) (string, error) {
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

func readContentFile(ctx context.Context, file string) ([]localStorageBJJKeyProviderFileContent, error) {
	fileContent, err := os.ReadFile(file)
	if err != nil {
		log.Error(ctx, "cannot read file", "err", err, "file", file)
		return nil, err
	}

	var localStorageFileContent []localStorageBJJKeyProviderFileContent
	if err := json.Unmarshal(fileContent, &localStorageFileContent); err != nil {
		log.Error(ctx, "cannot unmarshal file content", "err", err)
		return nil, err
	}

	return localStorageFileContent, nil
}
