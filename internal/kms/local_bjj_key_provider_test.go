package kms

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_New_LocalBJJKeyProvider(t *testing.T) {
	ctx := context.Background()
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())
	ls := NewLocalStorageFileProvider(tmpFile.Name())

	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
	})
	require.NoError(t, err)
	t.Run("should generate a new keyID using local storage manager", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, ls)
		keyID, err := localbbjKeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		keyIDParts := strings.Split(keyID.ID, ":")
		assert.Equal(t, 2, len(keyIDParts))
		assert.Equal(t, string(KeyTypeBabyJubJub), keyIDParts[0])
	})

	t.Run("should generate a new keyID using local storage manager with identity", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, ls)
		did := randomDID(t)
		keyID, err := localbbjKeyProvider.New(&did)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		keyIDParts := strings.Split(keyID.ID, "/")
		assert.Equal(t, 2, len(keyIDParts))
		assert.Equal(t, did.String(), keyIDParts[0])
		keyIDParts = strings.Split(keyIDParts[1], ":")
		assert.Equal(t, string(KeyTypeBabyJubJub), keyIDParts[0])
	})

	t.Run("should generate a new keyID using aws secret manager", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, awsStorageProvider)
		keyID, err := localbbjKeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		keyIDParts := strings.Split(keyID.ID, ":")
		assert.Equal(t, 2, len(keyIDParts))
		assert.Equal(t, string(KeyTypeBabyJubJub), keyIDParts[0])
	})

	t.Run("should generate a new keyID using aws secret manager with identity", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, awsStorageProvider)
		did := randomDID(t)
		keyID, err := localbbjKeyProvider.New(&did)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		keyIDParts := strings.Split(keyID.ID, "/")
		assert.Equal(t, 2, len(keyIDParts))
		assert.Equal(t, did.String(), keyIDParts[0])
		keyIDParts = strings.Split(keyIDParts[1], ":")
		assert.Equal(t, string(KeyTypeBabyJubJub), keyIDParts[0])
	})
}

func Test_LinkToIdentity_LocalBJJKeyProvider(t *testing.T) {
	ctx := context.Background()
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())
	ls := NewLocalStorageFileProvider(tmpFile.Name())

	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
	})
	require.NoError(t, err)
	t.Run("should link key to identity using local storage manager", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, ls)
		keyID, err := localbbjKeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		did := randomDID(t)
		keyID, err = localbbjKeyProvider.LinkToIdentity(ctx, keyID, did)
		assert.NoError(t, err)
		assert.NotNil(t, keyID)
		assert.Equal(t, did.String(), keyID.ID)
		assert.Equal(t, KeyTypeBabyJubJub, keyID.Type)
	})

	t.Run("should link key to identity using aws storage manager", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, awsStorageProvider)
		keyID, err := localbbjKeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		did := randomDID(t)
		keyID, err = localbbjKeyProvider.LinkToIdentity(ctx, keyID, did)
		assert.NoError(t, err)
		assert.NotNil(t, keyID)
		assert.Equal(t, did.String(), keyID.ID)
		assert.Equal(t, KeyTypeBabyJubJub, keyID.Type)
	})
}

