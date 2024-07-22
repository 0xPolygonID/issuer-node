package api

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/vault/api"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/errors"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/cache"
	"github.com/polygonid/sh-id-platform/pkg/credentials/revocation_status"
	"github.com/polygonid/sh-id-platform/pkg/helpers"
	networkPkg "github.com/polygonid/sh-id-platform/pkg/network"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

var (
	storage        *db.Storage
	vaultCli       *api.Client
	cfg            config.Configuration
	bjjKeyProvider kms.KeyProvider
	keyStore       *kms.KMS
	cachex         cache.Cache
	schemaLoader   ld.DocumentLoader
)

const ipfsGatewayURL = "http://localhost:8080"

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
		KeyStore: config.VaultTest(),
		Ethereum: config.Ethereum{
			URL:            "https://polygon-mumbai.g.alchemy.com/v2/xaP2_",
			ResolverPrefix: "polygon:mumbai",
		},
	}
	s, teardown, err := tests.NewTestStorage(&cfgForTesting)
	defer teardown()
	if err != nil {
		log.Error(ctx, "failed to acquire test database", "err", err)
		os.Exit(1)
	}
	storage = s

	cachex = cache.NewMemoryCache()

	vaultCli, err = providers.VaultClient(ctx, providers.Config{
		Address:             cfgForTesting.KeyStore.Address,
		UserPassAuthEnabled: cfgForTesting.KeyStore.UserPassEnabled,
		Pass:                cfgForTesting.KeyStore.UserPassPassword,
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
		log.Error(ctx, "failed to register bjj Key Provider", "err", err)
		os.Exit(1)
	}

	err = keyStore.RegisterKeyProvider(kms.KeyTypeEthereum, ethKeyProvider)
	if err != nil {
		log.Error(ctx, "failed to register eth Key Provider", "err", err)
		os.Exit(1)
	}

	cfg.ServerUrl = "https://testing.env"
	cfg.Ethereum = cfgForTesting.Ethereum
	schemaLoader = loader.NewDocumentLoader(ipfsGatewayURL)
	m.Run()
}

func getHandler(ctx context.Context, server StrictServerInterface) http.Handler {
	mux := chi.NewRouter()
	RegisterStatic(mux)
	return HandlerFromMux(NewStrictHandlerWithOptions(
		server,
		middlewares(ctx),
		StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  errors.RequestErrorHandlerFunc,
			ResponseErrorHandlerFunc: errors.ResponseErrorHandlerFunc,
		},
	), mux)
}

func middlewares(ctx context.Context) []StrictMiddlewareFunc {
	usr, pass := authOk()
	return []StrictMiddlewareFunc{
		LogMiddleware(ctx),
		BasicAuthMiddleware(ctx, usr, pass),
	}
}

func authOk() (string, string) {
	return "user", "password"
}

func authWrong() (string, string) {
	return "", ""
}

func lookupPostgresURL() string {
	con, ok := os.LookupEnv("POSTGRES_TEST_DATABASE")
	if !ok {
		return ""
	}
	return con
}

type KMSMock struct{}

func (kpm *KMSMock) RegisterKeyProvider(kt kms.KeyType, kp kms.KeyProvider) error {
	return nil
}

func (kpm *KMSMock) CreateKey(kt kms.KeyType, identity *w3c.DID) (kms.KeyID, error) {
	var key kms.KeyID
	return key, nil
}

func (kpm *KMSMock) PublicKey(keyID kms.KeyID) ([]byte, error) {
	var pubKey []byte
	return pubKey, nil
}

func (kpm *KMSMock) Sign(ctx context.Context, keyID kms.KeyID, data []byte) ([]byte, error) {
	var signed []byte
	return signed, nil
}

func (kpm *KMSMock) KeysByIdentity(ctx context.Context, identity w3c.DID) ([]kms.KeyID, error) {
	var keys []kms.KeyID
	return keys, nil
}

func (kpm *KMSMock) LinkToIdentity(ctx context.Context, keyID kms.KeyID, identity w3c.DID) (kms.KeyID, error) {
	var key kms.KeyID
	return key, nil
}

// TODO: add package manager mocks
func NewPackageManagerMock() *iden3comm.PackageManager {
	return &iden3comm.PackageManager{}
}

func NewPublisherMock() ports.Publisher {
	return nil
}

func NewIdentityMock() ports.IdentityService { return nil }

func NewClaimsMock() ports.ClaimsService {
	return nil
}

func NewSchemaMock() ports.SchemaService {
	return nil
}

func NewConnectionsMock() ports.ConnectionsService {
	return nil
}

func NewLinkMock() ports.LinkService {
	return nil
}

type repos struct {
	claims         ports.ClaimsRepository
	connection     ports.ConnectionsRepository
	identity       ports.IndentityRepository
	idenMerkleTree ports.IdentityMerkleTreeRepository
	identityState  ports.IdentityStateRepository
	sessions       ports.SessionRepository
	revocation     ports.RevocationRepository
}

type servicex struct {
	credentials ports.ClaimsService
	identity    ports.IdentityService
}

type infra struct {
	db     *db.Storage
	pubSub *pubsub.Mock
}

type testServer struct {
	*Server
	Repos    repos
	Services servicex
	Infra    infra
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	repos := repos{
		claims:         repositories.NewClaims(),
		connection:     repositories.NewConnections(),
		identity:       repositories.NewIdentity(),
		idenMerkleTree: repositories.NewIdentityMerkleTreeRepository(),
		identityState:  repositories.NewIdentityState(),
		sessions:       repositories.NewSessionCached(cachex),
		revocation:     repositories.NewRevocation(),
	}
	pubSub := pubsub.NewMock()

	networkResolver, err := networkPkg.NewResolver(context.Background(), cfg, keyStore, helpers.CreateFile(t))
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)

	mtService := services.NewIdentityMerkleTrees(repos.idenMerkleTree)
	qrService := services.NewQrStoreService(cachex)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, repos.identity, repos.idenMerkleTree, repos.identityState, mtService, qrService, repos.claims, repos.revocation, repos.connection, storage, nil, repos.sessions, pubSub, *networkResolver, rhsFactory, revocationStatusResolver)
	connectionService := services.NewConnection(repos.connection, repos.claims, storage)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(repos.claims, identityService, qrService, mtService, repos.identityState, schemaLoader, storage, cfg.ServerUrl, pubSub, ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	accountService := services.NewAccountService(*networkResolver)
	server := NewServer(&cfg, identityService, accountService, connectionService, claimsService, qrService, NewPublisherMock(), NewPackageManagerMock(), *networkResolver, nil)

	return &testServer{
		Server: server,
		Repos:  repos,
		Services: servicex{
			credentials: claimsService,
			identity:    identityService,
		},
		Infra: infra{
			db:     storage,
			pubSub: pubSub,
		},
	}
}
