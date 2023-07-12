package api

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/vault/api"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/iden3comm"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/errors"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/pkg/cache"
)

var (
	storage        *db.Storage
	vaultCli       *api.Client
	cfg            config.Configuration
	bjjKeyProvider kms.KeyProvider
	keyStore       *kms.KMS
	cachex         cache.Cache
)

const ipfsGateway = "http://localhost:8080"

func TestMain(m *testing.M) {
	ctx := context.Background()
	conn := lookupPostgresURL()
	if conn == "" {
		conn = "postgres://postgres:postgres@localhost:5435"
	}

	cfgForTesting := config.Configuration{
		Database: config.Database{
			URL: conn,
		},
		KeyStore: config.VaultTest(),
	}
	s, teardown, err := tests.NewTestStorage(&cfgForTesting)
	defer teardown()
	if err != nil {
		log.Error(ctx, "failed to acquire test database", "err", err)
		os.Exit(1)
	}
	storage = s

	cachex = cache.NewMemoryCache()

	vaultCli, err = providers.NewVaultClient(cfgForTesting.KeyStore.Address, cfgForTesting.KeyStore.Token)
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

	cfg.ServerUrl = "https://testing.env/"

	m.Run()
}

func getHandler(ctx context.Context, server *Server) http.Handler {
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

func (kpm *KMSMock) CreateKey(kt kms.KeyType, identity *core.DID) (kms.KeyID, error) {
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

func (kpm *KMSMock) KeysByIdentity(ctx context.Context, identity core.DID) ([]kms.KeyID, error) {
	var keys []kms.KeyID
	return keys, nil
}

func (kpm *KMSMock) LinkToIdentity(ctx context.Context, keyID kms.KeyID, identity core.DID) (kms.KeyID, error) {
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
