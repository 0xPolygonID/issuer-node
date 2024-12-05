package kms

import (
	"context"
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

const (
	aliasPrefix       = "alias/"
	awsKmdKeyIDPrefix = "ETH/"
	awsKmsKeyIDParts  = 2
)

type awsKmsEthKeyProvider struct {
	keyType                  KeyType
	reIdenKeyPathHex         *regexp.Regexp // RE of key path bounded to identity
	kmsClient                *kms.Client
	issuerETHTransferKeyPath string
}

// AwKmsEthKeyProviderConfig - configuration for AWS KMS Ethereum key provider
type AwKmsEthKeyProviderConfig struct {
	AccessKey string
	SecretKey string
	Region    string
	URL       string
}

// NewAwsKMSEthKeyProvider - creates new key provider for Ethereum keys stored in AWS KMS
func NewAwsKMSEthKeyProvider(ctx context.Context, keyType KeyType, issuerETHTransferKeyPath string, awsKmsEthKeyProviderConfig AwKmsEthKeyProviderConfig) (KeyProvider, error) {
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

	var options []func(*kms.Options)
	if strings.ToLower(awsKmsEthKeyProviderConfig.Region) == "local" {
		options = make([]func(*kms.Options), 1)
		options[0] = func(o *kms.Options) {
			o.BaseEndpoint = aws.String(awsKmsEthKeyProviderConfig.URL)
		}
	}

	svc := kms.NewFromConfig(cfg, options...)
	return &awsKmsEthKeyProvider{
		keyType:                  keyType,
		reIdenKeyPathHex:         reIdenKeyPathHex,
		kmsClient:                svc,
		issuerETHTransferKeyPath: issuerETHTransferKeyPath,
	}, nil
}

func (awsKeyProv *awsKmsEthKeyProvider) New(identity *w3c.DID) (KeyID, error) {
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
		log.Error(ctx, "failed to create key", "err", err)
		return KeyID{}, fmt.Errorf("failed to create key: %v", err)
	}
	log.Info(ctx, "keyArn", "keyArn", keyArn.KeyMetadata.Arn)

	//inputPublicKey := &kms.GetPublicKeyInput{
	//	KeyId: aws.String(*keyArn.KeyMetadata.Arn),
	//}

	//publicKeyResult, err := awsKeyProv.kmsClient.GetPublicKey(ctx, inputPublicKey)
	//if err != nil {
	//	return KeyID{}, fmt.Errorf("failed to get public key: %v", err)
	//}
	//pubKeyHex := hex.EncodeToString(publicKeyResult.PublicKey)
	//keyID.ID = keyPathForAws(identity, awsKeyProv.keyType, pubKeyHex)
	//base64ID := base64.StdEncoding.EncodeToString([]byte(keyID.ID))
	//
	//aliasName := aliasPrefix + base64ID
	//err = awsKeyProv.createAlias(ctx, aliasName, *keyArn.KeyMetadata.KeyId)
	//
	//if err != nil {
	//	log.Error(ctx, "failed to create alias: %v", err)
	//	return KeyID{}, fmt.Errorf("failed to create alias: %v", err)
	//}
	keyID.ID = awsKmdKeyIDPrefix + *keyArn.KeyMetadata.KeyId
	return keyID, nil
}

