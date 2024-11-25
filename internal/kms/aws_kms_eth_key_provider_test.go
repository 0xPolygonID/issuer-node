package kms

import (
	"context"
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewInAWSKMS(t *testing.T) {
	ctx := context.Background()
	awsStorageProvider, err := NewAwsKMSEthKeyProvider(ctx, ethereum, "pbkey", AwKmsEthKeyProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
		URL:       "http://localhost:4566",
	})
	require.NoError(t, err)

	t.Run("should create a new KEYID", func(t *testing.T) {
		keyID, err := awsStorageProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		assert.Equal(t, ethereum, string(keyID.Type))
	})
}

func Test_PublicKeyInAWSKMS(t *testing.T) {
	ctx := context.Background()
	awsStorageProvider, err := NewAwsKMSEthKeyProvider(ctx, ethereum, "pbkey", AwKmsEthKeyProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
		URL:       "http://localhost:4566",
	})
	require.NoError(t, err)

	t.Run("should get public key", func(t *testing.T) {
		keyID, err := awsStorageProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		assert.Equal(t, ethereum, string(keyID.Type))

		publicKey, err := awsStorageProvider.PublicKey(keyID)
		assert.NoError(t, err)
		assert.NotEmpty(t, publicKey)
	})
}

func Test_LinkToIdentityInAWSKMS(t *testing.T) {
	ctx := context.Background()
	awsStorageProvider, err := NewAwsKMSEthKeyProvider(ctx, ethereum, "pbkey", AwKmsEthKeyProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
		URL:       "http://localhost:4566",
	})
	require.NoError(t, err)

	t.Run("should link the key to an identity", func(t *testing.T) {
		keyID, err := awsStorageProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		assert.Equal(t, ethereum, string(keyID.Type))

		identity := randomDID(t)

		keyID, err = awsStorageProvider.LinkToIdentity(ctx, keyID, identity)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
	})
}

func Test_ListByIdentityInAWSKMS(t *testing.T) {
	ctx := context.Background()
	awsStorageProvider, err := NewAwsKMSEthKeyProvider(ctx, ethereum, "pbkey", AwKmsEthKeyProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
		URL:       "http://localhost:4566",
	})
	require.NoError(t, err)

	t.Run("should link the key to an identity", func(t *testing.T) {
		keyID, err := awsStorageProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		assert.Equal(t, ethereum, string(keyID.Type))

		identity := randomDID(t)

		keyID, err = awsStorageProvider.LinkToIdentity(ctx, keyID, identity)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		keyIDs, err := awsStorageProvider.ListByIdentity(ctx, identity)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyIDs)
		assert.Len(t, keyIDs, 1)
	})
}

func Test_SignInAWSKMS(t *testing.T) {
	ctx := context.Background()
	awsStorageProvider, err := NewAwsKMSEthKeyProvider(ctx, ethereum, "pbkey", AwKmsEthKeyProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
		URL:       "http://localhost:4566",
	})
	require.NoError(t, err)
	data := make([]byte, 32)
	_, err = io.ReadFull(rand.Reader, data)
	assert.NoError(t, err)

	t.Run("should sign a message", func(t *testing.T) {
		keyID, err := awsStorageProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		assert.Equal(t, ethereum, string(keyID.Type))

		identity := randomDID(t)

		keyID, err = awsStorageProvider.LinkToIdentity(ctx, keyID, identity)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		keyIDs, err := awsStorageProvider.ListByIdentity(ctx, identity)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyIDs)
		assert.Len(t, keyIDs, 1)

		signature, err := awsStorageProvider.Sign(ctx, keyIDs[0], data)
		assert.NoError(t, err)
		assert.NotEmpty(t, signature)
	})
}
