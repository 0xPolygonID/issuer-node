package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/event"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/gateways"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/redis"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/cache"
	"github.com/polygonid/sh-id-platform/pkg/http"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Error(context.Background(), "cannot load config", "err", err)
		return
	}

	ctx, cancel := context.WithCancel(log.NewContext(context.Background(), cfg.Log.Level, cfg.Log.Mode, os.Stdout))
	defer cancel()

	if err := cfg.SanitizeAPIUI(ctx); err != nil {
		log.Error(ctx, "there are errors in the configuration that prevent server to start", "err", err)
		return
	}

	if cfg.APIUI.Issuer == "" {
		log.Error(ctx, "issuer DID is not set")
		return
	}

	rdb, err := redis.Open(cfg.Cache.RedisUrl)
	if err != nil {
		log.Error(ctx, "cannot connect to redis", "err", err, "host", cfg.Cache.RedisUrl)
		return
	}

	storage, err := db.NewStorage(cfg.Database.URL)
	if err != nil {
		log.Error(ctx, "cannot connect to database", "err", err)
		return
	}

	ps := pubsub.NewRedis(rdb)
	ps.WithLogger(log.Error)
	cachex := cache.NewRedisCache(rdb)

	connectionsRepository := repositories.NewConnections()

	connectionsService := services.NewConnection(connectionsRepository, storage)
	credentialsService, err := newCredentialsService(cfg, storage, cachex, ps)
	if err != nil {
		log.Error(ctx, "cannot initialize the credential service", "err", err)
		return
	}

	notificationGateway := gateways.NewPushNotificationClient(http.DefaultHTTPClientWithRetry)
	notificationService := services.NewNotification(notificationGateway, connectionsService, credentialsService)
	ctxCancel, cancel := context.WithCancel(ctx)
	defer func() {
		log.Info(ctx, "Shutting down...")
		cancel()
		if err := rdb.Close(); err != nil {
			log.Error(ctx, "closing redis connection", "err", err)
		}
	}()

	ps.Subscribe(ctxCancel, event.CreateCredentialEvent, notificationService.SendCreateCredentialNotification)
	ps.Subscribe(ctxCancel, event.CreateConnectionEvent, notificationService.SendCreateConnectionNotification)

	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)

	<-gracefulShutdown
}

func newCredentialsService(cfg *config.Configuration, storage *db.Storage, cachex cache.Cache, ps pubsub.Client) (ports.ClaimsService, error) {
	vaultCli, err := providers.NewVaultClient(cfg.KeyStore.Address, cfg.KeyStore.Token)
	if err != nil {
		return nil, fmt.Errorf("cannot init vault client: err %s", err.Error())
	}

	identityRepository := repositories.NewIdentity()
	claimsRepository := repositories.NewClaims()
	mtRepository := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepository := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	keyStore, err := kms.Open(cfg.KeyStore.PluginIden3MountPath, vaultCli)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize kms: err %s", err.Error())
	}

	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	var schemaLoader loader.Factory
	if cfg.SchemaCache == nil || !*cfg.SchemaCache {
		schemaLoader = loader.HTTPFactory
	} else {
		schemaLoader = loader.CachedFactory(loader.HTTPFactory, cachex)
	}

	mtService := services.NewIdentityMerkleTrees(mtRepository)
	identityService := services.NewIdentity(keyStore, identityRepository, mtRepository, identityStateRepository, mtService, claimsRepository, revocationRepository, nil, storage, rhsp, nil, nil, ps)
	claimsService := services.NewClaim(
		claimsRepository,
		identityService,
		mtService,
		identityStateRepository,
		schemaLoader,
		storage,
		services.ClaimCfg{
			RHSEnabled: cfg.ReverseHashService.Enabled,
			RHSUrl:     cfg.ReverseHashService.URL,
			Host:       cfg.ServerUrl,
		},
		ps,
		cfg.IFPS.GatewayURL,
	)

	return claimsService, nil
}
