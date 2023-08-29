package main

import (
	"context"
	"encoding/json"
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
	core "github.com/iden3/go-iden3-core"
	proof "github.com/iden3/merkletree-proof"

	"github.com/polygonid/sh-id-platform/internal/api_ui"
	"github.com/polygonid/sh-id-platform/internal/buildinfo"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
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
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
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

	if err := cfg.SanitizeAPIUI(ctx); err != nil {
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

	var schemaLoader loader.Factory
	schemaLoader = loader.MultiProtocolFactory(cfg.IFPS.GatewayURL)
	if cfg.APIUI.SchemaCache != nil && *cfg.APIUI.SchemaCache {
		schemaLoader = loader.CachedFactory(schemaLoader, cachex)
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

	keyStore, err := kms.Open(cfg.KeyStore.PluginIden3MountPath, vaultCli)
	if err != nil {
		log.Error(ctx, "cannot initialize kms", "err", err)
		return
	}

	err = config.CheckDID(ctx, cfg, vaultCli)
	if err != nil {
		log.Error(ctx, "cannot initialize did", "err", err)
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

	verifier, err := auth.NewVerifierWithExplicitError(verificationKeyLoader, authLoaders.DefaultSchemaLoader{IpfsURL: cfg.IFPS.GatewayURL}, resolvers)
	if err != nil {
		log.Error(ctx, "failed init verifier", "err", err)
		return
	}

	circuitsLoaderService := loaders.NewCircuits(cfg.Circuit.Path)

	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	if cfg.ReverseHashService.Enabled {
		rhsp = reverse_hash.NewRhsPublisher(&proof.HTTPReverseHashCli{
			URL:         cfg.ReverseHashService.URL,
			HTTPTimeout: reverse_hash.DefaultRHSTimeOut,
		}, true)
	}

	// repositories initialization
	identityRepository := repositories.NewIdentity()
	claimsRepository := repositories.NewClaims()
	mtRepository := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepository := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	sessionRepository := repositories.NewSessionCached(cachex)
	linkRepository := repositories.NewLink(*storage)
	schemaRepository := repositories.NewSchema(*storage)

	// services initialization
	mtService := services.NewIdentityMerkleTrees(mtRepository)
	qrService := services.NewQrStoreService(cachex)
	identityService := services.NewIdentity(keyStore, identityRepository, mtRepository, identityStateRepository, mtService, qrService, claimsRepository, revocationRepository, connectionsRepository, storage, rhsp, verifier, sessionRepository, ps)
	schemaService := services.NewSchema(schemaRepository, schemaLoader)
	claimsService := services.NewClaim(claimsRepository, identityService, qrService, mtService, identityStateRepository, schemaLoader, storage, services.ClaimCfg{
		RHSEnabled: cfg.ReverseHashService.Enabled,
		RHSUrl:     cfg.ReverseHashService.URL,
		Host:       cfg.APIUI.ServerURL,
	}, ps, cfg.IFPS.GatewayURL)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	linkService := services.NewLinkService(storage, claimsService, qrService, claimsRepository, linkRepository, schemaRepository, schemaLoader, sessionRepository, ps, cfg.IFPS.GatewayURL)
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

	publisher := gateways.NewPublisher(storage, identityService, claimsService, mtService, keyStore, transactionService, proofService, publisherGateway, cfg.Ethereum.ConfirmationTimeout, ps)

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

	if !identifierExists(ctx, &cfg.APIUI.IssuerDID, identityService) {
		log.Error(ctx, "issuer DID must exist")
		return
	}

	mux := chi.NewRouter()
	mux.Use(
		chiMiddleware.RequestID,
		log.ChiMiddleware(ctx),
		chiMiddleware.Recoverer,
		cors.AllowAll().Handler,
		chiMiddleware.NoCache,
	)
	api_ui.HandlerWithOptions(
		api_ui.NewStrictHandlerWithOptions(
			api_ui.NewServer(cfg, identityService, claimsService, schemaService, connectionsService, linkService, qrService, publisher, packageManager, serverHealth),
			middlewares(ctx, cfg.APIUI.APIUIAuth),
			api_ui.StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  errors.RequestErrorHandlerFunc,
				ResponseErrorHandlerFunc: errors.ResponseErrorHandlerFunc,
			}),
		api_ui.ChiServerOptions{
			BaseRouter:       mux,
			ErrorHandlerFunc: errorHandlerFunc,
		},
	)
	api_ui.RegisterStatic(mux)

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

func identifierExists(ctx context.Context, did *core.DID, service ports.IdentityService) bool {
	_, err := service.GetByDID(ctx, *did)
	return err == nil
}

func middlewares(ctx context.Context, auth config.APIUIAuth) []api_ui.StrictMiddlewareFunc {
	return []api_ui.StrictMiddlewareFunc{
		api_ui.LogMiddleware(ctx),
		api_ui.BasicAuthMiddleware(ctx, auth.User, auth.Password),
	}
}

func errorHandlerFunc(w http.ResponseWriter, _ *http.Request, err error) {
	switch err.(type) {
	case *api_ui.InvalidParamFormatError:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"message": err.Error()})
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
