package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/polygonid/sh-id-platform/internal/api"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func main() {
	cfg, err := config.Load("internal/config")
	if err != nil {
		log.Fatal(err)
	}

	storage, err := db.NewStorage(cfg.Database.Url)
	if err != nil {
		log.Fatal(err)
	}

	vaultCli, err := providers.NewVaultClient(cfg.KeyStore.Address, cfg.KeyStore.Token)
	if err != nil {
		log.Fatal("cannot init vault client: ", err)
	}

	bjjKeyProvider, err := kms.NewVaultPluginIden3KeyProvider(vaultCli, cfg.KeyStore.PluginIden3MountPath, kms.KeyTypeBabyJubJub)
	if err != nil {
		log.Errorf("cannot create BabyJubJub key provider: %+v", err)
	}

	keyStore := kms.NewKMS()
	err = keyStore.RegisterKeyProvider(kms.KeyTypeBabyJubJub, bjjKeyProvider)
	if err != nil {
		log.Errorf("cannot register BabyJubJub key provider: %+v", err)
	}

	identityRepo := repositories.NewIdentity(storage.Pgx)
	claimsRepo := repositories.NewClaims(storage.Pgx)
	identityStateRepo := repositories.NewIdentityState(storage.Pgx)
	mtRepo := repositories.NewIdentityMerkleTreeRepository(storage.Pgx)
	mtService := services.NewIdentityMerkleTrees(mtRepo)

	service := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, storage)

	mux := echo.New()
	api.RegisterStatic(mux)
	api.RegisterHandlers(mux, api.NewStrictHandler(api.NewServer(service), nil))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: mux,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Error starting http server: %w", err)
		}
	}()

	<-quit
	log.Info("Shutting down")
}
