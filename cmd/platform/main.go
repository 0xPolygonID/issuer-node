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
	auth "github.com/iden3/go-iden3-auth/v2"
	authLoaders "github.com/iden3/go-iden3-auth/v2/loaders"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	iden3commProtocol "github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/api"
	"github.com/polygonid/sh-id-platform/internal/buildinfo"
	"github.com/polygonid/sh-id-platform/internal/cache"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/errors"
	"github.com/polygonid/sh-id-platform/internal/gateways"
	"github.com/polygonid/sh-id-platform/internal/health"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/network"
	"github.com/polygonid/sh-id-platform/internal/packagemanager"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/pubsub"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/internal/reversehash"
	"github.com/polygonid/sh-id-platform/internal/revocationstatus"
	circuitLoaders "github.com/polygonid/sh-id-platform/pkg/loaders"
)

var build = buildinfo.Revision()

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info(ctx, "starting issuer node...", "revision: ", build)

	cfg, err := config.Load()
	if err != nil {
		log.Error(ctx, "cannot load config", "err", err)
		return
	}
	log.Config(cfg.Log.Level, cfg.Log.Mode, os.Stdout)

	storage, err := db.NewStorage(cfg.Database.URL)
	if err != nil {
		log.Error(ctx, "cannot connect to database", "err", err)
		return
	}

	cachex, err := cache.NewCacheClient(ctx, *cfg)
	if err != nil {
		log.Error(ctx, "cannot initialize cache", "err", err)
		return
	}
	ps, err := pubsub.NewPubSub(ctx, *cfg)
	if err != nil {
		log.Error(ctx, "cannot initialize pubsub", "err", err)
		return
	}

	// TODO: Cache only if cfg.APIUI.SchemaCache == true
	schemaLoader := loader.NewDocumentLoader(cfg.IPFS.GatewayURL, cfg.SchemaCache)

	vaultCfg := providers.Config{
		UserPassAuthEnabled: cfg.KeyStore.VaultUserPassAuthEnabled,
		Pass:                cfg.KeyStore.VaultUserPassAuthPassword,
		Address:             cfg.KeyStore.Address,
		Token:               cfg.KeyStore.Token,
		TLSEnabled:          cfg.KeyStore.TLSEnabled,
		CertPath:            cfg.KeyStore.CertPath,
	}

	keyStore, err := config.KeyStoreConfig(ctx, cfg, vaultCfg)
	if err != nil {
		log.Error(ctx, "cannot initialize key store", "err", err)
		return
	}

	circuitsLoaderService := circuitLoaders.NewCircuits(cfg.Circuit.Path)

	reader, err := network.GetReaderFromConfig(cfg, ctx)
	if err != nil {
		log.Error(ctx, "cannot read network resolver file", "err", err)
		return
	}
	networkResolver, err := network.NewResolver(ctx, *cfg, keyStore, reader)
	if err != nil {
		log.Error(ctx, "failed initialize network resolver", "err", err)
		return
	}

	rhsFactory := reversehash.NewFactory(*networkResolver, reversehash.DefaultRHSTimeOut)
	// repositories initialization
	identityRepository := repositories.NewIdentity()
	claimsRepository := repositories.NewClaim()
	connectionsRepository := repositories.NewConnection()
	mtRepository := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepository := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	schemaRepository := repositories.NewSchema(*storage)
	linkRepository := repositories.NewLink(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)

	// services initialization
	mtService := services.NewIdentityMerkleTrees(mtRepository)
	qrService := services.NewQrStoreService(cachex)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepository, storage)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			iden3commProtocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			iden3commProtocol.RevocationStatusRequestMessageType: {"*"},
			iden3commProtocol.DiscoverFeatureQueriesMessageType:  {"*"},
		},
		*cfg.MediaTypeManager.Enabled,
	)

	universalDIDResolverUrl := auth.UniversalResolverURL
	if cfg.UniversalDIDResolver.UniversalResolverURL != nil && *cfg.UniversalDIDResolver.UniversalResolverURL != "" {
		universalDIDResolverUrl = *cfg.UniversalDIDResolver.UniversalResolverURL
	}
	universalDIDResolverHandler := packagemanager.NewUniversalDIDResolverHandler(universalDIDResolverUrl)

	packageManager, err := packagemanager.New(ctx, networkResolver.GetSupportedContracts(), cfg.Circuit.Path, universalDIDResolverHandler)
	if err != nil {
		log.Error(ctx, "failed init package packagemanager", "err", err)
		return
	}

	verificationKeyLoader := &authLoaders.FSKeyLoader{Dir: cfg.Circuit.Path + "/authV2"}
	verifier, err := auth.NewVerifier(verificationKeyLoader, networkResolver.GetStateResolvers(), auth.WithDIDResolver(universalDIDResolverHandler))
	if err != nil {
		log.Error(ctx, "failed init verifier", "err", err)
		return
	}

	revocationStatusResolver := revocationstatus.NewRevocationStatusResolver(*networkResolver)
	identityService := services.NewIdentity(keyStore, identityRepository, mtRepository, identityStateRepository, mtService, qrService, claimsRepository, revocationRepository, connectionsRepository, storage, verifier, sessionRepository, ps, *networkResolver, rhsFactory, revocationStatusResolver)
	claimsService := services.NewClaim(claimsRepository, identityService, qrService, mtService, identityStateRepository, schemaLoader, storage, cfg.ServerUrl, ps, cfg.IPFS.GatewayURL, revocationStatusResolver, mediaTypeManager, cfg.UniversalLinks)
	proofService := services.NewProver(circuitsLoaderService)
	schemaService := services.NewSchema(schemaRepository, schemaLoader)
	linkService := services.NewLinkService(storage, claimsService, qrService, claimsRepository, linkRepository, schemaRepository, schemaLoader, sessionRepository, ps, identityService, *networkResolver, cfg.UniversalLinks)
	discoveryService := services.NewDiscovery(mediaTypeManager, packageManager)

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

	serverHealth := health.New(health.Monitors{
		"postgres": storage.Ping,
		//"redis": func(rdb *redis2.Client) health.Pinger {
		//	return func(ctx context.Context) error { return rdb.Ping(ctx).Err() }
		//}(rdb),
	})
	serverHealth.Run(ctx, health.DefaultPingPeriod)

	mux := chi.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"localhost", "127.0.0.1", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	mux.Use(
		chiMiddleware.RequestID,
		log.ChiMiddleware(ctx),
		chiMiddleware.Recoverer,
		corsMiddleware.Handler,
		chiMiddleware.NoCache,
	)
	api.HandlerWithOptions(
		api.NewStrictHandlerWithOptions(
			api.NewServer(cfg, identityService, accountService, connectionsService, claimsService, qrService, publisher, packageManager, *networkResolver, serverHealth, schemaService, linkService, discoveryService),
			middlewares(ctx, cfg.HTTPBasicAuth),
			api.StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  errors.RequestErrorHandlerFunc,
				ResponseErrorHandlerFunc: errors.ResponseErrorHandlerFunc,
			}),
		api.ChiServerOptions{
			BaseRouter:       mux,
			ErrorHandlerFunc: api.ErrorHandlerFunc,
		})
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