func Test_ListByIdentity_LocalBJJKeyProvider(t *testing.T) {
	ctx := context.Background()
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())
	ls := NewLocalStorageFileProvider(tmpFile.Name())

	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
	})
	require.NoError(t, err)
	t.Run("should list keys by identity using local storage manager", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, ls)
		did := randomDID(t)
		keyID, err := localbbjKeyProvider.New(&did)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		keyIDs, err := localbbjKeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keyIDs, 1)
		assert.Equal(t, keyID, keyIDs[0])
	})

	t.Run("should list keys by identity using aws storage manager", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, awsStorageProvider)
		did := randomDID(t)
		keyID, err := localbbjKeyProvider.New(&did)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		keyIDs, err := localbbjKeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keyIDs, 1)
		assert.Equal(t, keyID, keyIDs[0])
	})

	t.Run("should list keys by identity using local storage manager and linking identity", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, ls)
		did := randomDID(t)
		keyID1, err := localbbjKeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID1.ID)

		keyID, err := localbbjKeyProvider.LinkToIdentity(ctx, keyID1, did)
		assert.NoError(t, err)

		keyIDs, err := localbbjKeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keyIDs, 1)
		assert.Equal(t, KeyTypeBabyJubJub, keyIDs[0].Type)
		assert.Equal(t, keyID.ID+"/"+keyID1.ID, keyIDs[0].ID)
	})

	t.Run("should list keys by identity using aws storage manager and linking identity", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, awsStorageProvider)
		did := randomDID(t)
		keyID1, err := localbbjKeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID1.ID)

		keyID, err := localbbjKeyProvider.LinkToIdentity(ctx, keyID1, did)
		assert.NoError(t, err)

		keyIDs, err := localbbjKeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keyIDs, 1)
		assert.Equal(t, KeyTypeBabyJubJub, keyIDs[0].Type)
		assert.Equal(t, keyID.ID+"/"+keyID1.ID, keyIDs[0].ID)
	})
}

func Test_PublicKey_LocalBJJKeyProvider(t *testing.T) {
	ctx := context.Background()
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())
	ls := NewLocalStorageFileProvider(tmpFile.Name())

	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
	})
	require.NoError(t, err)

	t.Run("should get public key using local storage manager", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, ls)
		keyID, err := localbbjKeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		publicKey, err := localbbjKeyProvider.PublicKey(keyID)
		assert.NoError(t, err)
		assert.NotNil(t, publicKey)
	})

	t.Run("should get public key using aws storage manager", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, awsStorageProvider)
		keyID, err := localbbjKeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		publicKey, err := localbbjKeyProvider.PublicKey(keyID)
		assert.NoError(t, err)
		assert.NotNil(t, publicKey)
	})
}

func Test_Sign_LocalBJJKeyProvider(t *testing.T) {
	ctx := context.Background()
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())
	ls := NewLocalStorageFileProvider(tmpFile.Name())

	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
	})
	require.NoError(t, err)

	t.Run("should sign digest using local storage manager | linking did", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, ls)
		keyID, err := localbbjKeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		did := randomDID(t)
		keyID, err = localbbjKeyProvider.LinkToIdentity(ctx, keyID, did)
		assert.NoError(t, err)

		keys, err := localbbjKeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keys, 1)

		data := []byte("data")
		signature, err := localbbjKeyProvider.Sign(ctx, keys[0], data)
		assert.NoError(t, err)
		assert.NotNil(t, signature)
	})

	t.Run("should sign digest using aws storage manager | linking did", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, awsStorageProvider)
		keyID, err := localbbjKeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		did := randomDID(t)
		keyID, err = localbbjKeyProvider.LinkToIdentity(ctx, keyID, did)
		assert.NoError(t, err)

		keys, err := localbbjKeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keys, 1)

		data := []byte("data")
		signature, err := localbbjKeyProvider.Sign(ctx, keys[0], data)
		assert.NoError(t, err)
		assert.NotNil(t, signature)
	})

	t.Run("should sign digest using local storage manager", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, ls)
		did := randomDID(t)
		keyID, err := localbbjKeyProvider.New(&did)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		keys, err := localbbjKeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keys, 1)

		data := []byte("data")
		signature, err := localbbjKeyProvider.Sign(ctx, keys[0], data)
		assert.NoError(t, err)
		assert.NotNil(t, signature)
	})

	t.Run("should sign digest using aws storage manager", func(t *testing.T) {
		localbbjKeyProvider := NewLocalBJJKeyProvider(KeyTypeBabyJubJub, awsStorageProvider)
		did := randomDID(t)
		keyID, err := localbbjKeyProvider.New(&did)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		keys, err := localbbjKeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keys, 1)

		data := []byte("data")
		signature, err := localbbjKeyProvider.Sign(ctx, keys[0], data)
		assert.NoError(t, err)
		assert.NotNil(t, signature)
	})
}
