package kms

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SaveKeyMaterial(t *testing.T) {
	ctx := context.Background()
	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
	})
	require.NoError(t, err)

	t.Run("should save key bjj material", func(t *testing.T) {
		did := randomDID(t)
		id := getKeyID(&did, KeyTypeBabyJubJub, "key_data")
		err := awsStorageProvider.SaveKeyMaterial(ctx, map[string]string{
			jsonKeyType: string(KeyTypeBabyJubJub),
			jsonKeyData: "key_data",
		}, id)
		assert.NoError(t, err)
	})

	t.Run("should save key bjj material with empty did", func(t *testing.T) {
		did := randomDID(t)
		keyID := did.String() + "/BJJ:key_data"
		id := getKeyID(nil, KeyTypeBabyJubJub, keyID)
		err := awsStorageProvider.SaveKeyMaterial(ctx, map[string]string{
			jsonKeyType: string(KeyTypeBabyJubJub),
			jsonKeyData: "key_data",
		}, id)
		assert.NoError(t, err)
	})

	t.Run("should save key eth material", func(t *testing.T) {
		did := randomDID(t)
		id := getKeyID(&did, KeyTypeEthereum, "key_data")
		err := awsStorageProvider.SaveKeyMaterial(ctx, map[string]string{
			jsonKeyType: string(KeyTypeEthereum),
			jsonKeyData: "key_data",
		}, id)
		assert.NoError(t, err)
	})

	t.Run("should save key eth material with empty did", func(t *testing.T) {
		did := randomDID(t)
		keyID := did.String() + "/EYH:key_data"
		id := getKeyID(nil, KeyTypeEthereum, keyID)
		err := awsStorageProvider.SaveKeyMaterial(ctx, map[string]string{
			jsonKeyType: string(KeyTypeEthereum),
			jsonKeyData: "key_data",
		}, id)
		assert.NoError(t, err)
	})

	t.Run("should get an error | wrong id", func(t *testing.T) {
		id := getKeyID(nil, KeyTypeEthereum, "ETH")
		err := awsStorageProvider.SaveKeyMaterial(ctx, map[string]string{
			jsonKeyType: string(KeyTypeEthereum),
			jsonKeyData: "key_data",
		}, id)
		assert.Error(t, err)
	})
}

func Test_searchByIdentity(t *testing.T) {
	ctx := context.Background()
	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
	})
	require.NoError(t, err)

	t.Run("should get identity for BJJ", func(t *testing.T) {
		did := randomDID(t)
		id := getKeyID(&did, KeyTypeBabyJubJub, "key_data")
		err := awsStorageProvider.SaveKeyMaterial(ctx, map[string]string{
			jsonKeyType:    string(KeyTypeBabyJubJub),
			jsonKeyData:    "key_data",
			jsonPrivateKey: "private_key",
		}, id)
		assert.NoError(t, err)

		keyIDs, err := awsStorageProvider.searchByIdentity(ctx, did, KeyTypeBabyJubJub)
		require.NoError(t, err)
		require.Len(t, keyIDs, 1)
		keyID := keyIDs[0]
		assert.Equal(t, KeyID{Type: KeyTypeBabyJubJub, ID: did.String() + "/BJJ:key_data"}, keyID)
	})

	t.Run("should get identity for BJJ from ETH", func(t *testing.T) {
		did := randomDID(t)
		id := did.String() + "/BJJ:key_data"
		err := awsStorageProvider.SaveKeyMaterial(ctx, map[string]string{
			jsonKeyType:    string(KeyTypeBabyJubJub),
			jsonKeyData:    "key_data",
			jsonPrivateKey: "private_key",
		}, id)
		assert.NoError(t, err)

		keyIDs, err := awsStorageProvider.searchByIdentity(ctx, did, KeyTypeBabyJubJub)
		require.NoError(t, err)
		require.Len(t, keyIDs, 1)
		keyID := keyIDs[0]
		assert.Equal(t, KeyID{Type: KeyTypeBabyJubJub, ID: did.String() + "/BJJ:key_data"}, keyID)
	})

	t.Run("should get identity for ETH", func(t *testing.T) {
		did := randomDID(t)
		id := getKeyID(&did, KeyTypeEthereum, "key_data")
		err := awsStorageProvider.SaveKeyMaterial(ctx, map[string]string{
			jsonKeyType: string(KeyTypeEthereum),
			jsonKeyData: "key_data",
		}, id)
		assert.NoError(t, err)

		keyIDs, err := awsStorageProvider.searchByIdentity(ctx, did, KeyTypeEthereum)
		require.NoError(t, err)
		require.Len(t, keyIDs, 1)
		keyID := keyIDs[0]
		assert.Equal(t, KeyID{Type: KeyTypeEthereum, ID: did.String() + "/ETH:key_data"}, keyID)
	})
}

