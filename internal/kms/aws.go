package kms

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/polygonid/sh-id-platform/internal/log"
)

func LoadAWSConfig(ctx context.Context) (aws.Config, error) {
	// Backward-compatible behaviour for AWS SDK configuration
	// env variables (DEPRECATED)
	// "ISSUER_KMS_AWS_ACCESS_KEY"
	// "ISSUER_KMS_AWS_SECRET_KEY"
	// "ISSUER_KMS_AWS_REGION"
	accessKey := strings.TrimSpace(os.Getenv("ISSUER_KMS_AWS_ACCESS_KEY"))
	secretKey := strings.TrimSpace(os.Getenv("ISSUER_KMS_AWS_SECRET_KEY"))
	region := strings.TrimSpace(os.Getenv("ISSUER_KMS_AWS_REGION"))

	if accessKey != "" && secretKey != "" {
		log.Info(ctx, "Loading AWS config with static credentials")

		options := []func(*config.LoadOptions) error{
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
			),
		}

		if region != "" {
			options = append(options, config.WithRegion(region))
		}

		return config.LoadDefaultConfig(ctx, options...)
	}

	log.Info(ctx, "Loading AWS config with default credentials")

	return config.LoadDefaultConfig(ctx)
}

func AwsSecretsManager(ctx context.Context) (*secretsmanager.Client, error) {
	cfg, err := LoadAWSConfig(ctx)

	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}

	// LocalStack/OpenStack mode
	// https://docs.localstack.cloud/aws/integrations/aws-sdks/go/
	// Region is provided from AWS_REGION env variable
	url := strings.TrimSpace(os.Getenv("ISSUER_KMS_AWS_URL"))
	var options []func(*secretsmanager.Options)
	if url != "" {
		options = append(options, func(o *secretsmanager.Options) {
			o.BaseEndpoint = aws.String(url)
		})
	}
	return secretsmanager.NewFromConfig(cfg, options...), nil
}

func AwsKms(ctx context.Context) (*kms.Client, error) {
	cfg, err := LoadAWSConfig(ctx)

	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}

	var options []func(*kms.Options)

	// LocalStack/OpenStack mode
	// https://docs.localstack.cloud/aws/integrations/aws-sdks/go/
	// Region is provided from AWS_REGION env variable
	url := strings.TrimSpace(os.Getenv("ISSUER_KMS_AWS_URL"))
	if url != "" {
		options = append(options, func(o *kms.Options) {
			o.BaseEndpoint = aws.String(url)
		})
	}

	return kms.NewFromConfig(cfg, options...), nil
}
