package reverse_hash

import (
	"context"
	"errors"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	proof "github.com/iden3/merkletree-proof"
	proofEth "github.com/iden3/merkletree-proof/eth"
	proofHttp "github.com/iden3/merkletree-proof/http"

	"github.com/polygonid/issuer-node/internal/kms"
	"github.com/polygonid/issuer-node/pkg/blockchain/eth"
)

// RHSMode is a mode of RHS
type RHSMode string

const (
	// RHSModeOffChain is a mode when we use off-chain RHS
	RHSModeOffChain RHSMode = "OffChain"
	// RHSModeOnChain is a mode when we use on-chain RHS
	RHSModeOnChain RHSMode = "OnChain"
	// RHSModeNone is a mode when we don't use RHS
	RHSModeNone RHSMode = "None"
)

// Factory is a factory for creating RhsPublishers
type Factory struct {
	url                      string
	ethClient                *eth.Client
	onChainTreeStoreContract ethCommon.Address
	responseTimeout          time.Duration
}

// NewFactory creates new instance of Factory
func NewFactory(url string, ethClient *eth.Client, contract ethCommon.Address, rpcTimeout time.Duration) Factory {
	return Factory{
		url:                      url,
		ethClient:                ethClient,
		onChainTreeStoreContract: contract,
		responseTimeout:          rpcTimeout,
	}
}

// BuildPublishers creates new instance of RhsPublisher
func (f *Factory) BuildPublishers(ctx context.Context, rhsMode RHSMode, kmsKey *kms.KeyID) ([]RhsPublisher, error) {
	switch rhsMode {
	case RHSModeOffChain:
		rhsCli, err := f.initOffChainRHS()
		if err != nil {
			return nil, err
		}
		return []RhsPublisher{NewRhsPublisher(rhsCli, false)}, nil
	case RHSModeOnChain:
		onChainCli, err := f.initOnChainRHSCli(ctx, kmsKey)
		if err != nil {
			return nil, err
		}
		return []RhsPublisher{NewRhsPublisher(onChainCli, false)}, nil
	case RHSModeNone:
		return []RhsPublisher{}, nil
	default:
		return nil, errors.New("unknown rhs mode")
	}
}

func (f *Factory) initOffChainRHS() (proof.ReverseHashCli, error) {
	if f.url == "" {
		return nil, errors.New("rhs url must be configured")
	}
	return &proofHttp.ReverseHashCli{
		URL:         f.url,
		HTTPTimeout: f.responseTimeout,
	}, nil
}

func (f *Factory) initOnChainRHSCli(ctx context.Context, kmsKey *kms.KeyID) (proof.ReverseHashCli, error) {
	// TODO:
	// This can be a  problem in the future.
	// Since between counting the miner tip and using this transaction option can be a big time gap.
	// And while executing a transaction, we can have bigger tips on the network than we counted.
	txOpts, err := f.ethClient.CreateTxOpts(ctx, *kmsKey)
	if err != nil {
		return nil, err
	}
	cli, err := proofEth.NewReverseHashCli(f.onChainTreeStoreContract, f.ethClient.GetEthereumClient(), txOpts, f.responseTimeout)
	if err != nil {
		return nil, err
	}
	return cli, nil
}
