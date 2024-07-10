package kms

import (
	"context"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/log"
)

const aliasPrefix = "alias/"

type awsEthKeyProvider struct {
	keyType          KeyType
	reIdenKeyPathHex *regexp.Regexp // RE of key path bounded to identity
	kmsClient        *kms.Client
}

// AwEthKeyProviderConfig - configuration for AWS KMS Ethereum key provider
type AwEthKeyProviderConfig struct {
	AccessKey string
	SecretKey string
	Region    string
}

// NewAwsEthKeyProvider - creates new key provider for Ethereum keys stored in AWS KMS
func NewAwsEthKeyProvider(ctx context.Context, keyType KeyType, awsKmsEthKeyProviderConfig AwEthKeyProviderConfig) (KeyProvider, error) {
	keyTypeRE := regexp.QuoteMeta(string(keyType))
	reIdenKeyPathHex := regexp.MustCompile("^(?i).*/" + keyTypeRE + ":([a-f0-9]{64})$")
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(awsKmsEthKeyProviderConfig.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsKmsEthKeyProviderConfig.AccessKey,
			awsKmsEthKeyProviderConfig.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}
	svc := kms.NewFromConfig(cfg)
	return &awsEthKeyProvider{
		keyType:          keyType,
		reIdenKeyPathHex: reIdenKeyPathHex,
		kmsClient:        svc,
	}, nil
}

func (awsKeyProv *awsEthKeyProvider) New(identity *w3c.DID) (KeyID, error) {
	ctx := context.Background()
	keyID := KeyID{Type: awsKeyProv.keyType}

	input := &kms.CreateKeyInput{
		KeySpec:     types.KeySpecEccSecgP256k1,
		KeyUsage:    types.KeyUsageTypeSignVerify,
		Origin:      types.OriginTypeAwsKms,
		Description: aws.String("Key from issuer node"),
	}

	keyArn, err := awsKeyProv.kmsClient.CreateKey(ctx, input)
	if err != nil {
		log.Error(ctx, "failed to create key: %v", err)
		return KeyID{}, fmt.Errorf("failed to create key: %v", err)
	}
	log.Info(ctx, "keyArn: %v", keyArn.KeyMetadata.Arn)
	inputPublicKey := &kms.GetPublicKeyInput{
		KeyId: aws.String(*keyArn.KeyMetadata.Arn),
	}

	publicKeyResult, err := awsKeyProv.kmsClient.GetPublicKey(ctx, inputPublicKey)
	if err != nil {
		return KeyID{}, fmt.Errorf("failed to get public key: %v", err)
	}
	pubKeyHex := hex.EncodeToString(publicKeyResult.PublicKey)
	keyID.ID = keyPathForAws(identity, awsKeyProv.keyType, pubKeyHex)

	aliasName := aliasPrefix + keyID.ID
	err = awsKeyProv.createAlias(ctx, aliasName, *keyArn.KeyMetadata.KeyId)
	if err != nil {
		log.Error(ctx, "failed to create alias: %v", err)
		return KeyID{}, fmt.Errorf("failed to create alias: %v", err)
	}
	keyID.ID = aliasName
	return keyID, nil
}

// PublicKey returns public key for given keyID
func (awsKeyProv *awsEthKeyProvider) PublicKey(keyID KeyID) ([]byte, error) {
	ctx := context.Background()
	inputPublicKey := &kms.GetPublicKeyInput{
		KeyId: aws.String(keyID.ID),
	}
	publicKeyResult, err := awsKeyProv.kmsClient.GetPublicKey(ctx, inputPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %v", err)
	}
	return publicKeyResult.PublicKey, nil
}

