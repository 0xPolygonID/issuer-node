package kms

import (
	"context"
	"crypto/rand"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_New_LocalEd25519Provider(t *testing.T) {
	ctx := context.Background()
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())
	ls := NewFileStorageManager(tmpFile.Name())

	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
		URL:       "http://localhost:4566",
	})
	require.NoError(t, err)
	t.Run("should generate a new keyID using local storage manager", func(t *testing.T) {
		localEd25519KeyProvider := NewLocalEd25519KeyProvider(KeyTypeEd25519, ls)
		keyID, err := localEd25519KeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		keyIDParts := strings.Split(keyID.ID, ":")
		assert.Equal(t, 2, len(keyIDParts))
		assert.Equal(t, string(KeyTypeEd25519), keyIDParts[0])
	})

	t.Run("should generate a new keyID using local storage manager with identity", func(t *testing.T) {
		localEd25519KeyProvider := NewLocalEd25519KeyProvider(KeyTypeEd25519, ls)
		did := randomDID(t)
		keyID, err := localEd25519KeyProvider.New(&did)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		keyIDParts := strings.Split(keyID.ID, "/")
		assert.Equal(t, 2, len(keyIDParts))
		assert.Equal(t, did.String(), keyIDParts[0])
		keyIDParts = strings.Split(keyIDParts[1], ":")
		assert.Equal(t, string(KeyTypeEd25519), keyIDParts[0])
	})

	t.Run("should generate a new keyID using aws secret manager", func(t *testing.T) {
		localEd25519KeyProvider := NewLocalEd25519KeyProvider(KeyTypeEd25519, awsStorageProvider)
		keyID, err := localEd25519KeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		keyIDParts := strings.Split(keyID.ID, ":")
		assert.Equal(t, 2, len(keyIDParts))
		assert.Equal(t, string(KeyTypeEd25519), keyIDParts[0])
	})

	t.Run("should generate a new keyID using aws secret manager with identity", func(t *testing.T) {
		localEd25519KeyProvider := NewLocalEd25519KeyProvider(KeyTypeEd25519, awsStorageProvider)
		did := randomDID(t)
		keyID, err := localEd25519KeyProvider.New(&did)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)
		keyIDParts := strings.Split(keyID.ID, "/")
		assert.Equal(t, 2, len(keyIDParts))
		assert.Equal(t, did.String(), keyIDParts[0])
		keyIDParts = strings.Split(keyIDParts[1], ":")
		assert.Equal(t, string(KeyTypeEd25519), keyIDParts[0])
	})
}

func Test_LinkToIdentity_LocalEd25519KeyProvider(t *testing.T) {
	ctx := context.Background()
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())
	ls := NewFileStorageManager(tmpFile.Name())

	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
		URL:       "http://localhost:4566",
	})
	require.NoError(t, err)
	t.Run("should link key to identity using local storage manager", func(t *testing.T) {
		localEd25519KeyProvider := NewLocalEd25519KeyProvider(KeyTypeEd25519, ls)
		keyID, err := localEd25519KeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		did := randomDID(t)
		newKeyID, err := localEd25519KeyProvider.LinkToIdentity(ctx, keyID, did)
		assert.NoError(t, err)
		assert.NotNil(t, keyID)
		assert.Equal(t, did.String()+"/"+keyID.ID, newKeyID.ID)
		assert.Equal(t, KeyTypeEd25519, keyID.Type)
	})

	t.Run("should link key to identity using aws storage manager", func(t *testing.T) {
		localEd25519KeyProvider := NewLocalEd25519KeyProvider(KeyTypeEd25519, awsStorageProvider)
		keyID, err := localEd25519KeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		did := randomDID(t)
		newKeyID, err := localEd25519KeyProvider.LinkToIdentity(ctx, keyID, did)
		assert.NoError(t, err)
		assert.NotNil(t, keyID)
		assert.Equal(t, did.String()+"/"+keyID.ID, newKeyID.ID)
		assert.Equal(t, KeyTypeEd25519, keyID.Type)
	})
}

