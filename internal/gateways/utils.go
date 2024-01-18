package gateways

import (
	"context"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
	"github.com/polygonid/sh-id-platform/pkg/network"
)

func getEthClient(ctx context.Context, identity *domain.Identity, resolver network.Resolver) (*eth.Client, error) {
	resolverPrefix, err := identity.GetResolverPrefix()
	if err != nil {
		log.Error(ctx, "failed to get networkResolver prefix", "err", err)
		return nil, err
	}

	client, err := resolver.GetEthClient(resolverPrefix)
	if err != nil {
		log.Error(ctx, "failed to get client", "err", err)
		return nil, err
	}

	return client, nil
}
