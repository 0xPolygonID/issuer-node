package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/gateways"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
	"github.com/polygonid/sh-id-platform/pkg/loaders"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Error(context.Background(), "cannot load config", err)
		panic(err)
	}

	onChainPublishStateFrecuency, err := time.ParseDuration(cfg.OnChainPublishStateFrecuency)
	if err != nil {
		panic("error converting onChainPublishStateFrecuency param")
	}

	onChainCheckStatusFrecuency, err := time.ParseDuration(cfg.OnChainCheckStatusFrecuency)
	if err != nil {
		panic("error converting onChainCheckStatusFrecuency param")
	}

	// Context with log
	ctx1 := log.NewContext(context.Background(), cfg.Log.Level, cfg.Log.Mode, os.Stdout)
	ctx, cancel := context.WithCancel(ctx1)

	storage, err := db.NewStorage(cfg.Database.URL)
	if err != nil {
		log.Error(ctx, "cannot connect to database", err)
		panic(err)
	}

	defer func(storage *db.Storage) {
		err := storage.Close()
		if err != nil {
			log.Error(ctx, "error closing database connection", err)
		}
	}(storage)

	vaultCli, err := providers.NewVaultClient(cfg.KeyStore.Address, cfg.KeyStore.Token)
	if err != nil {
		log.Error(ctx, "cannot init vault client: ", err)
		panic(err)
	}

	bjjKeyProvider, err := kms.NewVaultPluginIden3KeyProvider(vaultCli, cfg.KeyStore.PluginIden3MountPath, kms.KeyTypeBabyJubJub)
	if err != nil {
		log.Error(ctx, "cannot create BabyJubJub key provider", err)
		panic(err)
	}

	ethKeyProvider, err := kms.NewVaultPluginIden3KeyProvider(vaultCli, cfg.KeyStore.PluginIden3MountPath, kms.KeyTypeEthereum)
	if err != nil {
		log.Error(ctx, "cannot create Ethereum key provider", err)
		panic(err)
	}

	keyStore := kms.NewKMS()
	err = keyStore.RegisterKeyProvider(kms.KeyTypeBabyJubJub, bjjKeyProvider)
	if err != nil {
		log.Error(ctx, "cannot register BabyJubJub key provider", err)
		panic(err)
	}

	err = keyStore.RegisterKeyProvider(kms.KeyTypeEthereum, ethKeyProvider)
	if err != nil {
		log.Error(ctx, "cannot register Ethereum key provider", err)
		panic(err)
	}

	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepo := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	mtService := services.NewIdentityMerkleTrees(mtRepo)

	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, storage, rhsp)
	schemaService := services.NewSchema(storage)
	claimsService := services.NewClaim(
		claimsRepo,
		schemaService,
		identityService,
		mtService,
		identityStateRepo,
		storage,
		services.ClaimCfg{
			RHSEnabled: cfg.ReverseHashService.Enabled,
			RHSUrl:     cfg.ReverseHashService.URL,
			Host:       cfg.ServerUrl,
		},
	)

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
	proofService := initProofService(cfg, circuitsLoaderService)

	transactionService, err := gateways.NewTransaction(cl, cfg.Ethereum.ConfirmationBlockCount)
	if err != nil {
		log.Error(ctx, "error creating transaction service", err)
		panic("error creating transaction service")
	}
	publisherGateway, err := gateways.NewPublisherEthGateway(cl, common.HexToAddress(cfg.Ethereum.ContractAddress), keyStore, cfg.PublishingKeyPath)
	if err != nil {
		log.Error(ctx, "error creating publish gateway", err)
		panic("error creating publish gateway")
	}
	publisher := gateways.NewPublisher(storage, identityService, claimsService, mtService, keyStore, transactionService, proofService, publisherGateway, cfg.Ethereum.ConfirmationTimeout)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func(ctx context.Context) {
		ticker := time.NewTicker(onChainPublishStateFrecuency)
		for {
			select {
			case <-ticker.C:
				publisher.PublishState(ctx)
			case <-ctx.Done():
				log.Info(ctx, "finishing publish state job..")
			}
		}
	}(ctx)

	go func(ctx context.Context) {
		ticker := time.NewTicker(onChainCheckStatusFrecuency)
		for {
			select {
			case <-ticker.C:
				publisher.CheckTransactionStatus(ctx)
			case <-ctx.Done():
				log.Info(ctx, "finishing check transaction status job..")
			}
		}
	}(ctx)

	<-quit
	log.Info(ctx, "finishing app")
	cancel()
	log.Info(ctx, "Finshed")
}

func initProofService(config *config.Configuration, circuitLoaderService *loaders.Circuits) ports.ZKGenerator {
	log.Info(context.Background(), fmt.Sprintf("native prover enabled: %v", config.NativeProofGenerationEnabled))
	if config.NativeProofGenerationEnabled {
		proverConfig := &services.NativeProverConfig{
			CircuitsLoader: circuitLoaderService,
		}
		return services.NewNativeProverService(proverConfig)
	}

	// TODO: add another prover
	proverConfig := &gateways.ProverConfig{
		ServerURL:       config.Prover.ServerURL,
		ResponseTimeout: config.Prover.ResponseTimeout,
	}
	return gateways.NewProverService(proverConfig)
}
