package reverse_hash

import (
	"context"
	"errors"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	proof "github.com/iden3/merkletree-proof"
	proofEth "github.com/iden3/merkletree-proof/eth"
	proofHttp "github.com/iden3/merkletree-proof/http"

	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/network"
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
func (f *Factory) BuildPublishers(ctx context.Context, resolverPrefix string, credentialStatusType *verifiable.CredentialStatusType, kmsKey *kms.KeyID) ([]RhsPublisher, error) {

	if credentialStatusType == nil {
		rhsSettings, err := f.networkResolver.GetRhsSettings(ctx, resolverPrefix)
		if err != nil {
			return nil, err
		}

		credentialStatusType = &rhsSettings.DefaultAuthBJJCredentialStatus
	}

	switch *credentialStatusType {
	case verifiable.Iden3ReverseSparseMerkleTreeProof:
		rhsCli, err := f.initOffChainRHS(ctx, resolverPrefix)
		if err != nil {
			return nil, err
		}
		return []RhsPublisher{NewRhsPublisher(rhsCli, false)}, nil
	case verifiable.Iden3OnchainSparseMerkleTreeProof2023:
		onChainCli, err := f.initOnChainRHSCli(ctx, resolverPrefix, kmsKey)
		if err != nil {
			return nil, err
		}
		return []RhsPublisher{NewRhsPublisher(onChainCli, false)}, nil
	case verifiable.Iden3commRevocationStatusV1:
		return []RhsPublisher{}, nil
	default:
		return nil, errors.New("unknown credential status type")
	}
}

func (f *Factory) initOffChainRHS(ctx context.Context, resolverPrefix string) (proof.ReverseHashCli, error) {
	rhsSettings, err := f.networkResolver.GetRhsSettings(ctx, resolverPrefix)
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
	rhsSettings, err := f.networkResolver.GetRhsSettings(ctx, resolverPrefix)
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
		log.Error(ctx, "failed to create tx opts", "err", err)
		return nil, err
	}

	if rhsSettings.ContractAddress == nil || *rhsSettings.ContractAddress == "" {
		return nil, errors.New("rhs contract address must be configured")
	}

	contractAddress := ethCommon.HexToAddress(*rhsSettings.ContractAddress)
	cli, err := proofEth.NewReverseHashCli(ethClient.GetEthereumClient(), contractAddress, txOpts.From, txOpts.Signer)
	if err != nil {
		log.Error(ctx, "failed to create on-chain rhs client", "err", err)
		return nil, err
	}
	return cli, nil
}
