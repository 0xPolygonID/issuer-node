package api

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/vault/api"
	auth "github.com/iden3/go-iden3-auth/v2"
	authLoaders "github.com/iden3/go-iden3-auth/v2/loaders"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"

	cache2 "github.com/polygonid/sh-id-platform/internal/cache"
	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/errors"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/network"
	"github.com/polygonid/sh-id-platform/internal/packagemanager"
	"github.com/polygonid/sh-id-platform/internal/payments"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/pubsub"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/internal/reversehash"
	"github.com/polygonid/sh-id-platform/internal/revocationstatus"
)

var (
	storage        *db.Storage
	vaultCli       *api.Client
	cfg            config.Configuration
	bjjKeyProvider kms.KeyProvider
	keyStore       *kms.KMS
	cachex         cache2.Cache
	schemaLoader   ld.DocumentLoader
)

const ipfsGatewayURL = "http://localhost:8080"

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
		KeyStore:       vaultTest(),
		UniversalLinks: config.UniversalLinks{BaseUrl: "https://testing.env"},
	}
	s, teardown, err := tests.NewTestStorage(&cfgForTesting)
	defer teardown()
	if err != nil {
		log.Error(ctx, "failed to acquire test database", "err", err)
		os.Exit(1)
	}
	storage = s

	cachex = cache2.NewMemoryCache()

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
	cfg.UniversalLinks = config.UniversalLinks{BaseUrl: "https://testing.env"}
	cfg.Circuit.Path = "../../pkg/credentials/circuits"
	schemaLoader = loader.NewDocumentLoader(ipfsGatewayURL, false)
	m.Run()
}

func getHandler(ctx context.Context, server StrictServerInterface) http.Handler {
	mux := chi.NewRouter()
	RegisterStatic(mux)
	return HandlerWithOptions(
		NewStrictHandlerWithOptions(
			server,
			middlewares(ctx),
			StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  errors.RequestErrorHandlerFunc,
				ResponseErrorHandlerFunc: errors.ResponseErrorHandlerFunc,
			},
		),
		ChiServerOptions{
			BaseRouter:       mux,
			ErrorHandlerFunc: ErrorHandlerFunc,
		})
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

func NewClaimsMock() ports.ClaimService {
	return nil
}

func NewSchemaMock() ports.SchemaService {
	return nil
}

func NewConnectionsMock() ports.ConnectionService {
	return nil
}

func NewLinkMock() ports.LinkService {
	return nil
}

type repos struct {
	claims         ports.ClaimRepository
	connection     ports.ConnectionRepository
	identity       ports.IdentityRepository
	idenMerkleTree ports.IdentityMerkleTreeRepository
	identityState  ports.IdentityStateRepository
	links          ports.LinkRepository
	payments       ports.PaymentRepository
	schemas        ports.SchemaRepository
	sessions       ports.SessionRepository
	revocation     ports.RevocationRepository
	displayMethod  ports.DisplayMethodRepository
	keyRepository  ports.KeyRepository
	verification   ports.VerificationRepository
}

