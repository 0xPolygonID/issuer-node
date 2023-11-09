package kms

import (
	"context"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/stretchr/testify/require"
)

func TestEthKeyProvider(t *testing.T) {
	k := testKMSSetup(t)

	_, err := k.KMS.CreateKey(KeyTypeEthereum, nil)
	require.EqualError(t, err,
		"Ethereum keys can be created only for non-nil identities")

	identity, err := core.IDFromString("x2Uw18ATvY7mEsgfrrDipBmQQdPWAao4NmF56wGvp")
	require.NoError(t, err)

	did, err := core.ParseDIDFromID(identity)
	require.NoError(t, err)

	keyID, err := k.KMS.CreateKey(KeyTypeEthereum, did)
	require.NoError(t, err)
	require.Equal(t, KeyTypeEthereum, keyID.Type)

	sec, err := k.VaultCli.Logical().Read("secret/data/" + keyID.ID)
	require.NoError(t, err)
	require.Len(t, sec.Data["data"], 2)

	keyType, ok := sec.Data["data"].(map[string]interface{})[jsonKeyType].(string)
	require.True(t, ok)
	require.Equal(t, string(KeyTypeEthereum), keyType)

	privKeyHex, ok := sec.Data["data"].(map[string]interface{})[jsonKeyData].(string)
	require.True(t, ok)
	privKey, err := crypto.HexToECDSA(privKeyHex)
	require.NoError(t, err)

	pubKeyBytes, err := k.KMS.PublicKey(keyID)
	require.NoError(t, err)
	pubKey, err := DecodeETHPubKey(pubKeyBytes)
	require.NoError(t, err)

	require.True(t, privKey.PublicKey.Equal(pubKey))

	// Test listing
	keys, err := k.KMS.KeysByIdentity(context.Background(), *did)
	require.NoError(t, err)
	found := false
	for _, k := range keys {
		if k.Type != KeyTypeEthereum {
			continue
		}
		require.True(t,
			strings.HasPrefix(k.ID, keysPathPrefix+did.String()+"/"))
		if k == keyID {
			found = true
			break
		}
	}
	require.True(t, found)

	// Test signature
	text := []byte("abc")
	digest := crypto.Keccak256(text)
	sig, err := k.KMS.Sign(context.Background(), keyID, digest)
	require.NoError(t, err)
	require.True(t, crypto.VerifySignature(pubKeyBytes, digest, sig[:64]))

	sigPublicKeyECDSA, err := crypto.SigToPub(digest, sig)
	require.NoError(t, err)
	require.True(t, sigPublicKeyECDSA.Equal(pubKey))

	_, err = k.VaultCli.Logical().Delete("secret/data/" + keyID.ID)
	require.NoError(t, err)
}
