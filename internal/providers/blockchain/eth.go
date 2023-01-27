package blockchain

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
)

// InitEthClient returns a State Contract Instance
func InitEthClient(ethURL, contractAddress string) (*eth.State, error) {
	ec, err := ethclient.Dial(ethURL)
	if err != nil {
		return nil, fmt.Errorf("failed connect to eth node %s: %s", ethURL, err.Error())
	}
	stateContractInstance, err := eth.NewState(common.HexToAddress(contractAddress), ec)
	if err != nil {
		return nil, fmt.Errorf("error failed create state contract client: %s", err.Error())
	}

	return stateContractInstance, nil
}

// InitEthConnect opens a new eth connection
func InitEthConnect(cfg config.Ethereum) (*eth.Client, error) {
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
	})

	return cl, nil
}