// Sign signs data with keyID
func (awsKeyProv *awsEthKeyProvider) Sign(ctx context.Context, keyID KeyID, data []byte) ([]byte, error) {
	signInput := &kms.SignInput{
		KeyId:            aws.String(keyID.ID),
		Message:          data,
		MessageType:      types.MessageTypeDigest,
		SigningAlgorithm: types.SigningAlgorithmSpecEcdsaSha256,
	}

	result, err := awsKeyProv.kmsClient.Sign(ctx, signInput)
	if err != nil {
		log.Error(ctx, "failed to sign payload", "err:", err)
		return nil, fmt.Errorf("failed to sign payload: %v", err)
	}

	//nolint:all
	awsKMSPubKey, err := awsKeyProv.PublicKey(keyID)
	if err != nil {
		log.Error(ctx, "failed to get public key", "err:", err)
		return nil, fmt.Errorf("failed to get public key: %v", err)
	}

	// get ecdsa public key
	pk, err := DecodeAWSETHPubKey(ctx, awsKMSPubKey)
	if err != nil {
		log.Error(ctx, "failed to decode public key", "err:", err)
		return nil, fmt.Errorf("failed to decode public key: %v", err)
	}
	//TODO: Another option could be?: pubKeyBytes := secp256k1.S256().Marshal(pk.X, pk.Y)
	pubKeyBytes := crypto.FromECDSAPub(pk)
	signature, err := DecodeAWSETHSig(ctx, result.Signature, pubKeyBytes, data)
	if err != nil {
		log.Error(ctx, "failed to decode signature", "err:", err)
		return nil, fmt.Errorf("failed to decode signature: %v", err)

	}
	log.Info(ctx, "signature created:", "signature:", result.Signature)
	return signature, nil
}

// LinkToIdentity links key to identity
func (awsKeyProv *awsEthKeyProvider) LinkToIdentity(ctx context.Context, keyID KeyID, identity w3c.DID) (KeyID, error) {
	keyMetadata, err := awsKeyProv.getKeyInfoByAlias(ctx, keyID.ID)
	if err != nil {
		log.Error(ctx, "failed to get key metadata: %v", err)
		return KeyID{}, fmt.Errorf("failed to get key metadata: %v", err)
	}

	t := strings.Replace(identity.String(), ":", "", -1)
	aliasName := aliasPrefix + t
	if err := awsKeyProv.createAlias(ctx, aliasName, *keyMetadata.Arn); err != nil {
		log.Error(ctx, "failed to create alias: %v", err)
		return KeyID{}, fmt.Errorf("failed to create alias: %v", err)
	}
	keyID.ID = identity.String()
	return keyID, nil
}

// ListByIdentity returns list of keyIDs for given identity
func (awsKeyProv *awsEthKeyProvider) ListByIdentity(ctx context.Context, identity w3c.DID) ([]KeyID, error) {
	t := strings.Replace(identity.String(), ":", "", -1)
	aliasName := aliasPrefix + t
	metadata, err := awsKeyProv.getKeyInfoByAlias(ctx, aliasName)
	if err != nil {
		log.Info(ctx, "eth key not found in awsKeyProv kms", "err", err)
	}

	if metadata == nil {
		return []KeyID{}, nil
	}

	keyID := KeyID{
		Type: KeyTypeEthereum,
		ID:   *metadata.Arn,
	}
	return []KeyID{keyID}, nil
}

// createAlias creates alias for key
func (awsKeyProv *awsEthKeyProvider) createAlias(ctx context.Context, aliasName, targetKeyId string) error {
	input := &kms.CreateAliasInput{
		AliasName:   aws.String(aliasName),
		TargetKeyId: aws.String(targetKeyId),
	}
	_, err := awsKeyProv.kmsClient.CreateAlias(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create alias: %v", err)
	}

	log.Info(ctx, "alias created:", "aliasName:", aliasName)
	return nil
}

// getKeyInfoByAlias returns key metadata by alias
func (awsKeyProv *awsEthKeyProvider) getKeyInfoByAlias(ctx context.Context, aliasName string) (*types.KeyMetadata, error) {
	aliasInput := &kms.DescribeKeyInput{
		KeyId: aws.String(aliasName),
	}
	aliasOutput, err := awsKeyProv.kmsClient.DescribeKey(ctx, aliasInput)
	if err != nil {
		log.Error(ctx, "failed to describe key: %v", err)
		return nil, fmt.Errorf("failed to describe key: %v", err)
	}
	return aliasOutput.KeyMetadata, nil
}
