package kms

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/providers"
)

var cfg providers.Config

type TestKMS struct {
	KMS        *KMS
	VaultCli   *api.Client
	writtenIDs []KeyID
	t          testing.TB
}

// VaultTest returns the vault configuration to be used in tests.
// The vault token is obtained from environment vars.
// If there is no env var, it will try to parse the init.out file
// created by local docker image provided for TESTING purposes.
func vaultTest() providers.Config {
	return providers.Config{
		Address:             "http://localhost:8200",
		UserPassAuthEnabled: true,
		Pass:                "issuernodepwd",
		MountPath:           "iden3",
	}
}

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	cfg = vaultTest()
	return m.Run()
}

// TestKMSSetup checks if Vault is available and setup connection to it.
// Also, it registers cleanup function on test complete.
func testKMSSetup(t testing.TB) TestKMS {
	k := TestKMS{t: t}
	var err error

	k.VaultCli, err = providers.VaultClient(context.Background(), vaultTest())
	require.NoError(t, err)

	k.KMS = NewKMS()

	err = k.KMS.RegisterKeyProvider(KeyTypeEthereum, NewVaultEthProvider(k.VaultCli, KeyTypeEthereum))
	require.NoError(t, err)

	err = k.KMS.RegisterKeyProvider(KeyTypeBabyJubJub, NewVaultBJJKeyProvider(k.VaultCli, KeyTypeBabyJubJub))
	require.NoError(t, err)

	t.Cleanup(k.Close)
	return k
}

// Close cleans up Vault on test complete.
func (tk *TestKMS) Close() {
	for _, k := range tk.writtenIDs {
		_, err := tk.VaultCli.Logical().Delete(absVaultSecretPath(k.ID))
		assert.NoError(tk.t, err)
	}
}
