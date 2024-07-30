package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awskms "github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/ethereum/go-ethereum/crypto"
	vault "github.com/hashicorp/vault/api"
	"github.com/joho/godotenv"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
)

const (
	issuerKMSETHPlugin                  = "ISSUER_KMS_ETH_PROVIDER"
	issuerPublishKeyPath                = "ISSUER_PUBLISH_KEY_PATH"
	issuerKmsPluginLocalStorageFilePath = "ISSUER_KMS_PROVIDER_LOCAL_STORAGE_FILE_PATH"
	issuerKeyStoreToken                 = "ISSUER_KEY_STORE_TOKEN"
	issuerKeyStoreAddress               = "ISSUER_KEY_STORE_ADDRESS"
	issuerKeyStorePluginIden3MountPath  = "ISSUER_KEY_STORE_PLUGIN_IDEN3_MOUNT_PATH"
	issuerVaultUserPassAuthEnabled      = "ISSUER_VAULT_USERPASS_AUTH_ENABLED"
	issuerVaultUserPassAuthPasword      = "ISSUER_VAULT_USERPASS_AUTH_PASSWORD"
	aWSAccessKey                        = "ISSUER_KMS_ETH_PLUGIN_AWS_ACCESS_KEY"
	aWSSecretKey                        = "ISSUER_KMS_ETH_PLUGIN_AWS_SECRET_KEY"
	aWSRegion                           = "ISSUER_KMS_ETH_PLUGIN_AWS_REGION"

	jsonKeyPath      = "key_path"
	jsonKeyType      = "key_type"
	jsonPrivateKey   = "private_key"
	ethereum         = "ethereum"
	pluginFolderPath = "./localstoragekeys"
	envFile          = ".env-issuer"
)

type localStorageBJJKeyProviderFileContent struct {
	KeyType    string `json:"key_type"`
	KeyPath    string `json:"key_path"`
	PrivateKey string `json:"private_key"`
}

// This is a tool to import ethereum private key to different kms.
func main() {
	fPrivateKey := flag.String("privateKey", "", "metamask private key")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flag.Parse()
	if *fPrivateKey == "" {
		log.Error(ctx, "private key is required")
		return
	}

	_, err := crypto.HexToECDSA(*fPrivateKey)
	if err != nil {
		log.Error(ctx, "cannot convert private key to ECDSA", "err", err)
		return
	}

	err = godotenv.Load(envFile)
	if err != nil {
		log.Error(ctx, "Error loading .env-issuer file")
	}

	issuerKMSEthPluginVar := os.Getenv(issuerKMSETHPlugin)
	issuerKmsPluginLocalStorageFilePath := os.Getenv(issuerKmsPluginLocalStorageFilePath)

	if issuerKMSEthPluginVar != config.LocalStorage && issuerKMSEthPluginVar != config.Vault && issuerKMSEthPluginVar != config.AWS {
		log.Error(ctx, "issuer kms eth provider is not set or is not localstorage or vault or aws", "plugin: ", issuerKMSEthPluginVar)
		return
	}

	if issuerKmsPluginLocalStorageFilePath == "" {
		issuerKmsPluginLocalStorageFilePath = pluginFolderPath
	}

	issuerPublishKeyPathVar := os.Getenv(issuerPublishKeyPath)
	if issuerPublishKeyPathVar == "" {
		log.Error(ctx, "ISSUER_PUBLISH_KEY_PATH is not set")
		return
	}

	material := make(map[string]string)
	material[jsonKeyPath] = issuerPublishKeyPathVar
	material[jsonKeyType] = ethereum
	material[jsonPrivateKey] = *fPrivateKey

	if issuerKMSEthPluginVar == config.LocalStorage {
		file := issuerKmsPluginLocalStorageFilePath + "/" + kms.LocalStorageFileName
		if err := saveKeyMaterialToFile(ctx, file, material); err != nil {
			log.Error(ctx, "cannot save key material to file", "err", err)
			return
		}

		log.Info(ctx, "private key saved to file:", "path:", file)
		return
	}

	if issuerKMSEthPluginVar == config.Vault {
		var vaultCli *vault.Client
		var vaultErr error
		vaultTokenVar := os.Getenv(issuerKeyStoreToken)
		vaultAddressVar := os.Getenv(issuerKeyStoreAddress)
		vaultPluginIden3MountPath := os.Getenv(issuerKeyStorePluginIden3MountPath)
		userPassEnableVar := os.Getenv(issuerVaultUserPassAuthEnabled)
		userPassEnableAuthPassVar := os.Getenv(issuerVaultUserPassAuthPasword)
		userPassEnableVarBoolValue, err := strconv.ParseBool(userPassEnableVar)
		if err != nil {
			log.Error(ctx, "cannot parse userpass auth enabled value", "err", err)
			return
		}

		if userPassEnableVarBoolValue {
			if userPassEnableAuthPassVar == "" {
				log.Error(ctx, "userpass auth enabled but password is not set")
				return
			}
		}

		if !userPassEnableVarBoolValue {
			if vaultTokenVar == "" {
				log.Error(ctx, "vault token is not set")
				return
			}
		}

		vaultCli, vaultErr = providers.VaultClient(ctx, providers.Config{
			UserPassAuthEnabled: userPassEnableVarBoolValue,
			Pass:                userPassEnableAuthPassVar,
			Address:             vaultAddressVar,
			Token:               vaultTokenVar,
		})
		if vaultErr != nil {
			log.Error(ctx, "cannot initialize vault client", "err", vaultErr)
			return
		}
		data := make(map[string]any)
		data[jsonKeyType] = ethereum
		data[jsonPrivateKey] = *fPrivateKey

		_, err = vaultCli.Logical().Write(path.Join(vaultPluginIden3MountPath, "import", issuerPublishKeyPathVar), data)
		if err != nil {
			log.Error(ctx, "cannot save key material to vault", "err", err)
			return
		}

		log.Info(ctx, "private key saved to vault:", "path:", issuerPublishKeyPathVar)
		return
	}

	if issuerKMSEthPluginVar == config.AWS {
		awsAccessKey := os.Getenv(aWSAccessKey)
		awsSecretKey := os.Getenv(aWSSecretKey)
		awsRegion := os.Getenv(aWSRegion)

		if awsAccessKey == "" || awsSecretKey == "" || awsRegion == "" {
			log.Error(ctx, "aws access key, aws secret key, or aws region is not set")
			return
		}

		keyId, err := createEmptyKey(ctx, awsAccessKey, awsSecretKey, awsRegion, issuerPublishKeyPathVar)
		if err != nil {
			log.Error(ctx, "cannot create empty key", "err", err)
			return
		}
		log.Info(ctx, "key created", "keyId", *keyId)
		return
	}
}

