package blockchain

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iden3/contracts-abi/state/go/abi"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
)

// InitEthClient returns a State Contract Instance
func InitEthClient(addresses map[string]string, rpcs map[string]string) (map[string]*abi.State, error) {
	stateContracts := make(map[string]*abi.State, len(addresses))
	for chainID, address := range addresses {
		if _, ok := rpcs[chainID]; !ok {
			return nil, fmt.Errorf("rpc url for chain '%s' not found", chainID)
		}
		ec, err := ethclient.Dial(rpcs[chainID])
		if err != nil {
			return nil, fmt.Errorf("failed connect to eth node '%s': %v", rpcs[chainID], err)
		}
		stateContract, err := abi.NewState(common.HexToAddress(address), ec)
		if err != nil {
			return nil, fmt.Errorf(
				"error failed create state contract client for contract '%s' and rpc '%s': %v", address, rpcs[chainID], err)
		}
		stateContracts[chainID] = stateContract
	}
	return stateContracts, nil
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
		GasLess:                cfg.GasLess,
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
		GasLess:                cfg.Ethereum.GasLess,
		RPCResponseTimeout:     cfg.Ethereum.RPCResponseTimeout,
		WaitReceiptCycleTime:   cfg.Ethereum.WaitReceiptCycleTime,
		WaitBlockCycleTime:     cfg.Ethereum.WaitBlockCycleTime,
	}, kms), nil
}
