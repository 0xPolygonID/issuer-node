package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/buildinfo"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/gateways"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/cache"
	"github.com/polygonid/sh-id-platform/pkg/credentials/revocation_status"
	circuitLoaders "github.com/polygonid/sh-id-platform/pkg/loaders"
	"github.com/polygonid/sh-id-platform/pkg/network"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

var build = buildinfo.Revision()

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info(ctx, "starting pending publisher...", "revision", build)

	cfg, err := config.Load()
	if err != nil {
		log.Error(ctx, "cannot load config", "err", err)
		panic(err)
	}

	log.Config(cfg.Log.Level, cfg.Log.Mode, os.Stdout)

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

	storage, err := db.NewStorage(cfg.Database.URL)
	if err != nil {
		log.Error(ctx, "cannot connect to database", "err", err)
		panic(err)
	}

	defer func(storage *db.Storage) {
		err := storage.Close()
		if err != nil {
			log.Error(ctx, "error closing database connection", "err", err)
		}
	}(storage)

	// TODO: Cache only if cfg.APIUI.SchemaCache == true
	schemaLoader := loader.NewDocumentLoader(cfg.IPFS.GatewayURL)

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

	reader, err := network.GetReaderFromConfig(cfg, ctx)
	if err != nil {
		log.Error(ctx, "cannot read network resolver file", "err", err)
		return
	}
	networkResolver, err := network.NewResolver(ctx, *cfg, keyStore, reader)
	if err != nil {
		log.Error(ctx, "failed init eth resolver", "err", err)
		return
	}

	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepo := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	qrService := services.NewQrStoreService(cachex)

	connectionsRepository := repositories.NewConnections()

	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		*cfg.MediaTypeManager.Enabled,
	)

	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, qrService, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)
	claimsService := services.NewClaim(claimsRepo, identityService, qrService, mtService, identityStateRepo, schemaLoader, storage, cfg.ServerUrl, ps, cfg.IPFS.GatewayURL, revocationStatusResolver, mediaTypeManager, cfg.UniversalLinks)

	circuitsLoaderService := circuitLoaders.NewCircuits(cfg.Circuit.Path)
	proofService := initProofService(circuitsLoaderService)

	transactionService, err := gateways.NewTransaction(*networkResolver)
	if err != nil {
		log.Error(ctx, "error creating transaction service", "err", err)
		panic("error creating transaction service")
	}
	publisherGateway, err := gateways.NewPublisherEthGateway(*networkResolver, keyStore, cfg.PublishingKeyPath)
	if err != nil {
		log.Error(ctx, "error creating publish gateway", "err", err)
		panic("error creating publish gateway")
	}
	publisher := gateways.NewPublisher(storage, identityService, claimsService, mtService, keyStore, transactionService, proofService, publisherGateway, networkResolver, ps)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func(ctx context.Context) {
		ticker := time.NewTicker(cfg.OnChainCheckStatusFrequency)
		for {
			select {
			// TODO: Config this
			case <-ticker.C:
				publisher.CheckTransactionStatus(ctx, nil)
			case <-ctx.Done():
				log.Info(ctx, "finishing check transaction status job")
			}
		}
	}(ctx)

	go func() {
		http.Handle("/status", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("OK"))
			if err != nil {
				log.Error(ctx, "error writing response", "err", err)
			}
		}))
		log.Info(ctx, "Starting server at port 3005")
		err := http.ListenAndServe(":3005", nil)
		if err != nil {
			log.Error(ctx, "error starting server", "err", err)
		}
	}()

	<-quit
	log.Info(ctx, "finishing app")
	cancel()
	log.Info(ctx, "Finished")
}

func initProofService(circuitLoaderService *circuitLoaders.Circuits) ports.ZKGenerator {
	proverConfig := &services.NativeProverConfig{
		CircuitsLoader: circuitLoaderService,
	}
	return services.NewNativeProverService(proverConfig)
}
