package services

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
)

// AccountService is a service for account operations
type AccountService struct {
	ethConfig  config.Ethereum
	kms        *kms.KMS
	ethClients map[string]*eth.Client
}

const keyPairLen = 2

// NewAccountService returns new account service
func NewAccountService(ethConfig config.Ethereum, keyStore *kms.KMS) (*AccountService, error) {
	ethClients, err := InitEthClientsForAllSupportedChains(ethConfig, keyStore)
	if err != nil {
		log.Error(context.Background(), "cannot init eth clients", "err", err)
		return nil, err
	}
	return &AccountService{
		ethConfig:  ethConfig,
		kms:        keyStore,
		ethClients: ethClients,
	}, nil
}

// InitEthClientsForAllSupportedChains returns eth clients for all supported chains
func InitEthClientsForAllSupportedChains(ethConfig config.Ethereum, keyStore *kms.KMS) (map[string]*eth.Client, error) {
	supportedRpc, err := decode(ethConfig.SupportedRPC)
	if err != nil {
		log.Error(context.Background(), "cannot decode supported rpc", "err", err)
		return nil, err
	}
	clients := make(map[string]*eth.Client)
	for chainName, rpcURL := range supportedRpc {
		commonClient, err := ethclient.Dial(rpcURL)
		if err != nil {
			return nil, err
		}
		client := eth.NewClient(commonClient, &eth.ClientConfig{
			DefaultGasLimit:        ethConfig.DefaultGasLimit,
			ConfirmationTimeout:    ethConfig.ConfirmationTimeout,
			ConfirmationBlockCount: ethConfig.ConfirmationBlockCount,
			ReceiptTimeout:         ethConfig.ReceiptTimeout,
			MinGasPrice:            big.NewInt(int64(ethConfig.MinGasPrice)),
			MaxGasPrice:            big.NewInt(int64(ethConfig.MaxGasPrice)),
			RPCResponseTimeout:     ethConfig.RPCResponseTimeout,
			WaitReceiptCycleTime:   ethConfig.WaitReceiptCycleTime,
			WaitBlockCycleTime:     ethConfig.WaitBlockCycleTime,
		}, keyStore)

		clients[chainName] = client

	}
	return clients, nil
}

func decode(value string) (map[string]string, error) {
	contracts := make(map[string]string)
	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		kvpair := strings.Split(pair, "=")
		if len(kvpair) != keyPairLen {
			return contracts, fmt.Errorf("invalid map item: %q", pair)
		}
		contracts[kvpair[0]] = kvpair[1]

	}
	return contracts, nil
}

// GetBalanceByDID returns balance by DID
func (as *AccountService) GetBalanceByDID(ctx context.Context, did *w3c.DID) (*big.Int, error) {
	ethClient, err := as.GetEthClientForDID(ctx, did)
	if err != nil {
		log.Error(ctx, "cannot get eth client for DID", "err", err)
		return nil, err
	}
	id, err := core.IDFromDID(*did)
	if err != nil {
		log.Error(ctx, "cannot get id from DID", "err", err)
		return nil, err
	}
	ethAddress, err := core.EthAddressFromID(id)
	if err != nil {
		log.Error(ctx, "cannot get eth address from id", "err", err)
		return nil, err
	}
	commonAddress := ethCommon.BytesToAddress(ethAddress[:])
	return ethClient.BalanceAt(ctx, commonAddress)
}

// GetEthClientForDID returns eth client for chain mapped from DID
func (as *AccountService) GetEthClientForDID(ctx context.Context, did *w3c.DID) (*eth.Client, error) {
	chainName, err := common.ChainIDfromDID(*did)
	if err != nil {
		log.Error(ctx, "cannot get chain id from DID", "err", err, "did", did)
		return nil, err
	}

	ethClient, ok := as.ethClients[chainName]
	if !ok {
		return nil, fmt.Errorf("chain id is not registered for network %s", chainName)
	}
	return ethClient, nil
}
