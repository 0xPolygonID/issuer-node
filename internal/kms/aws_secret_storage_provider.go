package kms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
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
	URL       string
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

	var options []func(*secretsmanager.Options)
	if strings.ToLower(conf.Region) == "local" {
		options = make([]func(*secretsmanager.Options), 1)
		options[0] = func(o *secretsmanager.Options) {
			o.BaseEndpoint = aws.String(conf.URL)
		}
	}

	return &awsSecretStorageProvider{
		secretManager: secretsmanager.NewFromConfig(cfg, options...),
	}, nil
}

func (a *awsSecretStorageProvider) SaveKeyMaterial(ctx context.Context, keyMaterial map[string]string, id string) error {
	encodedSecretName := base64.StdEncoding.EncodeToString([]byte(id))
	log.Info(ctx, "SaveKeyMaterial", "secretName", id)
	awsSecretStorageKeyMaterial := secretStorageProviderKeyMaterial{
		KeyPath:    id,
		KeyType:    convertFromKeyType(KeyType(keyMaterial[jsonKeyType])),
		PrivateKey: keyMaterial[jsonKeyData],
	}
	secretValue, err := json.Marshal(awsSecretStorageKeyMaterial)
	if err != nil {
		return err
	}

	keyTypesParts := strings.Split(id, "/")
	if len(keyTypesParts) != partsNumber {
		return errors.New("invalid key id")
	}

	input := &secretsmanager.CreateSecretInput{
		Name:         aws.String(encodedSecretName),
		SecretString: aws.String(string(secretValue)),
		Description:  aws.String("Secret associated with the issuer node"),
		Tags: []types.Tag{
			{
				Key:   aws.String("keyType"),
				Value: aws.String(awsSecretStorageKeyMaterial.KeyType),
			},
			{
				Key:   aws.String("did"),
				Value: aws.String(keyTypesParts[0]),
			},
		},
	}
	_, err = a.secretManager.CreateSecret(ctx, input)
	return err
}

func (a *awsSecretStorageProvider) searchByIdentity(ctx context.Context, identity w3c.DID, keyType KeyType) ([]KeyID, error) {
	keyTypeToRead := convertFromKeyType(keyType)
	input := &secretsmanager.ListSecretsInput{
		Filters: []types.Filter{
			{
				Key:    types.FilterNameStringTypeTagValue,
				Values: []string{identity.String()},
			},
		},
	}

	result, err := a.secretManager.ListSecrets(ctx, input)
	if err != nil {
		log.Error(ctx, "error listing secrets", "err", err)
		return nil, err
	}

	keyIDs := make([]KeyID, 0)
	for _, secret := range result.SecretList {
		if secret.Tags == nil || len(secret.Tags) != 2 {
			continue
		}
		if aws.ToString(secret.Tags[0].Value) != keyTypeToRead {
			continue
		}
		secretName := aws.ToString(secret.Name)

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
			Type: convertToKeyType(secret.KeyType),
			ID:   secret.KeyPath,
		}
		keyIDs = append(keyIDs, keyID)
	}
	return keyIDs, nil
}

func (a *awsSecretStorageProvider) searchPrivateKey(ctx context.Context, keyID KeyID) (string, error) {
	encodedSecretName := base64.StdEncoding.EncodeToString([]byte(keyID.ID))
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

func (a *awsSecretStorageProvider) deleteKeyMaterial(ctx context.Context, keyID KeyID) error {
	encodedSecretName := base64.StdEncoding.EncodeToString([]byte(keyID.ID))
	input := &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(encodedSecretName),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	}
	_, err := a.secretManager.DeleteSecret(ctx, input)
	return err
}

func (a *awsSecretStorageProvider) getKeyMaterial(ctx context.Context, keyID KeyID) (map[string]string, error) {
	encodedSecretName := base64.StdEncoding.EncodeToString([]byte(keyID.ID))
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(encodedSecretName),
	}
	result, err := a.secretManager.GetSecretValue(ctx, input)
	if err != nil {
		log.Error(ctx, "error getting secret value", "err", err)
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
			return nil, ErrKeyNotFound
		}
		return nil, errors.New("error getting secret value from AWS")
	}

	var secretValue secretStorageProviderKeyMaterial
	if err := json.Unmarshal([]byte(aws.ToString(result.SecretString)), &secretValue); err != nil {
		return nil, err
	}

	return map[string]string{
		jsonKeyType: secretValue.KeyType,
		jsonKeyData: secretValue.PrivateKey,
	}, nil
}
