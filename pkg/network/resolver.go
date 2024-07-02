package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iden3/contracts-abi/state/go/abi"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"gopkg.in/yaml.v3"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
)

const (
	// OnChain is the type for revocation status on chain
	OnChain = "OnChain"
	// OffChain is the type for revocation status off chain
	OffChain = "OffChain"
	// None is the type for revocation status None
	None = "None"
)

type resolverPrefix string

// ResolverClientConfig holds the resolver client config
type ResolverClientConfig struct {
	client          *eth.Client
	contractAddress string
}

// Resolver holds the resolver
type Resolver struct {
	ethereumClients    map[resolverPrefix]ResolverClientConfig
	rhsSettings        map[resolverPrefix]RhsSettings
	supportedContracts map[string]*abi.State
}

// RhsSettings holds the rhs settings
type RhsSettings struct {
	Iden3CommAgentStatus string  `yaml:"directUrl"`
	Mode                 string  `yaml:"mode"`
	ContractAddress      *string `yaml:"contractAddress"`
	RhsUrl               *string `yaml:"rhsUrl"`
	ChainID              *string `yaml:"chainID"`
	PublishingKey        string  `yaml:"publishingKey"`
	SingleIssuer         bool
	CredentialStatusType verifiable.CredentialStatusType
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
	GasLess                bool          `yaml:"gasLess"`
	TransferAmountWei      *big.Int      `yaml:"transferAmountWei"`
	RhsSettings            RhsSettings   `yaml:"rhsSettings"`
	NetworkFlag            byte          `yaml:"networkFlag"`
	ChainID                string        `yaml:"chainID"`
	Method                 string        `yaml:"method"`
}

// NewResolver returns a new Resolver
func NewResolver(ctx context.Context, cfg config.Configuration, kms *kms.KMS, reader io.Reader) (*Resolver, error) {
	rs, err := parseResolversSettings(ctx, reader)
	if err != nil {
		return nil, errors.New("failed to parse resolver settings")
	}

	ethereumClients := make(map[resolverPrefix]ResolverClientConfig)
	rhsSettings := make(map[resolverPrefix]RhsSettings)
	supportedContracts := make(map[string]*abi.State)

	log.Info(ctx, "resolver settings file found", "path", cfg.NetworkResolverPath)
	log.Info(ctx, "the issuer node will use the resolver settings file for configuring multi chain feature")
	for chainName, chainSettings := range rs {
		for networkName, networkSettings := range chainSettings {
			if networkSettings.NetworkFlag != 0 {
				if err := registerCustomDIDMethod(ctx, chainName, networkName, networkSettings.ChainID, networkSettings.Method, networkSettings.NetworkFlag); err != nil {
					return nil, fmt.Errorf("failed to register custom DID method: %w", err)
				}
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
				GasLess:                networkSettings.GasLess,
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

			// TODO: Change this when two apis are merged
			if cfg.CredentialStatus.SingleIssuer {
				settings.Iden3CommAgentStatus = strings.TrimSuffix(cfg.APIUI.ServerURL, "/")
			} else {
				settings.Iden3CommAgentStatus = strings.TrimSuffix(cfg.ServerUrl, "/")
			}

			settings.SingleIssuer = cfg.CredentialStatus.SingleIssuer

			if settings.Mode == None {
				settings.CredentialStatusType = verifiable.Iden3commRevocationStatusV1
			}

			if settings.Mode == OffChain {
				if settings.RhsUrl == nil {
					return nil, fmt.Errorf("rhs url not found for %s", resolverPrefixKey)
				}
				settings.CredentialStatusType = verifiable.Iden3ReverseSparseMerkleTreeProof
			}

			if settings.Mode == OnChain {
				if settings.ContractAddress == nil {
					return nil, fmt.Errorf("contract address not found for %s", resolverPrefixKey)
				}
				settings.CredentialStatusType = verifiable.Iden3OnchainSparseMerkleTreeProof2023
			}

			rhsSettings[resolverPrefix(resolverPrefixKey)] = settings
			stateContract, err := abi.NewState(common.HexToAddress(networkSettings.ContractAddress), ethClient)
			if err != nil {
				return nil, fmt.Errorf("error failed create state contract client: %s", err.Error())
			}
			supportedContracts[resolverPrefixKey] = stateContract
		}
	}

	return &Resolver{
		ethereumClients:    ethereumClients,
		rhsSettings:        rhsSettings,
		supportedContracts: supportedContracts,
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
func (r *Resolver) GetRhsSettings(ctx context.Context, resolverPrefixKey string) (*RhsSettings, error) {
	rhsSettings, ok := r.rhsSettings[resolverPrefix(resolverPrefixKey)]
	if !ok {
		log.Error(ctx, "rhsSettings not found", "resolverPrefixKey", resolverPrefixKey)
		return nil, fmt.Errorf("rhsSettings not found for %s", resolverPrefixKey)
	}
	return &rhsSettings, nil
}

// GetRhsSettingsForBlockchainAndNetwork returns the rhs settings for blockchain and network
func (r *Resolver) GetRhsSettingsForBlockchainAndNetwork(ctx context.Context, blockchain, network string) (*RhsSettings, error) {
	resolverPrefixKey := fmt.Sprintf("%s:%s", blockchain, network)
	rhsSettings, err := r.GetRhsSettings(ctx, resolverPrefixKey)
	if err != nil {
		return nil, err
	}
	return rhsSettings, nil
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

// GetSupportedContracts returns the supported contracts
func (r *Resolver) GetSupportedContracts() map[string]*abi.State {
	return r.supportedContracts
}

func parseResolversSettings(_ context.Context, reader io.Reader) (ResolverSettings, error) {
	settings := ResolverSettings{}
	if err := yaml.NewDecoder(reader).Decode(&settings); err != nil {
		return nil, fmt.Errorf("invalid yaml file: %v", settings)
	}
	return settings, nil
}

// registerCustomDIDMethod registers custom DID methods
func registerCustomDIDMethod(ctx context.Context, blockchain string, network string, chainID string, method string, networkFlag byte) error {
	customChainID, err := strconv.Atoi(chainID)
	if err != nil {
		return fmt.Errorf("cannot convert chainID to int: %w", err)
	}
	params := core.DIDMethodNetworkParams{
		Method:      core.DIDMethod(method),
		Blockchain:  core.Blockchain(blockchain),
		Network:     core.NetworkID(network),
		NetworkFlag: networkFlag,
	}
	if err := core.RegisterDIDMethodNetwork(params, core.WithChainID(customChainID)); err != nil {
		log.Error(ctx, "cannot register custom DID method", "err", err, "customDID", chainID)
		return err
	}
	return nil
}
