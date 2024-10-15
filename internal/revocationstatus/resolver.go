package revocationstatus

import (
	"context"
	"errors"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/network"
)

const resolversLength = 3

type revocationCredentialStatusResolver interface {
	resolve(credentialStatusSettings network.RhsSettings, issuerDID w3c.DID, nonce uint64, issuerState string) *verifiable.CredentialStatus
}

// Resolver resolves credential status.
type Resolver struct {
	networkResolver network.Resolver
	resolvers       map[verifiable.CredentialStatusType]revocationCredentialStatusResolver
}

// NewRevocationStatusResolver - constructor
func NewRevocationStatusResolver(networkResolver network.Resolver) *Resolver {
	resolvers := make(map[verifiable.CredentialStatusType]revocationCredentialStatusResolver, resolversLength)
	resolvers[verifiable.Iden3ReverseSparseMerkleTreeProof] = &iden3ReverseSparseMerkleTreeProofResolver{}
	resolvers[verifiable.Iden3commRevocationStatusV1] = &iden3CommRevocationStatusV1Resolver{}
	resolvers[verifiable.Iden3OnchainSparseMerkleTreeProof2023] = &iden3OnChainSparseMerkleTreeProof2023Resolver{}
	return &Resolver{
		networkResolver: networkResolver,
		resolvers:       resolvers,
	}
}

// GetCredentialRevocationStatus - return a way to check credential revocation status.
// If status is not supported, an error is returned.
// If status is supported, a way to check revocation status is returned.
func (rsr *Resolver) GetCredentialRevocationStatus(ctx context.Context, issuerDID w3c.DID, nonce uint64, issuerState string, credentialStatusType verifiable.CredentialStatusType) (*verifiable.CredentialStatus, error) {
	if credentialStatusType == "" {
		credentialStatusType = verifiable.Iden3commRevocationStatusV1
	}
	resolver, ok := rsr.resolvers[credentialStatusType]
	if !ok {
		return nil, errors.New("unsupported credential credentialStatusType type")
	}

	resolverPrefix, err := common.ResolverPrefix(&issuerDID)
	if err != nil {
		return nil, err
	}

	settings, err := rsr.networkResolver.GetRhsSettings(ctx, resolverPrefix)
	if err != nil {
		return nil, err
	}

	return resolver.resolve(*settings, issuerDID, nonce, issuerState), nil
}
