package kms

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveKeyMaterialToFile_Success(t *testing.T) {
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())

	ls := NewLocalStorageFileManager(tmpFile.Name())
	ctx := context.Background()
	keyMaterial := map[string]string{jsonKeyType: string(KeyTypeEthereum), jsonKeyData: "0xABC123"}
	id := "key1"

	err = ls.saveKeyMaterialToFile(ctx, keyMaterial, id)
	require.NoError(t, err)

	content, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	var fileContent []localStorageBJJKeyProviderFileContent
	err = json.Unmarshal(content, &fileContent)
	require.NoError(t, err)

	assert.Equal(t, 1, len(fileContent))
	assert.Equal(t, id, fileContent[0].KeyPath)
	assert.Equal(t, ethereum, fileContent[0].KeyType)
	assert.Equal(t, keyMaterial[jsonKeyData], fileContent[0].PrivateKey)
}

func TestSaveKeyMaterialToFile_FailOnFileWrite(t *testing.T) {
	ls := NewLocalStorageFileManager("/path/to/non/existent/file")
	ctx := context.Background()
	keyMaterial := map[string]string{"type": "Ethereum", "data": "0xABC123"}
	id := "key1"

	err := ls.saveKeyMaterialToFile(ctx, keyMaterial, id)
	assert.Error(t, err)
}

func TestSearchByIdentityInFile_ReturnsKeyIDsOnMatch(t *testing.T) {
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())

	identity := "did:polygonid:polygon:amoy:2qQ68JkRcf3ybQNvgRV9BP6qLgBrXmUezqBi4wsEuV"
	fileContent := []localStorageBJJKeyProviderFileContent{
		{KeyPath: identity + "/ETH:0347fe70a2a9b752e8012d72851c35a13a1423bcdac4bde6ec036e1ea9317b36ac", KeyType: ethereum, PrivateKey: "0xABC123"},
		{KeyPath: "keys/" + identity + "/BJJ:cecf34ed27074e121f1e8a8cc75954ab2b28506258b87b3c9a20e33461f4b12a", KeyType: babyjubjub, PrivateKey: "0xDEF456"},
	}

	content, err := json.Marshal(fileContent)
	require.NoError(t, err)
	//nolint:all
	err = os.WriteFile("./kms.json", content, 0644)
	require.NoError(t, err)

	ls := NewLocalStorageFileManager(tmpFile.Name())
	ctx := context.Background()
	did, err := w3c.ParseDID(identity)
	require.NoError(t, err)

	keyIDs, err := ls.searchByIdentityInFile(ctx, *did, KeyTypeEthereum)
	require.NoError(t, err)
	require.Len(t, keyIDs, 1)
	assert.Equal(t, KeyID{Type: KeyTypeEthereum, ID: identity + "/ETH:0347fe70a2a9b752e8012d72851c35a13a1423bcdac4bde6ec036e1ea9317b36ac"}, keyIDs[0])

	keyIDs, err = ls.searchByIdentityInFile(ctx, *did, KeyTypeBabyJubJub)
	require.NoError(t, err)
	require.Len(t, keyIDs, 1)
	assert.Equal(t, KeyID{Type: KeyTypeBabyJubJub, ID: "keys/" + identity + "/BJJ:cecf34ed27074e121f1e8a8cc75954ab2b28506258b87b3c9a20e33461f4b12a"}, keyIDs[0])
}

//nolint:lll
func TestSearchByIdentityInFile_ReturnsErrorOnFileReadFailure(t *testing.T) {
	ls := NewLocalStorageFileManager("/path/to/nonexistent/file")
	ctx := context.Background()
	did, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qQ68JkRcf3ybQNvgRV9BP6qLgBrXmUezqBi4wsEuV")
	require.NoError(t, err)
	_, err = ls.searchByIdentityInFile(ctx, *did, KeyTypeEthereum)
	assert.Error(t, err)
}

func TestSearchByIdentityInFile_ReturnsEmptySliceWhenNoMatch(t *testing.T) {
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())

	fileContent := []localStorageBJJKeyProviderFileContent{
		{KeyPath: "key/did:example:456", KeyType: string(KeyTypeEthereum), PrivateKey: "0xABC123"},
	}
	content, err := json.Marshal(fileContent)
	require.NoError(t, err)
	//nolint:all
	err = os.WriteFile("./kms.json", content, 0644)
	require.NoError(t, err)

	ls := NewLocalStorageFileManager(tmpFile.Name())
	ctx := context.Background()

	did, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qQ68JkRcf3ybQNvgRV9BP6qLgBrXmUezqBi4wsEuV")
	require.NoError(t, err)

	keyIDs, err := ls.searchByIdentityInFile(ctx, *did, KeyTypeEthereum)
	require.NoError(t, err)
	assert.Empty(t, keyIDs)
}

//nolint:lll
func TestSearchPrivateKeyInFile_ReturnsPrivateKeyOnMatch(t *testing.T) {
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())

	fileContent := []localStorageBJJKeyProviderFileContent{
		{KeyPath: "key1", KeyType: "ETH", PrivateKey: "0xABC123"},
	}
	content, err := json.Marshal(fileContent)
	require.NoError(t, err)
	//nolint:all
	err = os.WriteFile("./kms.json", content, 0644)
	require.NoError(t, err)

	ls := NewLocalStorageFileManager(tmpFile.Name())
	ctx := context.Background()

	privateKey, err := ls.searchPrivateKeyInFile(ctx, KeyID{ID: "key1"})
	require.NoError(t, err)
	assert.Equal(t, "0xABC123", privateKey)
}

//nolint:lll
func TestSearchPrivateKeyInFile_ReturnsErrorWhenKeyNotFound(t *testing.T) {
	tmpFile, err := createTestFile(t)
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.Remove(tmpFile.Name())

	fileContent := []localStorageBJJKeyProviderFileContent{
		{KeyPath: "key1", KeyType: "Ethereum", PrivateKey: "0xABC123"},
	}
	content, err := json.Marshal(fileContent)
	require.NoError(t, err)
	//nolint:all
	err = os.WriteFile("./kms.json", content, 0644)
	require.NoError(t, err)

	ls := NewLocalStorageFileManager(tmpFile.Name())
	ctx := context.Background()

	_, err = ls.searchPrivateKeyInFile(ctx, KeyID{ID: "key2"})
	assert.Error(t, err)
}

func createTestFile(t *testing.T) (*os.File, error) {
	t.Helper()
	tmpFile, err := os.Create("./kms.json")
	assert.NoError(t, err)
	initFileContent := []byte("[]")
	_, err = tmpFile.Write(initFileContent)
	assert.NoError(t, err)
	require.NoError(t, tmpFile.Close())
	return tmpFile, err
}
