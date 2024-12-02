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

	t.Run("should get an error", func(t *testing.T) {
		publicKey, err := awsStorageProvider.PublicKey(KeyID{})
		assert.Error(t, err)
		assert.Empty(t, publicKey)
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
		assert.Equal(t, identity.String(), keyID.ID)
	})

	t.Run("should get an error", func(t *testing.T) {
		identity := randomDID(t)
		keyID, err := awsStorageProvider.LinkToIdentity(ctx, KeyID{}, identity)
		assert.Error(t, err)
		assert.Empty(t, keyID)
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

	t.Run("should link twio keys to an identity", func(t *testing.T) {
		keyID1, err := awsStorageProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID1.ID)
		assert.Equal(t, ethereum, string(keyID1.Type))

		keyID2, err := awsStorageProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID2.ID)
		assert.Equal(t, ethereum, string(keyID2.Type))

		identity := randomDID(t)

		keyID1, err = awsStorageProvider.LinkToIdentity(ctx, keyID1, identity)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID1.ID)

		keyID2, err = awsStorageProvider.LinkToIdentity(ctx, keyID2, identity)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID2.ID)

		keyIDs, err := awsStorageProvider.ListByIdentity(ctx, identity)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyIDs)
		assert.Len(t, keyIDs, 2)
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

	t.Run("should get an error", func(t *testing.T) {
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

		signature, err := awsStorageProvider.Sign(ctx, KeyID{}, data)
		assert.Error(t, err)
		assert.Empty(t, signature)
	})
}
