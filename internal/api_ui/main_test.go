package api_ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	vaultApi "github.com/hashicorp/vault/api"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/errors"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/pkg/cache"
)

var (
	storage        *db.Storage
	vaultCli       *vaultApi.Client
	cfg            config.Configuration
	bjjKeyProvider kms.KeyProvider
	keyStore       *kms.KMS
	cachex         cache.Cache
	schemaLoader   loader.DocumentLoader
)

const ipfsGatewayURL = "http://localhost:8080"

// VaultTest returns the vault configuration to be used in tests.
// The vault token is obtained from environment vars.
// If there is no env var, it will try to parse the init.out file
// created by local docker image provided for TESTING purposes.
func vaultTest() config.KeyStore {
	return config.KeyStore{
		Address:              "http://localhost:8200",
		PluginIden3MountPath: "iden3",
		UserPassEnabled:      true,
		UserPassPassword:     "issuernodepwd",
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
	vaultCli, err = providers.VaultClient(context.Background(), providers.Config{
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

	keyStore = kms.NewKMS()
	err = keyStore.RegisterKeyProvider(kms.KeyTypeBabyJubJub, bjjKeyProvider)
	if err != nil {
		log.Error(ctx, "failed to register Key Provider", "err", err)
		os.Exit(1)
	}

	cfg.ServerUrl = "https://testing.env"
	cfg.Ethereum = cfgForTesting.Ethereum
	schemaLoader = loader.NewDocumentLoader(ipfsGatewayURL)

	m.Run()
}

func getHandler(ctx context.Context, server *Server) http.Handler {
	mux := chi.NewRouter()
	RegisterStatic(mux)
	return HandlerWithOptions(
		NewStrictHandlerWithOptions(
			server,
			middlewares(ctx),
			StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  errors.RequestErrorHandlerFunc,
				ResponseErrorHandlerFunc: errors.ResponseErrorHandlerFunc,
			}),
		ChiServerOptions{
			BaseRouter:       mux,
			ErrorHandlerFunc: errorHandlerFunc,
		},
	)
}

func middlewares(ctx context.Context) []StrictMiddlewareFunc {
	usr, pass := authOk()
	return []StrictMiddlewareFunc{
		LogMiddleware(ctx),
		BasicAuthMiddleware(ctx, usr, pass),
	}
}

func errorHandlerFunc(w http.ResponseWriter, _ *http.Request, err error) {
	switch err.(type) {
	case *InvalidParamFormatError:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"message": err.Error()})
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func authOk() (string, string) {
	return "user", "password"
}

func authWrong() (string, string) {
	return "", ""
}

func checkQRfetchURL(t *testing.T, qrLink string) string {
	t.Helper()
	qrURL, err := url.Parse(qrLink)
	require.NoError(t, err)
	assert.Equal(t, "iden3comm", qrURL.Scheme)
	vals, err := url.ParseQuery(qrURL.RawQuery)
	require.NoError(t, err)
	val, found := vals["request_uri"]
	require.True(t, found)
	fetchURL, err := url.QueryUnescape(val[0])
	require.NoError(t, err)
	return fetchURL
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

func NewIdentityMock() ports.IdentityService {
	return nil
}

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
