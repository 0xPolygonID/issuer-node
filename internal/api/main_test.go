package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/vault/api"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/providers"
)

var (
	storage        *db.Storage
	vaultCli       *api.Client
	cfg            config.Configuration
	bjjKeyProvider kms.KeyProvider
	keyStore       *kms.KMS
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	conn := lookupPostgresURL()
	if conn == "" {
		conn = "postgres://postgres:postgres@localhost:5435"
	}

	cfg = config.Configuration{
		Database: config.Database{
			URL: conn,
		},
		KeyStore: config.KeyStore{
			Address:              "http://localhost:8300",
			Token:                "hvs.TUxhxUvzCcviQSfRsdcwuWWA",
			PluginIden3MountPath: "iden3",
		},
	}
	s, teardown, err := tests.NewTestStorage(&cfg)
	defer teardown()
	if err != nil {
		log.Println("failed to acquire test database")
		return 1
	}
	storage = s

	vaultCli, err = providers.NewVaultClient(cfg.KeyStore.Address, cfg.KeyStore.Token)
	if err != nil {
		log.Println("failed to acquire vault client")
		return 1
	}

	bjjKeyProvider, err = kms.NewVaultPluginIden3KeyProvider(vaultCli, cfg.KeyStore.PluginIden3MountPath, kms.KeyTypeBabyJubJub)
	if err != nil {
		log.Println("failed to create Iden3 Key Provider")
		return 1
	}

	keyStore = kms.NewKMS()
	err = keyStore.RegisterKeyProvider(kms.KeyTypeBabyJubJub, bjjKeyProvider)
	if err != nil {
		log.Println("failed to register Key Provider")
		return 1
	}

	return m.Run()
}

func getHandler(ctx context.Context, server *Server) http.Handler {
	mux := chi.NewRouter()
	RegisterStatic(mux)
	return HandlerFromMux(NewStrictHandler(server, middlewares(ctx)), mux)
}

func middlewares(ctx context.Context) []StrictMiddlewareFunc {
	return []StrictMiddlewareFunc{
		LogMiddleware(ctx),
	}
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
