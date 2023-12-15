package services

import (
	"context"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/issuer-node/internal/config"
	"github.com/polygonid/issuer-node/internal/kms"
	"github.com/polygonid/issuer-node/internal/log"
	"github.com/polygonid/issuer-node/pkg/blockchain/eth"
)

// AccountService is a service for account operations
type AccountService struct {
	rpcURL    string
	kms       *kms.KMS
	ethClient *eth.Client
}

// NewAccountService returns new account service
func NewAccountService(ethConfig config.Ethereum, keyStore *kms.KMS) *AccountService {
	commonClient, err := ethclient.Dial(ethConfig.URL)
	if err != nil {
		log.Warn(context.Background(), "cannot init eth client", "err", err)
	}
	ethClient := eth.NewClient(commonClient, &eth.ClientConfig{
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

	return &AccountService{
		rpcURL:    ethConfig.URL,
		kms:       keyStore,
		ethClient: ethClient,
	}
}

// GetBalanceByDID returns balance by DID
func (as *AccountService) GetBalanceByDID(ctx context.Context, did *w3c.DID) (*big.Int, error) {
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
	return as.ethClient.BalanceAt(ctx, commonAddress)
}
