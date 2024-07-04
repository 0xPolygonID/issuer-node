package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	redis2 "github.com/go-redis/redis/v8"
	vault "github.com/hashicorp/vault/api"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	iden3commProtocol "github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/api"
	"github.com/polygonid/sh-id-platform/internal/buildinfo"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/errors"
	"github.com/polygonid/sh-id-platform/internal/gateways"
	"github.com/polygonid/sh-id-platform/internal/health"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/redis"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/cache"
	"github.com/polygonid/sh-id-platform/pkg/credentials/revocation_status"
	circuitLoaders "github.com/polygonid/sh-id-platform/pkg/loaders"
	"github.com/polygonid/sh-id-platform/pkg/network"
	"github.com/polygonid/sh-id-platform/pkg/protocol"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

var build = buildinfo.Revision()

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info(ctx, "starting issuer node...", "revision", build)

	cfg, err := config.Load("")
	if err != nil {
		log.Error(ctx, "cannot load config", "err", err)
		return
	}
	log.Config(cfg.Log.Level, cfg.Log.Mode, os.Stdout)

	if err := cfg.Sanitize(ctx); err != nil {
		log.Error(ctx, "there are errors in the configuration that prevent server to start", "err", err)
		return
	}

	storage, err := db.NewStorage(cfg.Database.URL)
	if err != nil {
		log.Error(ctx, "cannot connect to database", "err", err)
		return
	}

	// Redis cache
	rdb, err := redis.Open(cfg.Cache.RedisUrl)
	if err != nil {
		log.Error(ctx, "cannot connect to redis", "err", err, "host", cfg.Cache.RedisUrl)
		return
	}
	ps := pubsub.NewRedis(rdb)
	ps.WithLogger(log.Error)
	cachex := cache.NewRedisCache(rdb)

	// TODO: Cache only if cfg.APIUI.SchemaCache == true
	schemaLoader := loader.NewDocumentLoader(cfg.IPFS.GatewayURL)

	var vaultCli *vault.Client
	var vaultErr error
	vaultCfg := providers.Config{
		UserPassAuthEnabled: cfg.VaultUserPassAuthEnabled,
		Address:             cfg.KeyStore.Address,
		Token:               cfg.KeyStore.Token,
		Pass:                cfg.VaultUserPassAuthPassword,
	}

	vaultCli, vaultErr = providers.VaultClient(ctx, vaultCfg)
	if vaultErr != nil {
		log.Error(ctx, "cannot initialize vault client", "err", err)
		return
	}

	if vaultCfg.UserPassAuthEnabled {
		go providers.RenewToken(ctx, vaultCli, vaultCfg)
	}

	keyStore, err := kms.Open(cfg.KeyStore.PluginIden3MountPath, vaultCli)
	if err != nil {
		log.Error(ctx, "cannot initialize kms", "err", err)
		return
	}

	circuitsLoaderService := circuitLoaders.NewCircuits(cfg.Circuit.Path)

	cfg.CredentialStatus.SingleIssuer = false
	reader, err := network.ReadFile(ctx, cfg.NetworkResolverPath)
	if err != nil {
		log.Error(ctx, "cannot read network resolver file", "err", err)
		return
	}
	networkResolver, err := network.NewResolver(ctx, *cfg, keyStore, reader)
	if err != nil {
		log.Error(ctx, "failed initialize network resolver", "err", err)
		return
	}

	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	// repositories initialization
	identityRepository := repositories.NewIdentity()
	claimsRepository := repositories.NewClaims()
	mtRepository := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepository := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()

	// services initialization
	mtService := services.NewIdentityMerkleTrees(mtRepository)
	qrService := services.NewQrStoreService(cachex)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			iden3commProtocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			iden3commProtocol.RevocationStatusRequestMessageType: {"*"},
		},
		*cfg.MediaTypeManager.Enabled,
	)

	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	identityService := services.NewIdentity(keyStore, identityRepository, mtRepository, identityStateRepository, mtService, qrService, claimsRepository, revocationRepository, nil, storage, nil, nil, ps, *networkResolver, rhsFactory, revocationStatusResolver)
	claimsService := services.NewClaim(claimsRepository, identityService, qrService, mtService, identityStateRepository, schemaLoader, storage, cfg.ServerUrl, ps, cfg.IPFS.GatewayURL, revocationStatusResolver, mediaTypeManager)
	proofService := gateways.NewProver(ctx, cfg, circuitsLoaderService)

	transactionService, err := gateways.NewTransaction(*networkResolver)
	if err != nil {
		log.Error(ctx, "error creating transaction service", "err", err)
		return
	}
	accountService := services.NewAccountService(*networkResolver)

	publisherGateway, err := gateways.NewPublisherEthGateway(*networkResolver, keyStore, cfg.PublishingKeyPath)
	if err != nil {
		log.Error(ctx, "error creating publish gateway", "err", err)
		return
	}

	publisher := gateways.NewPublisher(storage, identityService, claimsService, mtService, keyStore, transactionService, proofService, publisherGateway, networkResolver, ps)
	packageManager, err := protocol.InitPackageManager(ctx, networkResolver.GetSupportedContracts(), cfg.Circuit.Path)
	if err != nil {
		log.Error(ctx, "failed init package protocol", "err", err)
		return
	}

	serverHealth := health.New(health.Monitors{
		"postgres": storage.Ping,
		"redis": func(rdb *redis2.Client) health.Pinger {
			return func(ctx context.Context) error { return rdb.Ping(ctx).Err() }
		}(rdb),
	})
	serverHealth.Run(ctx, health.DefaultPingPeriod)

	mux := chi.NewRouter()
	mux.Use(
		chiMiddleware.RequestID,
		log.ChiMiddleware(ctx),
		chiMiddleware.Recoverer,
		cors.Handler(cors.Options{AllowedOrigins: []string{"*"}}),
		chiMiddleware.NoCache,
	)
	api.HandlerFromMux(
		api.NewStrictHandlerWithOptions(
			api.NewServer(cfg, identityService, accountService, claimsService, qrService, publisher, packageManager, *networkResolver, serverHealth),
			middlewares(ctx, cfg.HTTPBasicAuth),
			api.StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  errors.RequestErrorHandlerFunc,
				ResponseErrorHandlerFunc: errors.ResponseErrorHandlerFunc,
			}),
		mux)
	api.RegisterStatic(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: mux,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info(ctx, "server started", "port", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil {
			log.Error(ctx, "starting http server", "err", err)
		}
	}()

	<-quit
	log.Info(ctx, "Shutting down")
}

func middlewares(ctx context.Context, auth config.HTTPBasicAuth) []api.StrictMiddlewareFunc {
	return []api.StrictMiddlewareFunc{
		api.LogMiddleware(ctx),
		api.BasicAuthMiddleware(ctx, auth.User, auth.Password),
	}
}
