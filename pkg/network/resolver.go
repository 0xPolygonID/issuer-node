package network

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/yaml.v3"

	com "github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
)

type resolverPrefix string

// ResolverClientConfig holds the resolver client config
type ResolverClientConfig struct {
	client          *eth.Client
	contractAddress string
}

// Resolver holds the resolver
type Resolver struct {
	ethereumClients map[resolverPrefix]ResolverClientConfig
	rhsSettings     map[resolverPrefix]RhsSettings
}

// RhsSettings holds the rhs settings
type RhsSettings struct {
	DirectUrl            string  `yaml:"directUrl"`
	Mode                 string  `yaml:"mode"`
	ContractAddress      *string `yaml:"contractAddress"`
	RhsUrl               *string `yaml:"rhsUrl"`
	ChainID              *string `yaml:"chainID"`
	PublishingKey        string  `yaml:"publishingKey"`
	SingleIssuer         bool
	CredentialStatusType string
}

// ResolverSettings holds the resolver settings
type ResolverSettings map[string]map[string]struct {
	ContractAddress        string        `yaml:"contractAddress"`
	NetworkURL             string        `yaml:"networkURL"`
	DefaultGasLimit        int           `yaml:"defaultGasLimit"`
	ConfirmationTimeout    time.Duration `yaml:"confirmationTimeout"`
	ConfirmationBlockCount int64         `yaml:"confirmationBlockCount"`
	ReceiptTimeout         time.Duration `yaml:"receiptTimeout"`
	MinGasPrice            int           `yaml:"minGasPrice"`
	MaxGasPrice            int           `yaml:"maxGasPrice"`
	RPCResponseTimeout     time.Duration `yaml:"rpcResponseTimeout"`
	WaitReceiptCycleTime   time.Duration `yaml:"waitReceiptCycleTime"`
	WaitBlockCycleTime     time.Duration `yaml:"waitBlockCycleTime"`
	RhsSettings            RhsSettings   `yaml:"rhsSettings"`
}

// NewResolver returns a new Resolver
func NewResolver(ctx context.Context, cfg config.Configuration, kms *kms.KMS) (*Resolver, error) {
	rs, err := parseResolversSettings(ctx, cfg.NetworkResolverPath)

	ethereumClients := make(map[resolverPrefix]ResolverClientConfig)
	rhsSettings := make(map[resolverPrefix]RhsSettings)

	if err != nil {
		resolverPrefixEnv := cfg.Ethereum.ResolverPrefix
		if resolverPrefixEnv == "" {
			log.Error(ctx, "resolver prefix not found")
			return nil, err
		}

		ethClient, err := ethclient.Dial(cfg.Ethereum.URL)
		if err != nil {
			log.Error(ctx, "cannot connect to ethereum network", "err", err)
			return nil, err
		}

		c := eth.NewClient(ethClient, &eth.ClientConfig{
			DefaultGasLimit:        cfg.Ethereum.DefaultGasLimit,
			ConfirmationTimeout:    cfg.Ethereum.ConfirmationTimeout,
			ConfirmationBlockCount: cfg.Ethereum.ConfirmationBlockCount,
			ReceiptTimeout:         cfg.Ethereum.ReceiptTimeout,
			MinGasPrice:            big.NewInt(int64(cfg.Ethereum.MinGasPrice)),
			MaxGasPrice:            big.NewInt(int64(cfg.Ethereum.MaxGasPrice)),
			RPCResponseTimeout:     cfg.Ethereum.RPCResponseTimeout,
			WaitReceiptCycleTime:   cfg.Ethereum.WaitReceiptCycleTime,
			WaitBlockCycleTime:     cfg.Ethereum.WaitBlockCycleTime,
		}, kms)

		resolverClientConfig := &ResolverClientConfig{
			client:          c,
			contractAddress: cfg.Ethereum.ContractAddress,
		}
		ethereumClients[resolverPrefix(resolverPrefixEnv)] = *resolverClientConfig

		settings := RhsSettings{
			DirectUrl:            cfg.CredentialStatus.DirectStatus.URL,
			Mode:                 string(cfg.CredentialStatus.RHSMode),
			SingleIssuer:         cfg.CredentialStatus.SingleIssuer,
			ContractAddress:      com.ToPointer(cfg.CredentialStatus.OnchainTreeStore.SupportedTreeStoreContract),
			RhsUrl:               com.ToPointer(cfg.CredentialStatus.RHS.URL),
			ChainID:              com.ToPointer(cfg.CredentialStatus.OnchainTreeStore.ChainID),
			CredentialStatusType: cfg.CredentialStatus.CredentialStatusType,
		}
		rhsSettings[resolverPrefix(resolverPrefixEnv)] = settings
		return &Resolver{
			ethereumClients: ethereumClients,
			rhsSettings:     rhsSettings,
		}, nil
	}

	log.Info(ctx, "resolver settings file found", "path", cfg.NetworkResolverPath)
	log.Info(ctx, "the issuer node will use the resolver settings file for configuring multi chain feature")
	for chainName, chainSettings := range rs {
		for networkName, networkSettings := range chainSettings {
			if err != nil {
				return nil, err
			}

			resolverPrefixKey := fmt.Sprintf("%s:%s", chainName, networkName)
			ethClient, err := ethclient.Dial(networkSettings.NetworkURL)
			if err != nil {
				log.Error(ctx, "cannot connect to ethereum network", "err", err, "networkURL", networkSettings.NetworkURL)
				return nil, err
			}

			client := eth.NewClient(ethClient, &eth.ClientConfig{
				DefaultGasLimit:        networkSettings.DefaultGasLimit,
				ConfirmationTimeout:    networkSettings.ConfirmationTimeout,
				ConfirmationBlockCount: networkSettings.ConfirmationBlockCount,
				ReceiptTimeout:         networkSettings.ReceiptTimeout,
				MinGasPrice:            big.NewInt(int64(networkSettings.MinGasPrice)),
				MaxGasPrice:            big.NewInt(int64(networkSettings.MaxGasPrice)),
				RPCResponseTimeout:     networkSettings.RPCResponseTimeout,
				WaitReceiptCycleTime:   networkSettings.WaitReceiptCycleTime,
				WaitBlockCycleTime:     networkSettings.WaitBlockCycleTime,
			}, kms)

			resolverClientConfig := &ResolverClientConfig{
				client:          client,
				contractAddress: networkSettings.ContractAddress,
			}

			ethereumClients[resolverPrefix(resolverPrefixKey)] = *resolverClientConfig

			settings := networkSettings.RhsSettings
			settings.DirectUrl = cfg.CredentialStatus.DirectStatus.URL
			settings.SingleIssuer = cfg.CredentialStatus.SingleIssuer

			if settings.Mode == config.None {
				settings.CredentialStatusType = config.SparseMerkleTreeProof
			}

			if settings.Mode == config.OffChain {
				if settings.RhsUrl == nil {
					return nil, fmt.Errorf("rhs url not found for %s", resolverPrefixKey)
				}
				settings.CredentialStatusType = config.Iden3ReverseSparseMerkleTreeProof
			}

			if settings.Mode == config.OnChain {
				if settings.ContractAddress == nil {
					return nil, fmt.Errorf("contract address not found for %s", resolverPrefixKey)
				}
				settings.CredentialStatusType = config.Iden3OnchainSparseMerkleTreeProof2023
			}

			rhsSettings[resolverPrefix(resolverPrefixKey)] = settings
		}
	}

	return &Resolver{
		ethereumClients: ethereumClients,
		rhsSettings:     rhsSettings,
	}, nil
}

