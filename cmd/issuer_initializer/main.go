package main

import (
	"context"
	"fmt"
	"os"

	"github.com/polygonid/sh-id-platform/internal/buildinfo"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
)

var build = buildinfo.Revision()

func main() {
	log.Info(context.Background(), "starting issuer node...", "revision", build)
	cfg, err := config.Load("")
	if err != nil {
		log.Error(context.Background(), "cannot load config", "err", err)
		return
	}

	ctx, cancel := context.WithCancel(log.NewContext(context.Background(), cfg.Log.Level, cfg.Log.Mode, os.Stdout))
	defer cancel()

	if err := cfg.Sanitize(ctx); err != nil {
		log.Error(ctx, "there are errors in the configuration that prevent server to start", "err", err)
		return
	}

	storage, err := db.NewStorage(cfg.Database.URL)
	if err != nil {
		log.Error(ctx, "cannot connect to database", "err", err)
		return
	}

	vaultCli, err := providers.VaultClient(ctx, providers.Config{
		UserPassAuthEnabled: cfg.VaultUserPassAuthEnabled,
		Address:             cfg.KeyStore.Address,
		Token:               cfg.KeyStore.Token,
		Pass:                cfg.VaultUserPassAuthPassword,
	})
	if err != nil {
		log.Error(ctx, "cannot initialize vault client", "err", err)
		return
	}

	did, err := providers.GetDID(ctx, vaultCli)
	if err != nil {
		log.Info(ctx, "did not found in vault, creating new one")
	}

	if did != "" {
		log.Info(ctx, "did already created, skipping", "did", did)
		return
	}

	keyStore, err := kms.Open(cfg.KeyStore.PluginIden3MountPath, vaultCli)
	if err != nil {
		log.Error(ctx, "cannot initialize kms", "err", err)
		return
	}

	// repositories initialization
	identityRepository := repositories.NewIdentity()
	claimsRepository := repositories.NewClaims()
	mtRepository := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepository := repositories.NewIdentityState()

	// services initialization
	mtService := services.NewIdentityMerkleTrees(mtRepository)
	identityService := services.NewIdentity(keyStore, identityRepository, mtRepository, identityStateRepository, mtService, nil, claimsRepository, nil, nil, storage, nil, nil, nil, pubsub.NewMock())

	identity, err := identityService.Create(ctx, cfg.APIUI.IdentityMethod, cfg.APIUI.IdentityBlockchain, cfg.APIUI.IdentityNetwork, cfg.ServerUrl)
	if err != nil {
		log.Error(ctx, "error creating identifier", err)
		return
	}

	log.Info(ctx, "identifier crated successfully")

	if err := providers.SaveDID(ctx, vaultCli, identity.Identifier); err != nil {
		log.Error(ctx, "error saving identifier to vault", err)
		return
	}

	//nolint:all
	fmt.Printf(identity.Identifier)
	//nolint:all
	fmt.Printf("\n")
}
