package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/buildinfo"
	"github.com/polygonid/sh-id-platform/internal/cache"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/event"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/gateways"
	httpPkg "github.com/polygonid/sh-id-platform/internal/http"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	network2 "github.com/polygonid/sh-id-platform/internal/network"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/pubsub"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	reverse_hash2 "github.com/polygonid/sh-id-platform/internal/reverse_hash"
	"github.com/polygonid/sh-id-platform/internal/revocation_status"
)

var build = buildinfo.Revision()

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info(ctx, "starting issuer node...", "revision", build)

	cfg, err := config.Load()
	if err != nil {
		log.Error(ctx, "cannot load config", "err", err)
		return
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
		return
	}

	connectionsRepository := repositories.NewConnection()
	claimsRepository := repositories.NewClaim()

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

	connectionsService := services.NewConnection(connectionsRepository, claimsRepository, storage)
	credentialsService, err := newCredentialsService(ctx, cfg, storage, cachex, ps, keyStore)
	if err != nil {
		log.Error(ctx, "cannot initialize the credential service", "err", err)
		return
	}

	notificationGateway := gateways.NewPushNotificationClient(httpPkg.DefaultHTTPClientWithRetry)
	notificationService := services.NewNotification(notificationGateway, connectionsService, credentialsService)
	ctxCancel, cancel := context.WithCancel(ctx)
	defer func() {
		log.Info(ctx, "Shutting down...")
		cancel()
		if err := ps.Close(); err != nil {
			log.Error(ctx, "closing redis connection", "err", err)
		}
	}()

	ps.Subscribe(ctxCancel, event.CreateCredentialEvent, notificationService.SendCreateCredentialNotification)
	ps.Subscribe(ctxCancel, event.CreateConnectionEvent, notificationService.SendCreateConnectionNotification)
	ps.Subscribe(ctxCancel, event.CreateStateEvent, notificationService.SendRevokeCredentialNotification)

	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		http.Handle("/status", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("OK"))
			if err != nil {
				log.Error(ctx, "error writing response", "err", err)
			}
		}))
		log.Info(ctx, "Starting server at port 3004")
		err := http.ListenAndServe(":3004", nil)
		if err != nil {
			log.Error(ctx, "error starting server", "err", err)
		}
	}()

	<-gracefulShutdown
}

func newCredentialsService(ctx context.Context, cfg *config.Configuration, storage *db.Storage, cachex cache.Cache, ps pubsub.Client, keyStore *kms.KMS) (ports.ClaimService, error) {
	identityRepository := repositories.NewIdentity()
	claimsRepository := repositories.NewClaim()
	mtRepository := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepository := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()

	reader, err := network2.GetReaderFromConfig(cfg, ctx)
	if err != nil {
		log.Error(ctx, "cannot read network resolver file", "err", err)
		return nil, err
	}
	networkResolver, err := network2.NewResolver(ctx, *cfg, keyStore, reader)
	if err != nil {
		log.Error(ctx, "failed initialize network resolver", "err", err)
		return nil, err
	}

	rhsFactory := reverse_hash2.NewFactory(*networkResolver, reverse_hash2.DefaultRHSTimeOut)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	schemaLoader := loader.NewDocumentLoader(cfg.IPFS.GatewayURL, cfg.SchemaCache)

	mtService := services.NewIdentityMerkleTrees(mtRepository)
	qrService := services.NewQrStoreService(cachex)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		*cfg.MediaTypeManager.Enabled,
	)

	identityService := services.NewIdentity(keyStore, identityRepository, mtRepository, identityStateRepository, mtService, qrService, claimsRepository, revocationRepository, nil, storage, nil, nil, ps, *networkResolver, rhsFactory, revocationStatusResolver)
	claimsService := services.NewClaim(claimsRepository, identityService, qrService, mtService, identityStateRepository, schemaLoader, storage, cfg.ServerUrl, ps, cfg.IPFS.GatewayURL, revocationStatusResolver, mediaTypeManager, cfg.UniversalLinks)

	return claimsService, nil
}
