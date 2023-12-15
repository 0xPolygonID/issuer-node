package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	vault "github.com/hashicorp/vault/api"

	"github.com/polygonid/issuer-node/internal/buildinfo"
	"github.com/polygonid/issuer-node/internal/config"
	"github.com/polygonid/issuer-node/internal/core/event"
	"github.com/polygonid/issuer-node/internal/core/ports"
	"github.com/polygonid/issuer-node/internal/core/services"
	"github.com/polygonid/issuer-node/internal/db"
	"github.com/polygonid/issuer-node/internal/gateways"
	"github.com/polygonid/issuer-node/internal/kms"
	"github.com/polygonid/issuer-node/internal/loader"
	"github.com/polygonid/issuer-node/internal/log"
	"github.com/polygonid/issuer-node/internal/providers"
	"github.com/polygonid/issuer-node/internal/redis"
	"github.com/polygonid/issuer-node/internal/repositories"
	"github.com/polygonid/issuer-node/pkg/blockchain/eth"
	"github.com/polygonid/issuer-node/pkg/cache"
	"github.com/polygonid/issuer-node/pkg/credentials/revocation_status"
	"github.com/polygonid/issuer-node/pkg/http"
	"github.com/polygonid/issuer-node/pkg/pubsub"
	"github.com/polygonid/issuer-node/pkg/reverse_hash"
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

	err = config.CheckDID(ctx, cfg, vaultCli)
	if err != nil {
		log.Error(ctx, "cannot initialize did", "err", err)
		return
	}

	connectionsService := services.NewConnection(connectionsRepository, storage)
	credentialsService, err := newCredentialsService(ctx, cfg, storage, cachex, ps, vaultCli)
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

func newCredentialsService(ctx context.Context, cfg *config.Configuration, storage *db.Storage, cachex cache.Cache, ps pubsub.Client, vaultCli *vault.Client) (ports.ClaimsService, error) {
	identityRepository := repositories.NewIdentity()
	claimsRepository := repositories.NewClaims()
	mtRepository := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepository := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	keyStore, err := kms.Open(cfg.KeyStore.PluginIden3MountPath, vaultCli)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize kms: err %s", err.Error())
	}

	commonClient, err := ethclient.Dial(cfg.Ethereum.URL)
	if err != nil {
		log.Error(ctx, "error dialing with ethclient", "err", err, "eth-url", cfg.Ethereum.URL)
		return nil, err
	}

	ethConn := eth.NewClient(commonClient, &eth.ClientConfig{
		DefaultGasLimit:        cfg.Ethereum.DefaultGasLimit,
		ConfirmationTimeout:    cfg.Ethereum.ConfirmationTimeout,
		ConfirmationBlockCount: cfg.Ethereum.ConfirmationBlockCount,
		ReceiptTimeout:         cfg.Ethereum.ReceiptTimeout,
		MinGasPrice:            big.NewInt(int64(cfg.Ethereum.MinGasPrice)),
		MaxGasPrice:            big.NewInt(int64(cfg.Ethereum.MaxGasPrice)),
		RPCResponseTimeout:     cfg.Ethereum.RPCResponseTimeout,
		WaitReceiptCycleTime:   cfg.Ethereum.WaitReceiptCycleTime,
		WaitBlockCycleTime:     cfg.Ethereum.WaitBlockCycleTime,
	}, keyStore)

	rhsFactory := reverse_hash.NewFactory(cfg.CredentialStatus.RHS.URL, ethConn, common.HexToAddress(cfg.CredentialStatus.OnchainTreeStore.SupportedTreeStoreContract), reverse_hash.DefaultRHSTimeOut)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(cfg.CredentialStatus)
	// TODO: Cache only if cfg.APIUI.SchemaCache == true
	schemaLoader := loader.NewDocumentLoader(cfg.IPFS.GatewayURL)

	mtService := services.NewIdentityMerkleTrees(mtRepository)
	qrService := services.NewQrStoreService(cachex)

	identityService := services.NewIdentity(keyStore, identityRepository, mtRepository, identityStateRepository, mtService, qrService, claimsRepository, revocationRepository, nil, storage, nil, nil, ps, cfg.CredentialStatus, rhsFactory, revocationStatusResolver)
	claimsService := services.NewClaim(claimsRepository, identityService, qrService, mtService, identityStateRepository, schemaLoader, storage, cfg.APIUI.ServerURL, ps, cfg.IPFS.GatewayURL, revocationStatusResolver)

	return claimsService, nil
}
