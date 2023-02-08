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

	"github.com/polygonid/sh-id-platform/internal/api"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
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
		log.Error(context.Background(), "cannot load config", err)
		panic(err)
	}

	// Context with log
	ctx := log.NewContext(context.Background(), cfg.Log.Level, cfg.Log.Mode, os.Stdout)

	storage, err := db.NewStorage(cfg.Database.URL)
	if err != nil {
		log.Error(ctx, "cannot connect to database", err)
		panic(err)
	}

	// Redis cache
	rdb, err := redis.Open(cfg.Cache.RedisUrl)
	if err != nil {
		log.Error(ctx, "cannot connect to redis", err, "host", cfg.Cache.RedisUrl)
		panic(err)
	}
	cachex := cache.NewRedisCache(rdb)
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)

	vaultCli, err := providers.NewVaultClient(cfg.KeyStore.Address, cfg.KeyStore.Token)
	if err != nil {
		log.Error(ctx, "cannot init vault client: ", err)
		panic(err)
	}

	keyStore, err := kms.Open(cfg.KeyStore.PluginIden3MountPath, vaultCli)
	if err != nil {
		log.Error(ctx, "cannot initialize kms: %+v", err)
		panic(err)
	}

	ethereumClient, err := blockchain.Open(cfg)
	if err != nil {
		panic("Error dialing with ethclient: " + err.Error())
	}

	stateContract, err := blockchain.InitEthClient(cfg.Ethereum.URL, cfg.Ethereum.ContractAddress)
	if err != nil {
		log.Error(ctx, "failed init ethereum client: %+v", err)
		panic(err)
	}

	ethConn, err := blockchain.InitEthConnect(cfg.Ethereum)
	if err != nil {
		log.Error(ctx, "failed init ethereum connect: %+v", err)
		panic(err)
	}

	circuitsLoaderService := loaders.NewCircuits(cfg.Circuit.Path)

	rhsp := reverse_hash.NewRhsPublisher(nil, false)

	// repositories initialization
	identityRepository := repositories.NewIdentity()
	claimsRepository := repositories.NewClaims()
	mtRepository := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepository := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()

	// services initialization
	mtService := services.NewIdentityMerkleTrees(mtRepository)
	identityService := services.NewIdentity(keyStore, identityRepository, mtRepository, identityStateRepository, mtService, claimsRepository, revocationRepository, storage, rhsp)
	schemaService := services.NewSchema(schemaLoader)
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
	proofService := gateways.NewProver(ctx, cfg, circuitsLoaderService)
	revocationService := services.NewRevocationService(ethConn, common.HexToAddress(cfg.Ethereum.ContractAddress))
	zkProofService := services.NewProofService(claimsService, revocationService, identityService, mtService, claimsRepository, keyStore, storage, stateContract, schemaLoader)
	transactionService, err := gateways.NewTransaction(ethereumClient, cfg.Ethereum.ConfirmationBlockCount)
	if err != nil {
		log.Error(ctx, "error creating transaction service", err)
		panic("error creating transaction service")
	}

	publisherGateway, err := gateways.NewPublisherEthGateway(ethereumClient, common.HexToAddress(cfg.Ethereum.ContractAddress), keyStore, cfg.PublishingKeyPath)
	if err != nil {
		log.Error(ctx, "error creating publish gateway", err)
		panic("error creating publish gateway")
	}

	publisher := gateways.NewPublisher(storage, identityService, claimsService, mtService, keyStore, transactionService, proofService, publisherGateway, cfg.Ethereum.ConfirmationTimeout)

	packageManager, err := protocol.InitPackageManager(ctx, stateContract, zkProofService, cfg.Circuit.Path)
	if err != nil {
		log.Error(ctx, "failed init package protocol:  %+v", err)
		panic(err)
	}

	serverHealth := health.New(storage.Pgx, cachex)

	spec, err := api.GetSwagger()
	if err != nil {
		log.Error(ctx, "cannot retrieve the openapi specification file: %+v", err)
		os.Exit(1)
	}

	spec.Servers = nil
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
			api.NewServer(cfg, identityService, claimsService, schemaService, publisher, packageManager, serverHealth),
			middlewares(ctx, cfg.HTTPBasicAuth),
			api.StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  api.RequestErrorHandlerFunc,
				ResponseErrorHandlerFunc: api.ResponseErrorHandlerFunc,
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
		log.Info(ctx, fmt.Sprintf("server started on port:%d", cfg.ServerPort))
		if err := server.ListenAndServe(); err != nil {
			log.Error(ctx, "Starting http server", err)
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
