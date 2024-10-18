package kms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/log"
)

type secretStorageProviderKeyMaterial struct {
	KeyType    string `json:"key_type"`
	KeyPath    string `json:"key_path"`
	PrivateKey string `json:"private_key"`
}

// AwsSecretStorageProviderConfig is a config for AwsSecretStorageProvider
// AccessKey and SecretKey are the AWS credentials
type AwsSecretStorageProviderConfig struct {
	AccessKey string
	SecretKey string
	Region    string
}

type awsSecretStorageProvider struct {
	secretManager *secretsmanager.Client
}

// NewAwsSecretStorageProvider creates a new instance of AwsSecretStorageProvider
func NewAwsSecretStorageProvider(ctx context.Context, conf AwsSecretStorageProviderConfig) (*awsSecretStorageProvider, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(conf.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(conf.AccessKey, conf.SecretKey, "")),
	)
	if err != nil {
		log.Error(ctx, "error loading AWS config", "err", err)
		return nil, err
	}
	return &awsSecretStorageProvider{
		secretManager: secretsmanager.NewFromConfig(cfg),
	}, nil
}

func (a *awsSecretStorageProvider) SaveKeyMaterial(ctx context.Context, keyMaterial map[string]string, id string) error {
	secretName, err := newtSecretName(id)
	if err != nil {
		return err
	}
	log.Info(ctx, "SaveKeyMaterial", "secretName", secretName)
	awsSecretStorageKeyMaterial := secretStorageProviderKeyMaterial{
		KeyPath:    id,
		KeyType:    keyMaterial[jsonKeyType],
		PrivateKey: keyMaterial[jsonKeyData],
	}
	secretValue, err := json.Marshal(awsSecretStorageKeyMaterial)
	if err != nil {
		return err
	}
	input := &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(string(secretValue)),
	}
	_, err = a.secretManager.CreateSecret(ctx, input)
	return err
}

func (a *awsSecretStorageProvider) searchByIdentity(ctx context.Context, identity w3c.DID, keyType KeyType) ([]KeyID, error) {
	secretName := getSecretNameForKeyTypeAndIdentity(keyType, identity)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	result, err := a.secretManager.GetSecretValue(ctx, input)
	if err != nil {
		log.Error(ctx, "error getting secret value", "err", err)
		return []KeyID{}, nil
	}
	valueAsBytes := []byte(aws.ToString(result.SecretString))
	var secret secretStorageProviderKeyMaterial
	err = json.Unmarshal(valueAsBytes, &secret)
	if err != nil {
		log.Error(ctx, "error unmarshalling secret value", "err", err)
		return nil, err
	}

	keyID := KeyID{
		Type: KeyType(secret.KeyType),
		ID:   secret.KeyPath,
	}
	return []KeyID{keyID}, nil
}

func (a *awsSecretStorageProvider) searchPrivateKey(ctx context.Context, keyID KeyID) (string, error) {
	encodedSecretName, err := getSecretNameForKeyID(keyID)
	if err != nil {
		return "", err
	}
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(encodedSecretName),
	}
	result, err := a.secretManager.GetSecretValue(ctx, input)
	if err != nil {
		log.Error(ctx, "error getting secret value", "err", err)
		return "", errors.New("error getting secret value from AWS")
	}

	var secretValue secretStorageProviderKeyMaterial
	if err := json.Unmarshal([]byte(aws.ToString(result.SecretString)), &secretValue); err != nil {
		return "", err
	}
	return secretValue.PrivateKey, nil
}

// newtSecretName returns the secret name for the given key id
// for a given id did/ETH:PRIVATE_KEY, the secret name will be ETH/did:PRIVATE_KEY
// for a given id did/BJJ:PRIVATE_KEY, the secret name will be BJJ/did:PRIVATE_KEY
// the secret name is base64 encoded
func newtSecretName(id string) (string, error) {
	idParts := strings.Split(id, "/")
	newId := ""
	if len(idParts) != partsNumber {
		return "", errors.New("invalid key id")
	}
	did := idParts[0]
	idParts = strings.Split(idParts[1], ":")
	if len(idParts) != partsNumber {
		return "", errors.New("invalid key id")
	}
	keyType := idParts[0]
	newId = fmt.Sprintf("%s/%s", keyType, did)
	secretName := base64.StdEncoding.EncodeToString([]byte(newId))
	return secretName, nil
}

// getSecretNameForKeyTypeAndIdentity returns the secret name for the given key type and identity
// the secret name is base64 encoded
// for a given keyType and identity, the secret name will be keyType/identity
// for instance ETH/did:example:1234 will be returned as base64 encoded string
func getSecretNameForKeyTypeAndIdentity(keyType KeyType, identity w3c.DID) string {
	id := fmt.Sprintf("%s/%s", keyType, identity.String())
	secretName := base64.StdEncoding.EncodeToString([]byte(id))
	return secretName
}

// getSecretNameForKeyID returns the secret name for the given key id
// the secret name is base64 encoded
func getSecretNameForKeyID(keyID KeyID) (string, error) {
	const partsNumber1 = 1
	secretName := ""
	keyIDParts := strings.Split(keyID.ID, "/")
	if len(keyIDParts) == partsNumber1 || len(keyIDParts) != partsNumber {
		return "", errors.New("invalid key id. expected format: did:example:1234/ETH:PRIVATE or pbkey")
	}
	secretName = fmt.Sprintf("%s/%s", keyID.Type, keyIDParts[0])
	encodedSecretName := base64.StdEncoding.EncodeToString([]byte(secretName))
	return encodedSecretName, nil
}
