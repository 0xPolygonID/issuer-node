package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iden3/contracts-abi/state/go/abi"
	"gopkg.in/yaml.v3"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
)

// ResolverSettings holds the resolver settings
type ResolverSettings map[string]map[string]struct {
	ContractAddress string `yaml:"contractAddress"`
	NetworkURL      string `yaml:"networkURL"`
}

// InitEthClient returns a State Contract Instance
func InitEthClient(ethURL, contractAddress string) (*abi.State, error) {
	ec, err := ethclient.Dial(ethURL)
	if err != nil {
		return nil, fmt.Errorf("failed connect to eth node %s: %s", ethURL, err.Error())
	}
	stateContractInstance, err := abi.NewState(common.HexToAddress(contractAddress), ec)
	if err != nil {
		return nil, fmt.Errorf("error failed create state contract client: %s", err.Error())
	}

	return stateContractInstance, nil
}

// InitEthClients returns a State Contract Instance
func InitEthClients(ctx context.Context, cfg config.Configuration) (map[string]*abi.State, error) {
	rs, err := parseResolversSettings(ctx, cfg.NetworkResolverPath)
	supportedContracts := map[string]*abi.State{}
	if err != nil {
		log.Info(ctx, "failed to parse resolvers settings")
		stateContract, err := InitEthClient(cfg.Ethereum.URL, cfg.Ethereum.ContractAddress)
		supportedContracts[cfg.Ethereum.ResolverPrefix] = stateContract
		return supportedContracts, err
	}

	for chainName, chainSettings := range rs {
		for networkName, networkSettings := range chainSettings {
			stateContract, err := InitEthClient(networkSettings.NetworkURL, networkSettings.ContractAddress)
			if err != nil {
				return nil, err
			}
			supportedContracts[fmt.Sprintf("%s:%s", chainName, networkName)] = stateContract
		}
	}

	return supportedContracts, nil
}

// InitEthConnect opens a new eth connection
func InitEthConnect(cfg config.Ethereum, kms *kms.KMS) (*eth.Client, error) {
	commonClient, err := ethclient.Dial(cfg.URL)
	if err != nil {
		return nil, err
	}

	cl := eth.NewClient(commonClient, &eth.ClientConfig{
		DefaultGasLimit:        cfg.DefaultGasLimit,
		ConfirmationTimeout:    cfg.ConfirmationTimeout,
		ConfirmationBlockCount: cfg.ConfirmationBlockCount,
		ReceiptTimeout:         cfg.ReceiptTimeout,
		MinGasPrice:            big.NewInt(int64(cfg.MinGasPrice)),
		MaxGasPrice:            big.NewInt(int64(cfg.MaxGasPrice)),
		RPCResponseTimeout:     cfg.RPCResponseTimeout,
		WaitReceiptCycleTime:   cfg.WaitReceiptCycleTime,
		WaitBlockCycleTime:     cfg.WaitBlockCycleTime,
	},
		kms,
	)

	return cl, nil
}

// Open returns an initialized eth Client with the given configuration
func Open(cfg *config.Configuration, kms *kms.KMS) (*eth.Client, error) {
	ethClient, err := ethclient.Dial(cfg.Ethereum.URL)
	if err != nil {
		return nil, err
	}

	return eth.NewClient(ethClient, &eth.ClientConfig{
		DefaultGasLimit:        cfg.Ethereum.DefaultGasLimit,
		ConfirmationTimeout:    cfg.Ethereum.ConfirmationTimeout,
		ConfirmationBlockCount: cfg.Ethereum.ConfirmationBlockCount,
		ReceiptTimeout:         cfg.Ethereum.ReceiptTimeout,
		MinGasPrice:            big.NewInt(int64(cfg.Ethereum.MinGasPrice)),
		MaxGasPrice:            big.NewInt(int64(cfg.Ethereum.MaxGasPrice)),
		RPCResponseTimeout:     cfg.Ethereum.RPCResponseTimeout,
		WaitReceiptCycleTime:   cfg.Ethereum.WaitReceiptCycleTime,
		WaitBlockCycleTime:     cfg.Ethereum.WaitBlockCycleTime,
	}, kms), nil
}

func parseResolversSettings(ctx context.Context, resolverSettingsPath string) (ResolverSettings, error) {
	f, err := os.Open(filepath.Clean(resolverSettingsPath))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Error(ctx, "failed to close setting file:", "err", err)
		}
	}()

	settings := ResolverSettings{}
	if err := yaml.NewDecoder(f).Decode(&settings); err != nil {
		return nil, fmt.Errorf("invalid yaml file: %v", settings)
	}
	return settings, nil
}