// PublicKey returns public key for given keyID
func (awsKeyProv *awsKmsEthKeyProvider) PublicKey(keyID KeyID) ([]byte, error) {
	ctx := context.Background()
	if keyID.ID == awsKeyProv.issuerETHTransferKeyPath {
		keyID.ID = aliasPrefix + awsKeyProv.issuerETHTransferKeyPath
	} else {
		keyIDParts := strings.Split(keyID.ID, "ETH/")
		if len(keyIDParts) == awsKmsKeyIDParts {
			keyID.ID = keyIDParts[1]
		}
	}

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
func (awsKeyProv *awsKmsEthKeyProvider) Sign(ctx context.Context, keyID KeyID, data []byte) ([]byte, error) {
	if keyID.ID == awsKeyProv.issuerETHTransferKeyPath {
		keyID.ID = aliasPrefix + awsKeyProv.issuerETHTransferKeyPath
	} else {
		keyIDParts := strings.Split(keyID.ID, awsKmdKeyIDPrefix)
		if len(keyIDParts) != awsKmsKeyIDParts {
			return nil, fmt.Errorf("invalid keyID: %v", keyID.ID)
		}
		keyID.ID = keyIDParts[1]
	}

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
func (awsKeyProv *awsKmsEthKeyProvider) LinkToIdentity(ctx context.Context, keyID KeyID, identity w3c.DID) (KeyID, error) {
	//base64ID := base64.StdEncoding.EncodeToString([]byte(keyID.ID))
	//alias := aliasPrefix + base64ID
	//keyMetadata, err := awsKeyProv.getKeyInfoByAlias(ctx, alias)
	//if err != nil {
	//	log.Error(ctx, "failed to get key metadata", "keyMetadata", keyMetadata, "err", err)
	//	return KeyID{}, fmt.Errorf("failed to get key metadata: %v", err)
	//}

	keyIDParts := strings.Split(keyID.ID, awsKmdKeyIDPrefix)
	if len(keyIDParts) != awsKmsKeyIDParts {
		return KeyID{}, fmt.Errorf("invalid keyID: %v", keyID.ID)
	}

	tagResourceInput := &kms.TagResourceInput{
		KeyId: aws.String(keyIDParts[1]),
		Tags: []types.Tag{
			{
				TagKey:   aws.String("keyType"),
				TagValue: aws.String(string(keyID.Type)),
			},
			{
				TagKey:   aws.String("did"),
				TagValue: aws.String(identity.String()),
			},
		},
	}

	resourceOutput, err := awsKeyProv.kmsClient.TagResource(ctx, tagResourceInput)
	if err != nil {
		log.Error(ctx, "failed to tag resource", "err", err)
		return KeyID{}, fmt.Errorf("failed to tag resource: %v", err)
	}

	log.Info(ctx, "resource tagged:", "resourceOutput:", resourceOutput.ResultMetadata)
	return keyID, nil
}

// ListByIdentity returns list of keyIDs for given identity
func (awsKeyProv *awsKmsEthKeyProvider) ListByIdentity(ctx context.Context, identity w3c.DID) ([]KeyID, error) {
	const limit = 500
	listKeysInput := &kms.ListKeysInput{
		Limit: aws.Int32(limit),
	}
	listKeysOutput, err := awsKeyProv.kmsClient.ListKeys(ctx, listKeysInput)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	keysToReturn := make([]KeyID, 0)
	for _, key := range listKeysOutput.Keys {
		describeInput := &kms.ListResourceTagsInput{
			KeyId: key.KeyId,
		}
		tagOutput, err := awsKeyProv.kmsClient.ListResourceTags(ctx, describeInput)
		if err != nil {
			log.Error(ctx, "failed to list tags", "keyID", aws.ToString(key.KeyId), "err", err)
			continue
		}

		for _, tag := range tagOutput.Tags {
			if aws.ToString(tag.TagKey) == "did" && aws.ToString(tag.TagValue) == identity.String() {
				id := "ETH/" + *key.KeyId
				keysToReturn = append(keysToReturn, KeyID{
					Type: KeyTypeEthereum,
					ID:   aws.ToString(&id),
				})
			}
		}
	}

	return keysToReturn, nil
}

func (awsKeyProv *awsKmsEthKeyProvider) Delete(ctx context.Context, keyID KeyID) error {
	const pendingWindowInDays = 7
	keyIDParts := strings.Split(keyID.ID, awsKmdKeyIDPrefix)
	if len(keyIDParts) != awsKmsKeyIDParts {
		return fmt.Errorf("invalid keyID: %v", keyID.ID)
	}
	_, err := awsKeyProv.kmsClient.ScheduleKeyDeletion(ctx, &kms.ScheduleKeyDeletionInput{
		KeyId:               aws.String(keyIDParts[1]),
		PendingWindowInDays: aws.Int32(pendingWindowInDays),
	})
	return err
}

func (awsKeyProv *awsKmsEthKeyProvider) Exists(ctx context.Context, keyID KeyID) (bool, error) {
	keyIDParts := strings.Split(keyID.ID, awsKmdKeyIDPrefix)
	if len(keyIDParts) != awsKmsKeyIDParts {
		return false, fmt.Errorf("invalid keyID: %v", keyID.ID)
	}
	_, err := awsKeyProv.getKeyInfoByAlias(ctx, keyIDParts[1])
	if err != nil {
		return false, nil
	}
	return true, nil
}

// getKeyInfoByAlias returns key metadata by alias
func (awsKeyProv *awsKmsEthKeyProvider) getKeyInfoByAlias(ctx context.Context, aliasName string) (*types.KeyMetadata, error) {
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
