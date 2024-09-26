package protocol

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/pkg/helpers"
	networkPkg "github.com/polygonid/sh-id-platform/pkg/network"
)

var cfgForTesting config.Configuration

func keyStoreConfig(t *testing.T) config.KeyStore {
	t.Helper()
	return config.KeyStore{
		Address:                   "http://localhost:8200",
		PluginIden3MountPath:      "iden3",
		VaultUserPassAuthEnabled:  true,
		VaultUserPassAuthPassword: "issuernodepwd",
	}
}

func kmsTest(t *testing.T, ctx context.Context) *kms.KMS {
	t.Helper()
	currentPath, err := os.Getwd()
	if err != nil {
		t.Fail()
	}

	circuitsPath := filepath.Join(currentPath, "pkg/credentials/circuits")
	if strings.Contains(currentPath, "pkg/protocol") {
		circuitsPath = strings.Replace(currentPath, "pkg/protocol", "pkg/credentials/circuits", 1)
	}
	cfgForTesting = config.Configuration{
		KeyStore: keyStoreConfig(t),
		Circuit: config.Circuit{
			Path: circuitsPath,
		},
	}

	vaultCli, err := providers.VaultClient(ctx, providers.Config{
		Address:             cfgForTesting.KeyStore.Address,
		UserPassAuthEnabled: cfgForTesting.KeyStore.VaultUserPassAuthEnabled,
		Pass:                cfgForTesting.KeyStore.VaultUserPassAuthPassword,
	})
	if err != nil {
		log.Error(ctx, "failed to acquire vault client", "err", err)
		t.Fail()
	}

	bjjKeyProvider, err := kms.NewVaultPluginIden3KeyProvider(vaultCli, cfgForTesting.KeyStore.PluginIden3MountPath, kms.KeyTypeBabyJubJub)
	if err != nil {
		log.Error(ctx, "failed to create Iden3 Key Provider", "err", err)
		t.Fail()
	}
	ethKeyProvider, err := kms.NewVaultPluginIden3KeyProvider(vaultCli, cfgForTesting.KeyStore.PluginIden3MountPath, kms.KeyTypeEthereum)
	if err != nil {
		log.Error(ctx, "failed to create Iden3 Key Provider", "err", err)
		t.Fail()
	}

	keyStore := kms.NewKMS()
	err = keyStore.RegisterKeyProvider(kms.KeyTypeBabyJubJub, bjjKeyProvider)
	if err != nil {
		log.Error(ctx, "failed to register bjj Key Provider", "err", err)
		t.Fail()
	}

	err = keyStore.RegisterKeyProvider(kms.KeyTypeEthereum, ethKeyProvider)
	if err != nil {
		log.Error(ctx, "failed to register eth Key Provider", "err", err)
		t.Fail()
	}

	return keyStore
}

