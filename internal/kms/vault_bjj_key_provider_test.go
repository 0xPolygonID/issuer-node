package kms

import (
	"context"
	"encoding/hex"
	"math/big"
	"os"
	"strings"
	"testing"

	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/constants"
	"github.com/iden3/go-iden3-crypto/utils"
	"github.com/stretchr/testify/require"
)

func TestVaultBJJKeyProvider_PublicKey(t *testing.T) {
	dec := func(in string) []byte {
		r, err := hex.DecodeString(in)
		require.NoError(t, err)
		return r
	}
	testCases := []struct {
		title   string
		keyID   string
		want    []byte
		wantErr string
	}{
		{
			title: "key in root directory without identity",
			keyID: "BJJ:be0a12be07e1643e226594862871d048d94677f85baa1969683fa2a7e9e02923",
			want:  dec("be0a12be07e1643e226594862871d048d94677f85baa1969683fa2a7e9e02923"),
		},
		{
			title: "key with identity in path",
			keyID: "keys/did:iden3:polygon:mumbai:x2Uw18ATvY7mEsgfrrDipBmQQdPWAao4NmF56wGvp/BJJ:0980061591BCF8851dbf220bd9acba37d609010f4fb76b729e54f18f6bdc9784",
			want:  dec("0980061591bcf8851dbf220bd9acba37d609010f4fb76b729e54f18f6bdc9784"),
		},
		{
			title:   "incorrect path",
			keyID:   "keys/did:iden3:polygon:mumbai:x2Uw18ATvY7mEsgfrrDipBmQQdPWAao4NmF56wGvp/BJ0980061591BCF8851dbf220bd9acba37d609010f4fb76b729e54f18f6bdc9784",
			wantErr: "unable to get public key from key ID",
		},
	}
	keyProvider := NewVaultBJJKeyProvider(nil, KeyTypeBabyJubJub)
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.title, func(t *testing.T) {
			keyID := KeyID{
				Type: KeyTypeBabyJubJub,
				ID:   tc.keyID,
			}
			result, err := keyProvider.PublicKey(keyID)
			if tc.wantErr == "" {
				require.NoError(t, err)
				require.Equal(t, tc.want, result)
			} else {
				require.EqualError(t, err, tc.wantErr)
				require.Nil(t, result)
			}
		})
	}
}

func TestBJJKeyProvider_NoIdentity(t *testing.T) {
	if os.Getenv("TEST_MODE") == "GA" {
		t.Skip("SKIPPED")
	}

	k := testKMSSetup(t)

	keyID, err := k.KMS.CreateKey(KeyTypeBabyJubJub, nil)
	require.NoError(t, err)
	require.Equal(t, KeyTypeBabyJubJub, keyID.Type)

	privKey := testBJJKeyContent(t, k, keyID)

	pubKeyBytes, err := k.KMS.PublicKey(keyID)
	require.NoError(t, err)
	pubKey, err := DecodeBJJPubKey(pubKeyBytes)
	require.NoError(t, err)

	require.Equal(t, privKey.Public().Compress(), pubKey.Compress())

	// Test signature
	msg := new(big.Int).Sub(constants.Q, big.NewInt(10))
	digest := utils.SwapEndianness(msg.Bytes())
	sigBytes, err := k.KMS.Sign(context.Background(), keyID, digest)
	require.NoError(t, err)
	var sigComp babyjub.SignatureComp
	require.Len(t, sigBytes, len(sigComp))
	n := copy(sigComp[:], sigBytes)
	require.Equal(t, n, len(sigComp))

	sig, err := sigComp.Decompress()
	require.NoError(t, err)
	require.True(t, pubKey.VerifyPoseidon(msg, sig))

	identity, err := core.IDFromString("x2Uw18ATvY7mEsgfrrDipBmQQdPWAao4NmF56wGvp")
	require.NoError(t, err)

	did, err := core.ParseDIDFromID(identity)
	require.NoError(t, err)

	boundedKeyID, err := k.KMS.LinkToIdentity(context.Background(), keyID, *did)
	require.NoError(t, err)

	// check that old key ID is removed
	sec, err := k.VaultCli.Logical().Read("secret/data/" + keyID.ID)
	require.NoError(t, err)
	require.Nil(t, sec.Data["data"])

	// repeat tests for bounded keyID
	testBoundedBJJKey(t, k, boundedKeyID, *did)
}

func testBJJKeyContent(t testing.TB, k TestKMS, keyID KeyID) babyjub.PrivateKey {
	t.Helper()
	sec, err := k.VaultCli.Logical().Read("secret/data/" + keyID.ID)
	require.NoError(t, err)
	require.Len(t, sec.Data["data"], 2)

	keyType := sec.Data["data"].(map[string]interface{})[jsonKeyType].(string)
	require.Equal(t, string(KeyTypeBabyJubJub), keyType)

	privKeyHex := sec.Data["data"].(map[string]interface{})[jsonKeyData].(string)
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	require.NoError(t, err)
	var privKey babyjub.PrivateKey
	require.Equal(t, 32, copy(privKey[:], privKeyBytes))

	pubKeyBytes, err := k.KMS.PublicKey(keyID)
	require.NoError(t, err)
	compPubKey := privKey.Public().Compress()
	require.Equal(t, pubKeyBytes, compPubKey[:])

	return privKey
}

func testBoundedBJJKey(t *testing.T, k TestKMS, keyID KeyID, identity core.DID) {
	t.Helper()
	privKey := testBJJKeyContent(t, k, keyID)

	// Try to bind already bounded key to other identity: expect error
	otherIdentity, err := core.IDFromString("x2Uw18ATvY7mEsgfrrDipBmQQdPWAao4NmF56wGvp")
	require.NoError(t, err)
	otherDID, err := core.ParseDIDFromID(otherIdentity)
	require.NoError(t, err)

	_, err = k.KMS.LinkToIdentity(context.Background(), keyID, *otherDID)
	require.EqualError(t, err, "key ID does not looks like unbound")

	// Check public key
	pubKeyBytes, err := k.KMS.PublicKey(keyID)
	require.NoError(t, err)
	pubKey, err := DecodeBJJPubKey(pubKeyBytes)
	require.NoError(t, err)

	require.Equal(t, privKey.Public().Compress(), pubKey.Compress())

	// Test listing
	keys, err := k.KMS.KeysByIdentity(context.Background(), identity)
	require.NoError(t, err)
	found := false
	for _, k := range keys {
		if k.Type != KeyTypeBabyJubJub {
			continue
		}
		require.True(t,
			strings.HasPrefix(k.ID, keysPathPrefix+identity.String()+"/"))
		if k == keyID {
			found = true
			break
		}
	}
	require.True(t, found)

	// Test signature
	msg := new(big.Int).Sub(constants.Q, big.NewInt(10))
	digest := utils.SwapEndianness(msg.Bytes())
	sigBytes, err := k.KMS.Sign(context.Background(), keyID, digest)
	require.NoError(t, err)
	var sigComp babyjub.SignatureComp
	require.Len(t, sigBytes, len(sigComp))
	n := copy(sigComp[:], sigBytes)
	require.Equal(t, n, len(sigComp))

	sig, err := sigComp.Decompress()
	require.NoError(t, err)
	require.True(t, pubKey.VerifyPoseidon(msg, sig))

	_, err = k.VaultCli.Logical().Delete("secret/data/" + keyID.ID)
	require.NoError(t, err)
}