// GetEthClient returns the eth client
func (r *Resolver) GetEthClient(resolverPrefixKey string) (*eth.Client, error) {
	resolverClientConfig, ok := r.ethereumClients[resolverPrefix(resolverPrefixKey)]
	if !ok {
		return nil, fmt.Errorf("ethClient not found for %s", resolverPrefixKey)
	}
	return resolverClientConfig.client, nil
}

// GetContractAddress returns the contract address
func (r *Resolver) GetContractAddress(resolverPrefixKey string) (*common.Address, error) {
	resolverClientConfig, ok := r.ethereumClients[resolverPrefix(resolverPrefixKey)]
	if !ok {
		return nil, fmt.Errorf("contract address not found for %s", resolverPrefixKey)
	}

	contractAddress := common.HexToAddress(resolverClientConfig.contractAddress)
	return &contractAddress, nil
}

// GetRhsSettings returns the rhs settings
func (r *Resolver) GetRhsSettings(resolverPrefixKey string) (*RhsSettings, error) {
	rhsSettings, ok := r.rhsSettings[resolverPrefix(resolverPrefixKey)]
	if !ok {
		return nil, fmt.Errorf("rhsSettings not found for %s", resolverPrefixKey)
	}
	return &rhsSettings, nil
}

// GetConfirmationBlockCount returns the confirmation block count
func (r *Resolver) GetConfirmationBlockCount(resolverPrefixKey string) (int64, error) {
	resolverClientConfig, ok := r.ethereumClients[resolverPrefix(resolverPrefixKey)]
	if !ok {
		return 0, fmt.Errorf("contract address not found for %s", resolverPrefixKey)
	}
	confirmationBlockCount := resolverClientConfig.client.GetConfirmationBlockCount()
	return confirmationBlockCount, nil
}

// GetConfirmationTimeout returns the confirmation timeout
func (r *Resolver) GetConfirmationTimeout(resolverPrefixKey string) (time.Duration, error) {
	resolverClientConfig, ok := r.ethereumClients[resolverPrefix(resolverPrefixKey)]
	if !ok {
		return 0, fmt.Errorf("contract address not found for %s", resolverPrefixKey)
	}
	confirmationTimeout := resolverClientConfig.client.GetConfirmationConfirmationTimeout()
	return confirmationTimeout, nil
}

func parseResolversSettings(ctx context.Context, resolverSettingsPath string) (ResolverSettings, error) {
	if _, err := os.Stat(resolverSettingsPath); errors.Is(err, os.ErrNotExist) {
		log.Info(ctx, "resolver settings file not found", "path", resolverSettingsPath)
		log.Info(ctx, "issuer node wil not run supporting multi chain feature")
		return nil, fmt.Errorf("resolver settings file not found: %s", resolverSettingsPath)
	}

	if info, _ := os.Stat(resolverSettingsPath); info.Size() == 0 {
		log.Info(ctx, "resolver settings file is empty")
		return nil, fmt.Errorf("resolver settings file is empty: %s", resolverSettingsPath)
	}

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
