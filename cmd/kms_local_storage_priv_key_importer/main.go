package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"

	"github.com/joho/godotenv"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
)

const (
	// IssuerKmsPlugin is the environment variable that defines the issuer kms plugin
	IssuerKmsPlugin = "ISSUER_KMS_PLUGIN"
	// IssuerKmsPluginLocalStorageFilePath is the environment variable that defines the issuer kms plugin local storage file path
	IssuerKmsPluginLocalStorageFilePath = "ISSUER_KMS_PLUGIN_LOCAL_STORAGE_FILE_PATH"
	jsonKeyPath                         = "key_path"
	jsonKeyType                         = "key_type"
	jsonKeyData                         = "key_data"
	pbkey                               = "pbkey"
	ethereum                            = "ethereum"
	pluginFolderPath                    = "./localstoragekeys"
	envFile                             = ".env-issuer"
)

type localStorageBJJKeyProviderFileContent struct {
	KeyType    string `json:"key_type"`
	KeyPath    string `json:"key_path"`
	PrivateKey string `json:"private_key"`
}

func main() {
	fPrivateKey := flag.String("privateKey", "", "metamask private key")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.Parse()

	if *fPrivateKey == "" {
		log.Error(ctx, "private key is required")
		return
	}

	if os.Getenv(IssuerKmsPlugin) == "" {
		err := godotenv.Load(envFile)
		if err != nil {
			log.Error(ctx, "Error loading .env-issuer file")
		}
	}

	if os.Getenv(IssuerKmsPlugin) == "" || os.Getenv(IssuerKmsPlugin) != config.LocalStorage {
		log.Error(ctx, "issuer kms plugin is not set or is not local storage", "plugin: ", os.Getenv(IssuerKmsPlugin))
		return
	}

	if os.Getenv(IssuerKmsPluginLocalStorageFilePath) == "" {
		err := godotenv.Load(envFile)
		if err != nil {
			log.Error(ctx, "Error loading .env-issuer file")
		}
	}

	pluginFolderPathVar := ""
	if os.Getenv(IssuerKmsPluginLocalStorageFilePath) != "" {
		pluginFolderPathVar = os.Getenv(IssuerKmsPluginLocalStorageFilePath)
	} else {
		pluginFolderPathVar = pluginFolderPath
	}

	_, err := kms.OpenLocalPath(pluginFolderPathVar)
	if err != nil {
		log.Error(ctx, "cannot open local storage path", "err", err)
		return
	}

	material := make(map[string]string)
	material[jsonKeyPath] = pbkey
	material[jsonKeyType] = ethereum
	material[jsonKeyData] = *fPrivateKey

	file := pluginFolderPath + "/" + kms.LocalStorageFileName
	if err := saveKeyMaterialToFile(ctx, file, material); err != nil {
		log.Error(ctx, "cannot save key material to file", "err", err)
		return
	}

	log.Info(ctx, "private key saved to file")
}

func saveKeyMaterialToFile(ctx context.Context, file string, keyMaterial map[string]string) error {
	localStorageFileContent, err := readContentFile(ctx, file)
	if err != nil {
		return err
	}

	for _, keyMaterialContentFile := range localStorageFileContent {
		if keyMaterialContentFile.KeyPath == keyMaterial[jsonKeyPath] {
			log.Error(ctx, "key already exists", "keyPath", keyMaterial[jsonKeyPath])
			return errors.New("key already exists")
		}
	}

	localStorageFileContent = append(localStorageFileContent, localStorageBJJKeyProviderFileContent{
		KeyPath:    keyMaterial[jsonKeyPath],
		KeyType:    keyMaterial[jsonKeyType],
		PrivateKey: keyMaterial[jsonKeyData],
	})

	newFileContent, err := json.Marshal(localStorageFileContent)
	if err != nil {
		log.Error(ctx, "cannot marshal file content", "err", err)
		return err
	}
	// nolint: all
	if err := os.WriteFile(file, newFileContent, 0644); err != nil {
		log.Error(ctx, "cannot write file", "err", err)
		return err
	}

	return nil
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
