package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awskms "github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/ethereum/go-ethereum/crypto"
	vault "github.com/hashicorp/vault/api"
	"github.com/joho/godotenv"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
)

const (
	issuerKMSETHProvider                = "ISSUER_KMS_ETH_PROVIDER"
	issuerPublishKeyPath                = "ISSUER_PUBLISH_KEY_PATH"
	issuerKmsPluginLocalStorageFilePath = "ISSUER_KMS_PROVIDER_LOCAL_STORAGE_FILE_PATH"
	issuerKeyStoreToken                 = "ISSUER_KEY_STORE_TOKEN"
	issuerKeyStoreAddress               = "ISSUER_KEY_STORE_ADDRESS"
	issuerKeyStorePluginIden3MountPath  = "ISSUER_KEY_STORE_PLUGIN_IDEN3_MOUNT_PATH"
	issuerVaultUserPassAuthEnabled      = "ISSUER_VAULT_USERPASS_AUTH_ENABLED"
	issuerVaultUserPassAuthPasword      = "ISSUER_VAULT_USERPASS_AUTH_PASSWORD"
	awsAccessKey                        = "ISSUER_KMS_AWS_ACCESS_KEY"
	awsSecretKey                        = "ISSUER_KMS_AWS_SECRET_KEY"
	awsRegion                           = "ISSUER_KMS_AWS_REGION"
	awsURL                              = "ISSUER_KMS_AWS_URL"

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := godotenv.Load(envFile)
	if err != nil {
		log.Error(ctx, "Error loading .env-issuer file")
	}
	issuerKMSETHProviderToUse := os.Getenv(issuerKMSETHProvider)
	issuerKmsPluginLocalStorageFilePath := os.Getenv(issuerKmsPluginLocalStorageFilePath)

	fPrivateKey := flag.String("privateKey", "", "private key")
	flag.Parse()

	log.Info(ctx, "eth kms provider to use:", "provider", issuerKMSETHProviderToUse)

	if err := validate(issuerKMSETHProviderToUse, fPrivateKey, ctx); err != nil {
		return
	}

	issuerPublishKeyPathVar, err := getPrivateKey(ctx)
	if err != nil {
		return
	}

	if issuerKmsPluginLocalStorageFilePath == "" {
		issuerKmsPluginLocalStorageFilePath = pluginFolderPath
	}

	material := make(map[string]string)
	material[jsonKeyPath] = issuerPublishKeyPathVar
	material[jsonKeyType] = ethereum

	if issuerKMSETHProviderToUse == config.LocalStorage {
		material[jsonPrivateKey] = *fPrivateKey
		if err := saveKeyMaterialToFile(ctx, issuerKmsPluginLocalStorageFilePath, kms.LocalStorageFileName, material); err != nil {
			log.Error(ctx, "cannot save key material to file", "err", err)
			return
		}
		log.Info(ctx, "private key saved to file:", "path:", kms.LocalStorageFileName)
		return
	}

	if issuerKMSETHProviderToUse == config.Vault {
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

	if issuerKMSETHProviderToUse == config.AWSSM {
		awsAccessKey := os.Getenv(awsAccessKey)
		awsSecretKey := os.Getenv(awsSecretKey)
		awsRegion := os.Getenv(awsRegion)

		if awsAccessKey == "" || awsSecretKey == "" || awsRegion == "" {
			log.Error(ctx, "aws access key, aws secret key, or aws region is not set")
			return
		}

		cfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(awsRegion),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretKey, "")),
		)
		if err != nil {
			log.Error(ctx, "error loading AWSSM config", "err", err)
			return
		}

		var options []func(*secretsmanager.Options)
		if strings.ToLower(awsRegion) == "local" {
			awsURLEndpoint := os.Getenv(awsURL)
			options = make([]func(*secretsmanager.Options), 1)
			options[0] = func(o *secretsmanager.Options) {
				o.BaseEndpoint = aws.String(awsURLEndpoint)
			}
		}
		secretManager := secretsmanager.NewFromConfig(cfg, options...)
		secretName := base64.StdEncoding.EncodeToString([]byte(issuerPublishKeyPathVar))

		material[jsonPrivateKey] = *fPrivateKey
		secretValue, err := json.Marshal(material)
		if err != nil {
			log.Error(ctx, "cannot marshal secret value", "err", err)
			return
		}

		input := &secretsmanager.CreateSecretInput{
			Name:         aws.String(secretName),
			SecretString: aws.String(string(secretValue)),
		}
		_, err = secretManager.CreateSecret(ctx, input)
		if err != nil {
			log.Error(ctx, "cannot save key material to aws", "err", err)
			return
		}
		log.Info(ctx, "private key saved to aws:", "path:", issuerPublishKeyPathVar)
		return
	}

	if issuerKMSETHProviderToUse == config.AWSKMS {
		awsAccessKey := os.Getenv(awsAccessKey)
		awsSecretKey := os.Getenv(awsSecretKey)
		awsRegion := os.Getenv(awsRegion)
		awsURLEndpoint := os.Getenv(awsURL)

		if awsAccessKey == "" || awsSecretKey == "" || awsRegion == "" {
			log.Error(ctx, "aws access key, aws secret key, or aws region is not set")
			return
		}

		keyId, err := createEmptyKey(ctx, awsAccessKey, awsSecretKey, awsRegion, awsURLEndpoint, issuerPublishKeyPathVar)
		if err != nil {
			log.Error(ctx, "cannot create empty key", "err", err)
			return
		}
		log.Info(ctx, "key created", "keyId", *keyId)
		return
	}
}

