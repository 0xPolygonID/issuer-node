package kms

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/vault/api"
	core "github.com/iden3/go-iden3-core"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/providers"
)

func TestVaultPluginBJJProvider_Ethereum(t *testing.T) {
	if os.Getenv("TEST_MODE") == "GA" {
		fmt.Println("SKIPPED")
		t.Skip()
	}

	vaultCli, mountPath := setupPluginBJJProvider(t)

	getKeyPath := func(kID KeyID) keyPathT {
		return keyPathT{keyID: kID.ID, mountPath: mountPath}
	}

	keysPath := path.Join(mountPath, randString(6))
	kp, err := NewVaultPluginIden3KeyProvider(vaultCli, keysPath, KeyTypeEthereum)
	require.NoError(t, err)
	ctx := context.Background()

	// register callback to delete key
	rmKey := func(kPath keyPathT) {
		t.Cleanup(func() {
			_, err2 := vaultCli.Logical().Delete(kPath.keys())
			if err2 != nil {
				t.Error(err2)
			}
		})
	}

	// generate new random key not bound to any identity
	newKey, err := kp.New(nil)
	require.NoError(t, err)
	rmKey(getKeyPath(newKey))
	require.Equal(t, newKey.Type, KeyTypeEthereum)

	// link key to identity
	did := randomDID()

	require.NoError(t, err)

	boundKey, err := kp.LinkToIdentity(ctx, newKey, did)
	require.NoError(t, err)
	rmKey(getKeyPath(boundKey))
	require.Equal(t, newKey.Type, KeyTypeEthereum)

	// test link to same identity without error
	boundKey2, err := kp.LinkToIdentity(ctx, boundKey, did)
	require.NoError(t, err)
	rmKey(getKeyPath(boundKey2))
	require.Equal(t, boundKey, boundKey2)

	// test public key
	pubKeyBytes, err := kp.PublicKey(boundKey)
	require.NoError(t, err)
	privKey := getETHPrivateKey(t, vaultCli, getKeyPath(boundKey))
	pubKey2, ok := privKey.Public().(*ecdsa.PublicKey)
	require.True(t, ok)
	pubKey2Bytes := crypto.CompressPubkey(pubKey2)
	require.Equal(t, pubKeyBytes, pubKey2Bytes)

	// test signing
	digest := make([]byte, crypto.DigestLength)
	_, err = rand.Read(digest)
	require.NoError(t, err)
	sig1Bytes, err := kp.Sign(ctx, boundKey, digest)
	require.NoError(t, err)
	sig2Bytes, err := crypto.Sign(digest, privKey)
	require.NoError(t, err)
	require.Equal(t, sig1Bytes, sig2Bytes)

	// generate new random key bounded to identity
	newKey2, err := kp.New(&did)
	require.NoError(t, err)
	rmKey(getKeyPath(newKey2))
	require.Equal(t, newKey.Type, KeyTypeEthereum)

	// test list method sees both keys
	keyIDs, err := kp.ListByIdentity(ctx, did)
	require.NoError(t, err)
	sort.Slice(keyIDs, func(i, j int) bool {
		return keyIDs[i].ID < keyIDs[j].ID
	})

	wantKeyIDs := []KeyID{boundKey, newKey2}
	sort.Slice(wantKeyIDs, func(i, j int) bool {
		return wantKeyIDs[i].ID < wantKeyIDs[j].ID
	})

	require.Equal(t, wantKeyIDs, keyIDs)
}

func randString(ln int) string {
	bs := make([]byte, ln)
	_, err := rand.Read(bs)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bs)
}

func randomDID() core.DID {
	typ, err := core.BuildDIDType(core.DIDMethodIden3, core.Polygon, core.Mumbai)
	var genesis [27]byte
	if err != nil {
		panic(err)
	}

	_, err = rand.Read(genesis[:])
	if err != nil {
		panic(err)
	}

	id := core.NewID(typ, genesis)

	did, err := core.ParseDIDFromID(id)
	if err != nil {
		panic(err)
	}

	return *did
}

func setupPluginBJJProvider(t *testing.T) (vaultCli *api.Client, mountPath string) {
	var err error

	vaultCli, err = providers.NewVaultClient(testVaultConfig(t))
	require.NoError(t, err)

	mountPath = cfg.PluginIden3MountPath
	if mountPath == "" {
		t.Skip("IDEN3 plugin mount path is not set")
	}

	return
}

func getETHPrivateKey(t testing.TB, cli *api.Client, keyPath keyPathT) *ecdsa.PrivateKey {
	secret, err := cli.Logical().Read(keyPath.private())
	require.NoError(t, err)
	data, err := getSecretData(secret)
	require.NoError(t, err)
	require.Equal(t, "ethereum", data[jsonKeyType])

	keyBytes, err := hex.DecodeString(data["private_key"].(string))
	require.NoError(t, err)
	privKey, err := crypto.ToECDSA(keyBytes)
	require.NoError(t, err)

	return privKey
}

func (p keyPathT) private() string {
	return p.join("private")
}