func Test_searchPrivateKey(t *testing.T) {
	ctx := context.Background()
	awsStorageProvider, err := NewAwsSecretStorageProvider(ctx, AwsSecretStorageProviderConfig{
		AccessKey: "access_key",
		SecretKey: "secret_key",
		Region:    "local",
	})
	require.NoError(t, err)

	t.Run("should get private key for BJJ", func(t *testing.T) {
		did := randomDID(t)
		privateKey := "9d7abdd5a43573ab9b623c50b9fc8f4357329d3009fe0fc22c8931161d98a03d"
		id := getKeyID(&did, KeyTypeBabyJubJub, "BJJ:2290140c920a31a596937095f18a9ae15c1fe7091091be485f353968a4310380")
		err := awsStorageProvider.SaveKeyMaterial(ctx, map[string]string{
			jsonKeyType: string(KeyTypeBabyJubJub),
			jsonKeyData: privateKey,
		}, id)
		assert.NoError(t, err)

		keyIDs, err := awsStorageProvider.searchByIdentity(ctx, did, KeyTypeBabyJubJub)
		require.NoError(t, err)
		require.Len(t, keyIDs, 1)
		keyID := keyIDs[0]
		assert.Equal(t, KeyID{Type: KeyTypeBabyJubJub, ID: did.String() + "/BJJ:2290140c920a31a596937095f18a9ae15c1fe7091091be485f353968a4310380"}, keyID)

		privateKeyFromStore, err := awsStorageProvider.searchPrivateKey(ctx, keyID)
		require.NoError(t, err)
		assert.Equal(t, privateKey, privateKeyFromStore)
	})

	t.Run("should get private key for ETH", func(t *testing.T) {
		did := randomDID(t)
		privateKey := "9d7abdd5a43573ab9b623c50b9fc8f4357329d3009fe0fc22c8931161d98a03d"
		id := getKeyID(&did, KeyTypeEthereum, "ETH:2290140c920a31a596937095f18a9ae15c1fe7091091be485f353968a4310380")
		err := awsStorageProvider.SaveKeyMaterial(ctx, map[string]string{
			jsonKeyType: string(KeyTypeEthereum),
			jsonKeyData: privateKey,
		}, id)
		assert.NoError(t, err)

		keyIDs, err := awsStorageProvider.searchByIdentity(ctx, did, KeyTypeEthereum)
		require.NoError(t, err)
		require.Len(t, keyIDs, 1)
		keyID := keyIDs[0]
		assert.Equal(t, KeyID{Type: KeyTypeEthereum, ID: did.String() + "/ETH:2290140c920a31a596937095f18a9ae15c1fe7091091be485f353968a4310380"}, keyID)

		privateKeyFromStore, err := awsStorageProvider.searchPrivateKey(ctx, keyID)
		require.NoError(t, err)
		assert.Equal(t, privateKey, privateKeyFromStore)
	})

	t.Run("should get private key for BJJ | from eth identity", func(t *testing.T) {
		did := randomDID(t)
		privateKey := "9d7abdd5a43573ab9b623c50b9fc8f4357329d3009fe0fc22c8931161d98a03d"
		id := did.String() + "/BJJ:f6eb5b16318de6054ccc30047d9ba395c954e78b6f1ba0a8f52a6e46b7f2500f"
		err := awsStorageProvider.SaveKeyMaterial(ctx, map[string]string{
			jsonKeyType: string(KeyTypeBabyJubJub),
			jsonKeyData: privateKey,
		}, id)
		assert.NoError(t, err)

		keyIDs, err := awsStorageProvider.searchByIdentity(ctx, did, KeyTypeBabyJubJub)
		require.NoError(t, err)
		require.Len(t, keyIDs, 1)
		keyID := keyIDs[0]
		assert.Equal(t, KeyID{Type: KeyTypeBabyJubJub, ID: did.String() + "/BJJ:f6eb5b16318de6054ccc30047d9ba395c954e78b6f1ba0a8f52a6e46b7f2500f"}, keyID)

		privateKeyFromStore, err := awsStorageProvider.searchPrivateKey(ctx, keyID)
		require.NoError(t, err)
		assert.Equal(t, privateKey, privateKeyFromStore)
	})
}
