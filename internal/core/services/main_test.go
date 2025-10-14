package services

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/assert"

	cache2 "github.com/polygonid/sh-id-platform/internal/cache"
	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/network"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/pubsub"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/internal/reversehash"
	"github.com/polygonid/sh-id-platform/internal/revocationstatus"
)

var (
	storage         *db.Storage
	vaultCli        *api.Client
	bjjKeyProvider  kms.KeyProvider
	keyStore        *kms.KMS
	cachex          cache2.Cache
	docLoader       loader.DocumentLoader
	cfg             config.Configuration
	identityService ports.IdentityService
	claimsService   ports.ClaimService
)

const ipfsGatewayURL = "http://127.0.0.1:8080"

const ipfsGateway = "http://localhost:8080"

// VaultTest returns the vault configuration to be used in tests.
// The vault token is obtained from environment vars.
// If there is no env var, it will try to parse the init.out file
// created by local docker image provided for TESTING purposes.
func vaultTest() config.KeyStore {
	return config.KeyStore{
		Address:                   "http://localhost:8200",
		PluginIden3MountPath:      "iden3",
		VaultUserPassAuthEnabled:  true,
		VaultUserPassAuthPassword: "issuernodepwd",
	}
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	log.Config(log.LevelDebug, log.OutputText, os.Stdout)
	conn := lookupPostgresURL()
	if conn == "" {
		conn = "postgres://postgres:postgres@localhost:5435"
	}

	cfgForTesting := config.Configuration{
		Database: config.Database{
			URL: conn,
		},
		KeyStore: vaultTest(),
	}
	s, teardown, err := tests.NewTestStorage(&cfgForTesting)
	defer teardown()
	if err != nil {
		log.Error(ctx, "failed to acquire test database", "err", err)
		os.Exit(1)
	}
	storage = s

	vaultCli, err = providers.VaultClient(ctx, providers.Config{
		Address:             cfgForTesting.KeyStore.Address,
		UserPassAuthEnabled: cfgForTesting.KeyStore.VaultUserPassAuthEnabled,
		Pass:                cfgForTesting.KeyStore.VaultUserPassAuthPassword,
	})
	if err != nil {
		log.Error(ctx, "failed to acquire vault client", "err", err)
		os.Exit(1)
	}

	bjjKeyProvider, err = kms.NewVaultPluginIden3KeyProvider(vaultCli, cfgForTesting.KeyStore.PluginIden3MountPath, kms.KeyTypeBabyJubJub)
	if err != nil {
		log.Error(ctx, "failed to create Iden3 Key Provider", "err", err)
		os.Exit(1)
	}

	ethKeyProvider, err := kms.NewVaultPluginIden3KeyProvider(vaultCli, cfgForTesting.KeyStore.PluginIden3MountPath, kms.KeyTypeEthereum)
	if err != nil {
		log.Error(ctx, "failed to create Iden3 Key Provider", "err", err)
		os.Exit(1)
	}

	keyStore = kms.NewKMS()
	err = keyStore.RegisterKeyProvider(kms.KeyTypeBabyJubJub, bjjKeyProvider)
	if err != nil {
		log.Error(ctx, "failed to register Key Provider", "err", err)
		os.Exit(1)
	}

	err = keyStore.RegisterKeyProvider(kms.KeyTypeEthereum, ethKeyProvider)
	if err != nil {
		log.Error(ctx, "failed to register eth Key Provider", "err", err)
		os.Exit(1)
	}

	cachex = cache2.NewMemoryCache()

	docLoader = loader.NewDocumentLoader(ipfsGatewayURL, false)
	cfg.Ethereum = cfgForTesting.Ethereum
	cfg.ServerUrl = "http://localhost:3001"

	// repositories
	claimsRepository := repositories.NewClaim()
	connectionRepository := repositories.NewConnection()
	identityRepository := repositories.NewIdentity()
	idenMerkleTreeRepository := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepository := repositories.NewIdentityState()
	sessionsRepository := repositories.NewSessionCached(cachex)
	revocationRepository := repositories.NewRevocation()
	keyRepository := repositories.NewKey(*storage)

	pubSub := pubsub.NewMock()
	mtService := NewIdentityMerkleTrees(idenMerkleTreeRepository)
	qrService := NewQrStoreService(cachex)
	t := &testing.T{}
	networkResolver, err := network.NewResolver(context.Background(), cfg, keyStore, common.CreateFile(t))
	assert.NoError(t, err)
	rhsFactory := reversehash.NewFactory(*networkResolver, reversehash.DefaultRHSTimeOut)
	revocationStatusResolver := revocationstatus.NewRevocationStatusResolver(*networkResolver)
	mediaTypeManager := NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage), string(protocol.CredentialFetchRequestMessageType)},
			protocol.RevocationStatusRequestMessageType: {"*"},
			protocol.DiscoverFeatureQueriesMessageType:  {"*"},
		},
		true,
	)
	schemaLoader := loader.NewDocumentLoader(ipfsGatewayURL, false)
	identityService = NewIdentity(keyStore, identityRepository, idenMerkleTreeRepository, identityStateRepository, mtService, qrService, claimsRepository, revocationRepository, connectionRepository, s, nil, sessionsRepository, pubSub, *networkResolver, rhsFactory, revocationStatusResolver, keyRepository)
	claimsService = NewClaim(claimsRepository, identityService, qrService, mtService, identityStateRepository, schemaLoader, storage, cfg.ServerUrl, pubSub, ipfsGatewayURL, revocationStatusResolver, mediaTypeManager, cfg.UniversalLinks)

	m.Run()
}

func lookupPostgresURL() string {
	con, ok := os.LookupEnv("POSTGRES_TEST_DATABASE")
	if !ok {
		return ""
	}
	return con
}