func getPrivateKey(ctx context.Context) (string, error) {
	issuerPublishKeyPathVar := os.Getenv(issuerPublishKeyPath)
	if issuerPublishKeyPathVar == "" {
		log.Error(ctx, "ISSUER_PUBLISH_KEY_PATH is not set")
		return "", errors.New("ISSUER_PUBLISH_KEY_PATH is not set")
	}
	return issuerPublishKeyPathVar, nil
}

func validate(issuerKMSETHProviderToUse string, fPrivateKey *string, ctx context.Context) error {
	if issuerKMSETHProviderToUse != config.AWSKMS && *fPrivateKey == "" {
		log.Error(ctx, "private key is required. Please provide private key: --privateKey=<private key>")
		return errors.New("private key is required")
	}

	if issuerKMSETHProviderToUse != config.AWSKMS {
		_, err := crypto.HexToECDSA(*fPrivateKey)
		if err != nil {
			log.Error(ctx, "cannot convert private key to ECDSA", "err", err)
			return errors.New("cannot convert private key to ECDSA")
		}
	}

	if issuerKMSETHProviderToUse != config.LocalStorage && issuerKMSETHProviderToUse != config.Vault && issuerKMSETHProviderToUse != config.AWSSM && issuerKMSETHProviderToUse != config.AWSKMS {
		log.Error(ctx, "kms eth provider is invalid, supported values are: localstorage, vault, aws-sm and aws-kms")
		return errors.New("kms eth provider is invalid")
	}
	return nil
}

//
//nolint:unused
func createEmptyKey(ctx context.Context, awsAccessKey, awsSecretKey, awsRegion string, awsURL string, privateKeyAlias string) (*string, error) {
	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(awsRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretKey, "")),
	)
	if err != nil {
		log.Error(ctx, "cannot load aws config", "err", err)
		return nil, err
	}

	var options []func(*awskms.Options)
	if strings.ToLower(awsRegion) == "local" {
		options = make([]func(*awskms.Options), 1)
		options[0] = func(o *awskms.Options) {
			o.BaseEndpoint = aws.String(awsURL)
		}
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

func saveKeyMaterialToFile(ctx context.Context, folderPath, file string, keyMaterial map[string]string) error {
	localStorageFileContent, err := readContentFile(ctx, folderPath, file)
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
	filePath := filepath.Join(folderPath, file)
	// nolint: all
	if err := os.WriteFile(filePath, newFileContent, 0644); err != nil {
		log.Error(ctx, "cannot write file", "err", err)
		return err
	}

	return nil
}

func readContentFile(ctx context.Context, folderPath, fileName string) ([]localStorageBJJKeyProviderFileContent, error) {
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("error creating folder: %v", err)
	}
	filePath := filepath.Join(folderPath, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("error creating file: %v", err)
		}
		fileContent := []byte("[]")
		if _, err := file.Write(fileContent); err != nil {
			return nil, fmt.Errorf("error initiliazing file: %v", err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Error(ctx, "error closing file", "err", err)
			}
		}(file)
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Error(ctx, "cannot read file", "err", err, "file", filePath)
		return nil, err
	}

	var localStorageFileContent []localStorageBJJKeyProviderFileContent
	if err := json.Unmarshal(fileContent, &localStorageFileContent); err != nil {
		log.Error(ctx, "cannot unmarshal file content", "err", err)
		return nil, err
	}

	return localStorageFileContent, nil
}
