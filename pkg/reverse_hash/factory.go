package reverse_hash

import (
	"context"
	"errors"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	proof "github.com/iden3/merkletree-proof"
	proofEth "github.com/iden3/merkletree-proof/eth"
	proofHttp "github.com/iden3/merkletree-proof/http"

	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/network"
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
	responseTimeout time.Duration
	networkResolver network.Resolver
}

// NewFactory creates new instance of Factory
func NewFactory(networkResolver network.Resolver, rpcTimeout time.Duration) Factory {
	return Factory{
		networkResolver: networkResolver,
		responseTimeout: rpcTimeout,
	}
}

// BuildPublishers creates new instance of RhsPublisher
func (f *Factory) BuildPublishers(ctx context.Context, resolverPrefix string, kmsKey *kms.KeyID) ([]RhsPublisher, error) {
	rhsSettings, err := f.networkResolver.GetRhsSettings(resolverPrefix)
	if err != nil {
		return nil, err
	}
	rhsMode := RHSMode(rhsSettings.Mode)
	switch rhsMode {
	case RHSModeOffChain:
		rhsCli, err := f.initOffChainRHS(resolverPrefix)
		if err != nil {
			return nil, err
		}
		return []RhsPublisher{NewRhsPublisher(rhsCli, false)}, nil
	case RHSModeOnChain:
		onChainCli, err := f.initOnChainRHSCli(ctx, resolverPrefix, kmsKey)
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

func (f *Factory) initOffChainRHS(resolverPrefix string) (proof.ReverseHashCli, error) {
	rhsSettings, err := f.networkResolver.GetRhsSettings(resolverPrefix)
	if err != nil {
		return nil, err
	}
	if rhsSettings.RhsUrl == nil || *rhsSettings.RhsUrl == "" {
		return nil, errors.New("rhs url must be configured")
	}
	return &proofHttp.ReverseHashCli{
		URL:         *rhsSettings.RhsUrl,
		HTTPTimeout: f.responseTimeout,
	}, nil
}

func (f *Factory) initOnChainRHSCli(ctx context.Context, resolverPrefix string, kmsKey *kms.KeyID) (proof.ReverseHashCli, error) {
	// TODO:
	// This can be a  problem in the future.
	// Since between counting the miner tip and using this transaction option can be a big time gap.
	// And while executing a transaction, we can have bigger tips on the network than we counted.
	rhsSettings, err := f.networkResolver.GetRhsSettings(resolverPrefix)
	if err != nil {
		return nil, err
	}

	ethClient, err := f.networkResolver.GetEthClient(resolverPrefix)
	if err != nil {
		log.Error(ctx, "failed to get eth client", "err", err)
		return nil, err
	}

	txOpts, err := ethClient.CreateTxOpts(ctx, *kmsKey)
	if err != nil {
		return nil, err
	}

	if rhsSettings.ContractAddress == nil || *rhsSettings.ContractAddress == "" {
		return nil, errors.New("rhs contract address must be configured")
	}

	contractAddress := ethCommon.HexToAddress(*rhsSettings.ContractAddress)
	cli, err := proofEth.NewReverseHashCli(contractAddress, ethClient.GetEthereumClient(), txOpts, f.responseTimeout)
	if err != nil {
		return nil, err
	}
	return cli, nil
}