func Test_ListByIdentity_LocalEd25519KeyProvider(t *testing.T) {
	ctx := context.Background()
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())
	ls := NewFileStorageManager(tmpFile.Name())

	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
		URL:       "http://localhost:4566",
	})
	require.NoError(t, err)
	t.Run("should list keys by identity using local storage manager", func(t *testing.T) {
		localEd25519KeyProvider := NewLocalEd25519KeyProvider(KeyTypeEd25519, ls)
		keyID1, err := localEd25519KeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID1.ID)
		pbkey := strings.Split(keyID1.ID, ":")
		require.Len(t, pbkey, 2)

		did := randomDID(t)
		keyID, err := localEd25519KeyProvider.LinkToIdentity(ctx, keyID1, did)
		assert.NoError(t, err)
		assert.NotNil(t, keyID)

		keyIDs, err := localEd25519KeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keyIDs, 1)
		assert.Equal(t, KeyID{Type: KeyTypeEd25519, ID: did.String() + "/Ed25519:" + pbkey[1]}, keyIDs[0])
	})

	t.Run("should list keys by identity using aws storage manager", func(t *testing.T) {
		localEd25519KeyProvider := NewLocalEd25519KeyProvider(KeyTypeEd25519, awsStorageProvider)
		did := randomDID(t)
		keyID1, err := localEd25519KeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID1.ID)
		pbkey := strings.Split(keyID1.ID, ":")
		require.Len(t, pbkey, 2)

		keyID, err := localEd25519KeyProvider.LinkToIdentity(ctx, keyID1, did)
		assert.NoError(t, err)
		assert.NotNil(t, keyID)

		keyIDs, err := localEd25519KeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keyIDs, 1)
		assert.Equal(t, KeyID{Type: KeyTypeEd25519, ID: did.String() + "/Ed25519:" + pbkey[1]}, keyIDs[0])
	})
}

func Test_PublicKey_LocalEd25519KeyProvider(t *testing.T) {
	ctx := context.Background()
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())
	ls := NewFileStorageManager(tmpFile.Name())

	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
		URL:       "http://localhost:4566",
	})
	require.NoError(t, err)

	t.Run("should get public key using local storage manager", func(t *testing.T) {
		localEd25519KeyProvider := NewLocalEd25519KeyProvider(KeyTypeEd25519, ls)
		keyID1, err := localEd25519KeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID1.ID)

		did := randomDID(t)
		keyID, err := localEd25519KeyProvider.LinkToIdentity(ctx, keyID1, did)
		assert.NoError(t, err)
		assert.NotNil(t, keyID)

		keyIDs, err := localEd25519KeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keyIDs, 1)

		publicKey, err := localEd25519KeyProvider.PublicKey(keyIDs[0])
		assert.NoError(t, err)
		assert.NotNil(t, publicKey)
	})

	t.Run("should get public key using aws storage manager", func(t *testing.T) {
		localEd25519KeyProvider := NewLocalEd25519KeyProvider(KeyTypeEd25519, awsStorageProvider)
		keyID1, err := localEd25519KeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID1.ID)

		did := randomDID(t)
		keyID, err := localEd25519KeyProvider.LinkToIdentity(ctx, keyID1, did)
		assert.NoError(t, err)
		assert.NotNil(t, keyID)

		keyIDs, err := localEd25519KeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keyIDs, 1)

		publicKey, err := localEd25519KeyProvider.PublicKey(keyIDs[0])
		assert.NoError(t, err)
		assert.NotNil(t, publicKey)
	})
}

func Test_Sign_LocalEd25519KeyProvider(t *testing.T) {
	ctx := context.Background()
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())
	ls := NewFileStorageManager(tmpFile.Name())

	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
		URL:       "http://localhost:4566",
	})
	require.NoError(t, err)

	data := make([]byte, 32)
	_, err = io.ReadFull(rand.Reader, data)
	assert.NoError(t, err)

	t.Run("should sign digest using local storage manager | linking did", func(t *testing.T) {
		localEd25519KeyProvider := NewLocalEd25519KeyProvider(KeyTypeEd25519, ls)
		keyID, err := localEd25519KeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		did := randomDID(t)
		keyID, err = localEd25519KeyProvider.LinkToIdentity(ctx, keyID, did)
		assert.NoError(t, err)

		keys, err := localEd25519KeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keys, 1)

		signature, err := localEd25519KeyProvider.Sign(ctx, keys[0], data)
		assert.NoError(t, err)
		assert.NotNil(t, signature)
	})

	t.Run("should sign digest using aws storage manager | linking did", func(t *testing.T) {
		localEd25519KeyProvider := NewLocalEd25519KeyProvider(KeyTypeEd25519, awsStorageProvider)
		keyID, err := localEd25519KeyProvider.New(nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, keyID.ID)

		did := randomDID(t)
		keyID, err = localEd25519KeyProvider.LinkToIdentity(ctx, keyID, did)
		assert.NoError(t, err)

		keys, err := localEd25519KeyProvider.ListByIdentity(ctx, did)
		assert.NoError(t, err)
		assert.Len(t, keys, 1)

		signature, err := localEd25519KeyProvider.Sign(ctx, keys[0], data)
		assert.NoError(t, err)
		assert.NotNil(t, signature)
	})
}
