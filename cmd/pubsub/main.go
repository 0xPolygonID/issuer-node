package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	goredis "github.com/go-redis/redis/v8"

	"github.com/polygonid/sh-id-platform/internal/config"
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

	ctx := log.NewContext(context.Background(), cfg.Log.Level, cfg.Log.Mode, os.Stdout)

	if err := cfg.SanitizeAdmin(); err != nil {
		log.Error(ctx, "there are errors in the configuration that prevent server to start", "err", err)
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

	redisPubSub := pubsub.NewRedis(rdb)

	connectionsRepository := repositories.NewConnections()

	connectionsService := services.NewConnection(connectionsRepository, storage)
	credentialsService := newCredentialsService(ctx, cfg, storage, rdb)
	notificationGateway := gateways.NewPushNotificationClient(http.DefaultHTTPClientWithRetry)
	notificationService := services.NewNotification(cfg.APIUI.IssuerDID, notificationGateway, connectionsService, credentialsService)

	redisPubSub.Subscribe(ctx, pubsub.EventCreateCredential, notificationService.SendCreateCredentialNotification)

	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)
	<-gracefulShutdown
}

func newCredentialsService(ctx context.Context, cfg *config.Configuration, storage *db.Storage, rdb *goredis.Client) ports.ClaimsService {
	cachex := cache.NewRedisCache(rdb)

	vaultCli, err := providers.NewVaultClient(cfg.KeyStore.Address, cfg.KeyStore.Token)
	if err != nil {
		log.Error(ctx, "cannot init vault client: ", "err", err)
		panic(err)
	}
	identityRepository := repositories.NewIdentity()
	claimsRepository := repositories.NewClaims()
	mtRepository := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepository := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	keyStore, err := kms.Open(cfg.KeyStore.PluginIden3MountPath, vaultCli)
	if err != nil {
		log.Error(ctx, "cannot initialize kms", "err", err)
		panic(err)
	}
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	var schemaLoader loader.Factory
	if cfg.SchemaCache == nil || !*cfg.SchemaCache {
		schemaLoader = loader.HTTPFactory
	} else {
		schemaLoader = loader.CachedFactory(loader.HTTPFactory, cachex)
	}
	mtService := services.NewIdentityMerkleTrees(mtRepository)
	identityService := services.NewIdentity(keyStore, identityRepository, mtRepository, identityStateRepository, mtService, claimsRepository, revocationRepository, nil, storage, rhsp, nil, nil)
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
	)

	return claimsService
}