func TestJWS(t *testing.T) {
	const token = `eyJhbGciOiJFUzI1NksiLCJraWQiOiJkaWQ6ZXhhbXBsZToxMjMjSlV2cGxsTUVZVVoyam9PNTlVTnVpX1hZRHF4VnFpRkxMQUo4a2xXdVBCdyIsInR5cCI6ImFwcGxpY2F0aW9uL2lkZW4zY29tbS1zaWduZWQtanNvbiJ9.eyJ0eXBlIjoiaHR0cHM6Ly9pZGVuMy1jb21tdW5pY2F0aW9uLmlvL2F1dGhvcml6YXRpb24vMS4wL3Jlc3BvbnNlIiwiZnJvbSI6ImRpZDpleGFtcGxlOjEyMyIsImJvZHkiOnsic2NvcGUiOlt7InR5cGUiOiJ6ZXJva25vd2xlZGdlIiwiY2lyY3VpdF9pZCI6ImF1dGgiLCJwdWJfc2lnbmFscyI6WyIxIiwiMTgzMTE1NjA1MjUzODMzMTk3MTkzMTEzOTQ5NTcwNjQ4MjAwOTEzNTQ5NzYzMTA1OTk4MTg3OTcxNTcxODk1Njg2MjE0NjY5NTA4MTEiLCIzMjM0MTY5MjUyNjQ2NjYyMTc2MTcyODg1Njk3NDI1NjQ3MDM2MzI4NTA4MTYwMzU3NjEwODQwMDI3MjAwOTAzNzczNTMyOTc5MjAiXSwicHJvb2ZfZGF0YSI6eyJwaV9hIjpbIjExMTMwODQzMTUwNTQwNzg5Mjk5NDU4OTkwNTg2MDIwMDAwNzE5MjgwMjQ2MTUzNzk3ODgyODQzMjE0MjkwNTQxOTgwNTIyMzc1MDcyIiwiMTMwMDg0MTkxMjk0Mzc4MTcyMzAyMjAzMjM1NTgzNjg5MzgzMTEzMjkyMDc4Mzc4ODQ1NTUzMTgzODI1NDQ2NTc4NDYwNTc2MjcxMyIsIjEiXSwicGlfYiI6W1siMjA2MTU3Njg1MzY5ODg0MzgzMzY1Mzc3Nzc5MDkwNDIzNTIwNTYzOTI4NjIyNTE3ODU3MjI3OTY2Mzc1OTAyMTIxNjA1NjEzNTE2NTYiLCIxMDM3MTE0NDgwNjEwNzc3ODg5MDUzODg1NzcwMDg1NTEwODY2NzYyMjA0MjIxNTA5Njk3MTc0NzIwMzEwNTk5NzQ1NDYyNTgxNDA4MCJdLFsiMTk1OTg1NDEzNTA4MDQ0Nzg1NDkxNDEyMDc4MzUwMjg2NzExMTEwNjM5MTU2MzU1ODA2Nzk2OTQ5MDc2MzU5MTQyNzk5Mjg2Nzc4MTIiLCIxNTI2NDU1MzA0NTUxNzA2NTY2OTE3MTU4NDk0Mzk2NDMyMjExNzM5NzY0NTE0NzAwNjkwOTE2NzQyNzgwOTgzNzkyOTQ1ODAxMjkxMyJdLFsiMSIsIjAiXV0sInBpX2MiOlsiMTY0NDMzMDkyNzk4MjU1MDg4OTMwODYyNTEyOTAwMDM5MzY5MzUwNzczNDg3NTQwOTc0NzA4MTg1MjM1NTgwODI1MDIzNjQ4MjIwNDkiLCIyOTg0MTgwMjI3NzY2MDQ4MTAwNTEwMTIwNDA3MTUwNzUyMDUyMzM0NTcxODc2NjgxMzA0OTk5NTk1NTQ0MTM4MTU1NjExOTYzMjczIiwiMSJdLCJwcm90b2NvbCI6IiJ9fV19fQ._p8wS2JZELczn33_uB6EfmXzZ3RaizJVZIEclTT_UWS-xtPR6jpcthmRZGU1yrBQCNsf2ScWqvzzAV3DOJuKsg`
	const exampleDidDocJS = `{"@context":["https://www.w3.org/ns/did/v1","https://w3id.org/security/suites/secp256k1recovery-2020/v2",{"esrs2020":"https://identity.foundation/EcdsaSecp256k1RecoverySignature2020#","privateKeyJwk":{"@id":"esrs2020:privateKeyJwk","@type":"@json"},"publicKeyHex":"esrs2020:publicKeyHex","privateKeyHex":"esrs2020:privateKeyHex","ethereumAddress":"esrs2020:ethereumAddress"}],"id":"did:example:123","verificationMethod":[{"id":"did:example:123#JUvpllMEYUZ2joO59UNui_XYDqxVqiFLLAJ8klWuPBw","controller":"did:example:123","type":"EcdsaSecp256k1VerificationKey2019","publicKeyJwk":{"crv":"secp256k1","kid":"JUvpllMEYUZ2joO59UNui_XYDqxVqiFLLAJ8klWuPBw","kty":"EC","x":"_dV63sPUOOojf-RrM-4eAW7aa1hcPifqZmhsLqU1hHk","y":"Rjk_gUUlLupor-Z-KHs-2bMWhbpsOwAGCnO5sSQtaPc"}}],"authentication":["did:example:123#JUvpllMEYUZ2joO59UNui_XYDqxVqiFLLAJ8klWuPBw"]}`

	ctx := context.Background()

	keyStore := kmsTest(t, ctx)
	networkResolver, err := networkPkg.NewResolver(context.Background(), cfgForTesting, keyStore, helpers.CreateFile(t))
	require.NoError(t, err)

	packager, err := InitPackageManager(ctx, networkResolver.GetSupportedContracts(), cfgForTesting.Circuit.Path, func(did string) (*verifiable.DIDDocument, error) {
		didDoc := &verifiable.DIDDocument{}
		err := json.Unmarshal([]byte(exampleDidDocJS), didDoc)
		require.NoError(t, err)
		return didDoc, nil
	})
	require.NoError(t, err)
	message, mediaType, err := packager.Unpack([]byte(token))
	require.NoError(t, err)
	require.Equal(t, iden3comm.MediaType("application/iden3comm-signed-json"), mediaType)
	require.NotNil(t, message)
}
