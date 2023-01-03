package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/polygonid/sh-id-platform/internal/api"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Error(context.Background(), "cannot load config", err)
	}

	// Context with log
	ctx := log.NewContext(context.Background(), cfg.Runtime.LogLevel, cfg.Runtime.LogMode, os.Stdout)

	storage, err := db.NewStorage(cfg.Database.URL)
	if err != nil {
		log.Error(context.Background(), "cannot connect to database", err)
	}

	vaultCli, err := providers.NewVaultClient(cfg.KeyStore.Address, cfg.KeyStore.Token)
	if err != nil {
		log.Error(context.Background(), "cannot init vault client: ", err)
	}

	bjjKeyProvider, err := kms.NewVaultPluginIden3KeyProvider(vaultCli, cfg.KeyStore.PluginIden3MountPath, kms.KeyTypeBabyJubJub)
	if err != nil {
		log.Error(ctx, "cannot create BabyJubJub key provider: %+v", err)
	}

	keyStore := kms.NewKMS()
	err = keyStore.RegisterKeyProvider(kms.KeyTypeBabyJubJub, bjjKeyProvider)
	if err != nil {
		log.Error(ctx, "cannot register BabyJubJub key provider: %+v", err)
	}

	identityRepo := repositories.NewIdentity(storage.Pgx)
	claimsRepo := repositories.NewClaims(storage.Pgx)
	identityStateRepo := repositories.NewIdentityState(storage.Pgx)
	mtRepo := repositories.NewIdentityMerkleTreeRepository(storage.Pgx)
	mtService := services.NewIdentityMerkleTrees(mtRepo)

	service := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, storage)

	mux := chi.NewRouter()
	api.HandlerFromMux(api.NewStrictHandler(api.NewServer(cfg, service), middlewares(ctx)), mux)
	api.RegisterStatic(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: mux,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info(ctx, fmt.Sprintf("server started on port:%d", cfg.ServerPort))
		if err := server.ListenAndServe(); err != nil {
			log.Error(ctx, "Starting http server", err)
		}
	}()

	<-quit
	log.Info(ctx, "Shutting down")
}

func middlewares(ctx context.Context) []api.StrictMiddlewareFunc {
	return []api.StrictMiddlewareFunc{
		api.LogMiddleware(ctx),
	}
}
