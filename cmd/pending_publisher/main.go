package main

import (
	"context"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	vault "github.com/hashicorp/vault/api"

	"github.com/polygonid/sh-id-platform/internal/buildinfo"
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
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
	"github.com/polygonid/sh-id-platform/pkg/cache"
	"github.com/polygonid/sh-id-platform/pkg/loaders"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

var build = buildinfo.Revision()

func main() {
	log.Info(context.Background(), "starting issuer node...", "revision", build)

	cfg, err := config.Load("")
	if err != nil {
		log.Error(context.Background(), "cannot load config", "err", err)
		panic(err)
	}

	// Context with log
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
	ps := pubsub.NewRedis(rdb)
	ps.WithLogger(log.Error)
	cachex := cache.NewRedisCache(rdb)

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

	var schemaLoader loader.Factory
	schemaLoader = loader.MultiProtocolFactory(cfg.IFPS.GatewayURL)
	if cfg.APIUI.SchemaCache != nil && *cfg.APIUI.SchemaCache {
		schemaLoader = loader.CachedFactory(schemaLoader, cachex)
	}

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

	bjjKeyProvider, err := kms.NewVaultPluginIden3KeyProvider(vaultCli, cfg.KeyStore.PluginIden3MountPath, kms.KeyTypeBabyJubJub)
	if err != nil {
		log.Error(ctx, "cannot create BabyJubJub key provider", "err", err)
		panic(err)
	}

	ethKeyProvider, err := kms.NewVaultPluginIden3KeyProvider(vaultCli, cfg.KeyStore.PluginIden3MountPath, kms.KeyTypeEthereum)
	if err != nil {
		log.Error(ctx, "cannot create Ethereum key provider", "err", err)
		panic(err)
	}

	keyStore := kms.NewKMS()
	err = keyStore.RegisterKeyProvider(kms.KeyTypeBabyJubJub, bjjKeyProvider)
	if err != nil {
		log.Error(ctx, "cannot register BabyJubJub key provider", "err", err)
		panic(err)
	}

	err = keyStore.RegisterKeyProvider(kms.KeyTypeEthereum, ethKeyProvider)
	if err != nil {
		log.Error(ctx, "cannot register Ethereum key provider", "err", err)
		panic(err)
	}

	err = config.CheckDID(ctx, cfg, vaultCli)
	if err != nil {
		log.Error(ctx, "cannot initialize did", "err", err)
		return
	}

	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepo := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	qrService := services.NewQrStoreService(cachex)

	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	connectionsRepository := repositories.NewConnections()
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, qrService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	claimsService := services.NewClaim(claimsRepo, identityService, qrService, mtService, identityStateRepo, schemaLoader, storage, services.ClaimCfg{
		RHSEnabled:               cfg.ReverseHashService.Enabled,
		RHSUrl:                   cfg.ReverseHashService.URL,
		Host:                     cfg.ServerUrl,
		SCAddress:                cfg.OnChainRevocationService.SCAddress,
		OnChainRevocationEnabled: cfg.OnChainRevocationService.Enabled,
		ChainID:                  cfg.OnChainRevocationService.ChainID,
	}, ps, cfg.IFPS.GatewayURL)

	commonClient, err := ethclient.Dial(cfg.Ethereum.URL)
	if err != nil {
		panic("Error dialing with ethclient: " + err.Error())
	}

	cl := eth.NewClient(commonClient, &eth.ClientConfig{
		DefaultGasLimit:        cfg.Ethereum.DefaultGasLimit,
		ConfirmationTimeout:    cfg.Ethereum.ConfirmationTimeout,
		ConfirmationBlockCount: cfg.Ethereum.ConfirmationBlockCount,
		ReceiptTimeout:         cfg.Ethereum.ReceiptTimeout,
		MinGasPrice:            big.NewInt(int64(cfg.Ethereum.MinGasPrice)),
		MaxGasPrice:            big.NewInt(int64(cfg.Ethereum.MaxGasPrice)),
		RPCResponseTimeout:     cfg.Ethereum.RPCResponseTimeout,
		WaitReceiptCycleTime:   cfg.Ethereum.WaitReceiptCycleTime,
		WaitBlockCycleTime:     cfg.Ethereum.WaitBlockCycleTime,
	})

	circuitsLoaderService := loaders.NewCircuits(cfg.Circuit.Path)
	proofService := initProofService(ctx, cfg, circuitsLoaderService)

	transactionService, err := gateways.NewTransaction(cl, cfg.Ethereum.ConfirmationBlockCount)
	if err != nil {
		log.Error(ctx, "error creating transaction service", "err", err)
		panic("error creating transaction service")
	}
	publisherGateway, err := gateways.NewPublisherEthGateway(cl, common.HexToAddress(cfg.Ethereum.ContractAddress), keyStore, cfg.PublishingKeyPath)
	if err != nil {
		log.Error(ctx, "error creating publish gateway", "err", err)
		panic("error creating publish gateway")
	}
	publisher := gateways.NewPublisher(storage, identityService, claimsService, mtService, keyStore, transactionService, proofService, publisherGateway, cfg.Ethereum.ConfirmationTimeout, ps)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func(ctx context.Context) {
		ticker := time.NewTicker(cfg.OnChainCheckStatusFrequency)
		for {
			select {
			case <-ticker.C:
				publisher.CheckTransactionStatus(ctx)
			case <-ctx.Done():
				log.Info(ctx, "finishing check transaction status job")
			}
		}
	}(ctx)

	<-quit
	log.Info(ctx, "finishing app")
	cancel()
	log.Info(ctx, "Finished")
}

func initProofService(ctx context.Context, config *config.Configuration, circuitLoaderService *loaders.Circuits) ports.ZKGenerator {
	log.Info(ctx, "native prover enabled", "enabled", config.NativeProofGenerationEnabled)
	if config.NativeProofGenerationEnabled {
		proverConfig := &services.NativeProverConfig{
			CircuitsLoader: circuitLoaderService,
		}
		return services.NewNativeProverService(proverConfig)
	}

	proverConfig := &gateways.ProverConfig{
		ServerURL:       config.Prover.ServerURL,
		ResponseTimeout: config.Prover.ResponseTimeout,
	}
	return gateways.NewProverService(proverConfig)
}
