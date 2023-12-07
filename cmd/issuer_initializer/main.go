package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	vault "github.com/hashicorp/vault/api"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/buildinfo"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
	"github.com/polygonid/sh-id-platform/pkg/credentials/revocation_status"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

var build = buildinfo.Revision()

const (
	timeToWaitForVault = 5 * time.Second
)

func main() {
	log.Info(context.Background(), "starting issuer node...", "revision", build)
	cfg, err := config.Load("")
	if err != nil {
		log.Error(context.Background(), "cannot load config", "err", err)
		return
	}

	ctx, cancel := context.WithCancel(log.NewContext(context.Background(), cfg.Log.Level, cfg.Log.Mode, os.Stdout))
	defer cancel()

	if err := cfg.Sanitize(ctx); err != nil {
		log.Error(ctx, "there are errors in the configuration that prevent server to start", "err", err)
		return
	}

	storage, err := db.NewStorage(cfg.Database.URL)
	if err != nil {
		log.Error(ctx, "cannot connect to database", "err", err)
		return
	}

	var vaultCli *vault.Client
	var vaultErr error
	vaultAttempts := 10
	connected := false

	vaultCfg := providers.Config{
		UserPassAuthEnabled: cfg.VaultUserPassAuthEnabled,
		Address:             cfg.KeyStore.Address,
		Token:               cfg.KeyStore.Token,
		Pass:                cfg.VaultUserPassAuthPassword,
	}
	for i := 0; i < vaultAttempts; i++ {
		vaultCli, vaultErr = providers.VaultClient(ctx, vaultCfg)
		if vaultErr == nil {
			connected = true
			break
		}
		log.Error(ctx, "cannot connect to vault, retrying", "err", vaultErr)
		time.Sleep(timeToWaitForVault)
	}

	if !connected {
		log.Error(ctx, "cannot initialize vault client", "err", err)
		return
	}

	did, err := providers.GetDID(ctx, vaultCli)
	if err != nil {
		if errors.Is(err, providers.VaultConnErr) {
			log.Error(ctx, "cannot connect to vault", "err", err)
			return
		}
		log.Info(ctx, "did not found in vault, creating new one")
	}

	if did != "" {
		log.Info(ctx, "did already created, skipping", "info", "if you want to create new one, please remove did from vault: 'vault kv delete kv/did'")
		//nolint:all
		fmt.Printf("\n")
		//nolint:all
		fmt.Printf(did)
		//nolint:all
		fmt.Printf("\n")
		return
	}

	keyStore, err := kms.Open(cfg.KeyStore.PluginIden3MountPath, vaultCli)
	if err != nil {
		log.Error(ctx, "cannot initialize kms", "err", err)
		return
	}

	// repositories initialization
	identityRepository := repositories.NewIdentity()
	claimsRepository := repositories.NewClaims()
	mtRepository := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepository := repositories.NewIdentityState()

	// services initialization
	mtService := services.NewIdentityMerkleTrees(mtRepository)

	commonClient, err := ethclient.Dial(cfg.Ethereum.URL)
	if err != nil {
		log.Error(ctx, "error dialing with ethclient", "err", err, "rth-url", cfg.Ethereum.URL)
		return
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

	rhsFactory := reverse_hash.NewFactory(cfg.CredentialStatus.RHS.GetURL(), ethConn, common.HexToAddress(cfg.CredentialStatus.OnchainTreeStore.SupportedTreeStoreContract), reverse_hash.DefaultRHSTimeOut)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(cfg.CredentialStatus)
	cfg.CredentialStatus.SingleIssuer = true
	// this is needed to create the did with the correct auth core claim revocation status URL
	cfg.CredentialStatus.DirectStatus.URL = cfg.APIUI.ServerURL
	identityService := services.NewIdentity(keyStore, identityRepository, mtRepository, identityStateRepository, mtService, nil, claimsRepository, nil, nil, storage, nil, nil, nil, cfg.CredentialStatus, rhsFactory, revocationStatusResolver)

	didCreationOptions := &ports.DIDCreationOptions{
		Method:                  core.DIDMethod(cfg.APIUI.IdentityMethod),
		Network:                 core.NetworkID(cfg.APIUI.IdentityNetwork),
		Blockchain:              core.Blockchain(cfg.APIUI.IdentityBlockchain),
		KeyType:                 kms.KeyType(cfg.APIUI.KeyType),
		AuthBJJCredentialStatus: verifiable.CredentialStatusType(cfg.CredentialStatus.CredentialStatusType),
	}

	identity, err := identityService.Create(ctx, cfg.APIUI.ServerURL, didCreationOptions)
	if err != nil {
		log.Error(ctx, "error creating identifier", err)
		return
	}

	log.Info(ctx, "identifier crated successfully")

	if err := providers.SaveDID(ctx, vaultCli, identity.Identifier); err != nil {
		log.Error(ctx, "error saving identifier to vault", err)
		return
	}

	//nolint:all
	fmt.Printf("\n")
	//nolint:all
	fmt.Printf(identity.Identifier)
	//nolint:all
	fmt.Printf("\n")
}
