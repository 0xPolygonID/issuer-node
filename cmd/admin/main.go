package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	redis2 "github.com/go-redis/redis/v8"
	auth "github.com/iden3/go-iden3-auth"
	authLoaders "github.com/iden3/go-iden3-auth/loaders"
	"github.com/iden3/go-iden3-auth/pubsignals"
	"github.com/iden3/go-iden3-auth/state"

	"github.com/polygonid/sh-id-platform/internal/api_admin"
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
	"github.com/polygonid/sh-id-platform/internal/providers/blockchain"
	"github.com/polygonid/sh-id-platform/internal/redis"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/cache"
	"github.com/polygonid/sh-id-platform/pkg/loaders"
	"github.com/polygonid/sh-id-platform/pkg/protocol"
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
	cachex := cache.NewRedisCache(rdb)
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)

	vaultCli, err := providers.NewVaultClient(cfg.KeyStore.Address, cfg.KeyStore.Token)
	if err != nil {
		log.Error(ctx, "cannot init vault client: ", "err", err)
		return
	}

	keyStore, err := kms.Open(cfg.KeyStore.PluginIden3MountPath, vaultCli)
	if err != nil {
		log.Error(ctx, "cannot initialize kms", "err", err)
		return
	}

	ethereumClient, err := blockchain.Open(cfg)
	if err != nil {
		log.Error(ctx, "error dialing with ethereum client", "err", err)
		return
	}

	stateContract, err := blockchain.InitEthClient(cfg.Ethereum.URL, cfg.Ethereum.ContractAddress)
	if err != nil {
		log.Error(ctx, "failed init ethereum client", "err", err)
		return
	}

	ethConn, err := blockchain.InitEthConnect(cfg.Ethereum)
	if err != nil {
		log.Error(ctx, "failed init ethereum connect", "err", err)
		return
	}

	verificationKeyLoader := &authLoaders.FSKeyLoader{Dir: cfg.Circuit.Path + "/authV2"}
	resolvers := map[string]pubsignals.StateResolver{
		cfg.Ethereum.ResolverPrefix: state.ETHResolver{
			RPCUrl:          cfg.Ethereum.URL,
			ContractAddress: common.HexToAddress(cfg.Ethereum.ContractAddress),
		},
	}

	verifier := auth.NewVerifier(verificationKeyLoader, authLoaders.DefaultSchemaLoader{IpfsURL: "ipfs.io"}, resolvers)

	circuitsLoaderService := loaders.NewCircuits(cfg.Circuit.Path)

	rhsp := reverse_hash.NewRhsPublisher(nil, false)

	// repositories initialization
	identityRepository := repositories.NewIdentity()
	claimsRepository := repositories.NewClaims()
	mtRepository := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepository := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	sessionRepository := repositories.NewSessionCached(cachex)

	// services initialization
	mtService := services.NewIdentityMerkleTrees(mtRepository)
	identityService := services.NewIdentity(keyStore, identityRepository, mtRepository, identityStateRepository, mtService, claimsRepository, revocationRepository, connectionsRepository, storage, rhsp, verifier, sessionRepository)
	schemaService := services.NewSchema(schemaLoader)
	schemaAdminService := services.NewSchemaAdmin(repositories.NewSchema(storage.Pgx))
	claimsService := services.NewClaim(
		claimsRepository,
		schemaService,
		identityService,
		mtService,
		identityStateRepository,
		storage,
		services.ClaimCfg{
			RHSEnabled: cfg.ReverseHashService.Enabled,
			RHSUrl:     cfg.ReverseHashService.URL,
			Host:       cfg.ServerUrl,
		},
	)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	proofService := gateways.NewProver(ctx, cfg, circuitsLoaderService)
	revocationService := services.NewRevocationService(ethConn, common.HexToAddress(cfg.Ethereum.ContractAddress))
	zkProofService := services.NewProofService(claimsService, revocationService, identityService, mtService, claimsRepository, keyStore, storage, stateContract, schemaLoader)
	transactionService, err := gateways.NewTransaction(ethereumClient, cfg.Ethereum.ConfirmationBlockCount)
	if err != nil {
		log.Error(ctx, "error creating transaction service", "err", err)
		return
	}

	publisherGateway, err := gateways.NewPublisherEthGateway(ethereumClient, common.HexToAddress(cfg.Ethereum.ContractAddress), keyStore, cfg.PublishingKeyPath)
	if err != nil {
		log.Error(ctx, "error creating publish gateway", "err", err)
		return
	}

	publisher := gateways.NewPublisher(storage, identityService, claimsService, mtService, keyStore, transactionService, proofService, publisherGateway, cfg.Ethereum.ConfirmationTimeout)

	packageManager, err := protocol.InitPackageManager(ctx, stateContract, zkProofService, cfg.Circuit.Path)
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
	api_admin.HandlerFromMux(
		api_admin.NewStrictHandlerWithOptions(
			api_admin.NewServer(cfg, identityService, claimsService, schemaAdminService, connectionsService, publisher, packageManager, serverHealth),
			middlewares(ctx, cfg.APIUI.APIUIAuth),
			api_admin.StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  errors.RequestErrorHandlerFunc,
				ResponseErrorHandlerFunc: errors.ResponseErrorHandlerFunc,
			}),
		mux)
	api_admin.RegisterStatic(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.APIUI.ServerPort),
		Handler: mux,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info(ctx, "UI API server started", "port", cfg.APIUI.ServerPort)
		if err := server.ListenAndServe(); err != nil {
			log.Error(ctx, "starting HTTP UI API server", "err", err)
		}
	}()

	<-quit
	log.Info(ctx, "Shutting down")
}

func middlewares(ctx context.Context, auth config.APIUIAuth) []api_admin.StrictMiddlewareFunc {
	return []api_admin.StrictMiddlewareFunc{
		api_admin.LogMiddleware(ctx),
		api_admin.BasicAuthMiddleware(ctx, auth.User, auth.Password),
	}
}