func createEmptyKey(ctx context.Context, awsAccessKey, awsSecretKey, awsRegion string, privateKeyAlias string) (*string, error) {
	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(awsRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretKey, "")),
	)
	if err != nil {
		log.Error(ctx, "cannot load aws config", "err", err)
		return nil, err
	}

	svc := awskms.NewFromConfig(cfg)
	input := &awskms.CreateKeyInput{
		KeySpec:     types.KeySpecEccSecgP256k1,
		KeyUsage:    types.KeyUsageTypeSignVerify,
		Origin:      types.OriginTypeExternal,
		Description: aws.String("imported key"),
	}

	result, err := svc.CreateKey(ctx, input)
	if err != nil {
		log.Error(ctx, "cannot create key", "err", err)
		return nil, err
	}

	alias := "alias/" + privateKeyAlias
	inputAlias := &awskms.CreateAliasInput{
		AliasName:   aws.String(alias),
		TargetKeyId: result.KeyMetadata.Arn,
	}

	_, err = svc.CreateAlias(ctx, inputAlias)
	if err != nil {
		return nil, fmt.Errorf("failed to create alias: %v", err)
	}

	log.Info(ctx, "alias created:", "alias:", alias)
	return result.KeyMetadata.KeyId, nil
}

func saveKeyMaterialToFile(ctx context.Context, file string, keyMaterial map[string]string) error {
	localStorageFileContent, err := readContentFile(ctx, file)
	if err != nil {
		return err
	}

	for _, keyMaterialContentFile := range localStorageFileContent {
		if keyMaterialContentFile.KeyPath == keyMaterial[jsonKeyPath] {
			log.Error(ctx, "private key already exists", "keyPath", keyMaterial[jsonKeyPath])
			return errors.New("private key already exists")
		}
	}

	localStorageFileContent = append(localStorageFileContent, localStorageBJJKeyProviderFileContent{
		KeyPath:    keyMaterial[jsonKeyPath],
		KeyType:    keyMaterial[jsonKeyType],
		PrivateKey: keyMaterial[jsonPrivateKey],
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
