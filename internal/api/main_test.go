package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/hashicorp/vault/api"

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
	cfg = config.Configuration{
		Database: config.Database{
			URL: "postgres://postgres:postgres@localhost:5435",
		},
		KeyStore: config.KeyStore{
			Address:              "http://localhost:8300",
			Token:                "hvs.YxU2dLZljGpqLyPYu6VeYJde",
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