type servicex struct {
	credentials   ports.ClaimService
	identity      ports.IdentityService
	schema        ports.SchemaService
	links         ports.LinkService
	payments      ports.PaymentService
	qrs           ports.QrStoreService
	displayMethod ports.DisplayMethodService
	keyService    ports.KeyService
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

func newTestServer(t *testing.T, st *db.Storage) *testServer {
	t.Helper()
	if st == nil {
		st = storage
	}
	repos := repos{
		claims:         repositories.NewClaim(),
		connection:     repositories.NewConnection(),
		identity:       repositories.NewIdentity(),
		idenMerkleTree: repositories.NewIdentityMerkleTreeRepository(),
		identityState:  repositories.NewIdentityState(),
		links:          repositories.NewLink(*st),
		payments:       repositories.NewPayment(*st),
		sessions:       repositories.NewSessionCached(cachex),
		schemas:        repositories.NewSchema(*st),
		revocation:     repositories.NewRevocation(),
		displayMethod:  repositories.NewDisplayMethod(*st),
		keyRepository:  repositories.NewKey(*st),
		verification:   repositories.NewVerification(*st),
	}

	pubSub := pubsub.NewMock()

	networkResolver, err := network.NewResolver(context.Background(), cfg, keyStore, common.CreateFile(t))
	require.NoError(t, err)
	revocationStatusResolver := revocationstatus.NewRevocationStatusResolver(*networkResolver)

	paymentSettings, err := payments.SettingsFromReader(common.NewMyYAMLReader([]byte(`
80002:
  PaymentRails: 0xF8E49b922D5Fb00d3EdD12bd14064f275726D339
  PaymentOptions: 
    - ID: 1
      Name: AmoyNative
      Type: Iden3PaymentRailsRequestV1
    - ID: 2
      Name: Amoy USDT
      Type: Iden3PaymentRailsERC20RequestV1
      ContractAddress: 0x2FE40749812FAC39a0F380649eF59E01bccf3a1A
      Features: []
    - ID: 3
      Name: Amoy USDC
      Type: Iden3PaymentRailsERC20RequestV1
      ContractAddress: 0x2FE40749812FAC39a0F380649eF59E01bccf3a1A
      Features:
        - EIP-2612
59141:
  PaymentRails: 0x40E3EF221AA93F6Fe997c9b0393322823Bb207d3
  PaymentOptions: 
    - ID: 4
      Name: LineaSepoliaNative
      Type: Iden3PaymentRailsRequestV1
    - ID: 5
      Name: Linea Sepolia USDT
      Type: Iden3PaymentRailsERC20RequestV1
      ContractAddress: 0xb0101c1Ffdd1213B886FebeF6F07442e48990c9C
      Features: []
    - ID: 6
      Name: Linea Sepolia USDC
      Type: Iden3PaymentRailsERC20RequestV1
      ContractAddress: 0xb0101c1Ffdd1213B886FebeF6F07442e48990c9C
      Features:
        - EIP-2612
2442:
  PaymentRails: 0x09c269e74d8B47c98537Acd6CbEe8056806F4c70
  PaymentOptions: 
    - ID: 7
      Name: ZkEvmNative
      Type: Iden3PaymentRailsRequestV1
    - ID: 8
      Name: ZkEvm USDT
      Type: Iden3PaymentRailsERC20RequestV1
      ContractAddress: 0x986caE6ADcF5da2a1514afc7317FBdeE0B4048Db
      Features: []
    - ID: 9
      Name: ZkEvm USDC
      Type: Iden3PaymentRailsERC20RequestV1
      ContractAddress: 0x986caE6ADcF5da2a1514afc7317FBdeE0B4048Db
      Features:
        - EIP-2612
`,
	)))
	require.NoError(t, err)

	mtService := services.NewIdentityMerkleTrees(repos.idenMerkleTree)
	qrService := services.NewQrStoreService(cachex)
	rhsFactory := reversehash.NewFactory(*networkResolver, reversehash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, repos.identity, repos.idenMerkleTree, repos.identityState, mtService, qrService, repos.claims, repos.revocation, repos.connection, st, nil, repos.sessions, pubSub, *networkResolver, rhsFactory, revocationStatusResolver, repos.keyRepository)
	connectionService := services.NewConnection(repos.connection, repos.claims, st)
	displayMethodService := services.NewDisplayMethod(repos.displayMethod)
	schemaService := services.NewSchema(repos.schemas, schemaLoader, displayMethodService)
	paymentService := services.NewPaymentService(repos.payments, *networkResolver, schemaService, paymentSettings, keyStore)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(repos.claims, identityService, qrService, mtService, repos.identityState, schemaLoader, st, cfg.ServerUrl, pubSub, ipfsGatewayURL, revocationStatusResolver, mediaTypeManager, cfg.UniversalLinks)
	accountService := services.NewAccountService(*networkResolver)
	linkService := services.NewLinkService(storage, claimsService, qrService, repos.claims, repos.links, repos.schemas, schemaLoader, repos.sessions, pubSub, identityService, *networkResolver, cfg.UniversalLinks)
	keyService := services.NewKey(keyStore, claimsService, repos.keyRepository)

	universalDIDResolverUrl := auth.UniversalResolverURL
	if cfg.UniversalDIDResolver.UniversalResolverURL != nil && *cfg.UniversalDIDResolver.UniversalResolverURL != "" {
		universalDIDResolverUrl = *cfg.UniversalDIDResolver.UniversalResolverURL
	}
	universalDIDResolverHandler := packagemanager.NewUniversalDIDResolverHandler(universalDIDResolverUrl)
	verificationKeyLoader := &authLoaders.FSKeyLoader{Dir: cfg.Circuit.Path + "/authV2"}
	verifier, err := auth.NewVerifier(verificationKeyLoader, networkResolver.GetStateResolvers(), auth.WithDIDResolver(universalDIDResolverHandler))
	require.NoError(t, err)

	verificationService := services.NewVerificationService(networkResolver, cachex, repos.verification, verifier)

	server := NewServer(&cfg, identityService, accountService, connectionService, claimsService, qrService, NewPublisherMock(), NewPackageManagerMock(), *networkResolver, nil, schemaService, linkService, displayMethodService, keyService, paymentService, verificationService)

	return &testServer{
		Server: server,
		Repos:  repos,
		Services: servicex{
			credentials:   claimsService,
			identity:      identityService,
			links:         linkService,
			payments:      paymentService,
			qrs:           qrService,
			schema:        schemaService,
			displayMethod: displayMethodService,
			keyService:    keyService,
		},
		Infra: infra{
			db:     st,
			pubSub: pubSub,
		},
	}
}
