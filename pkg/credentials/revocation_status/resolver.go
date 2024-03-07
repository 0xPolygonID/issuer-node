package revocation_status

import (
	"context"
	"errors"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/config"
)

const resolversLength = 3

type revocationCredentialStatusResolver interface {
	resolve(credentialStatusSettings config.CredentialStatus, issuerDID w3c.DID, nonce uint64, issuerState string) *verifiable.CredentialStatus
}

// RevocationStatusResolver resolves credential status.
type RevocationStatusResolver struct {
	credentialStatusSettings config.CredentialStatus
	resolvers                map[verifiable.CredentialStatusType]revocationCredentialStatusResolver
}

// NewRevocationStatusResolver - constructor
func NewRevocationStatusResolver(credentialStatusSettings config.CredentialStatus) *RevocationStatusResolver {
	resolvers := make(map[verifiable.CredentialStatusType]revocationCredentialStatusResolver, resolversLength)
	resolvers[verifiable.Iden3ReverseSparseMerkleTreeProof] = &iden3ReverseSparseMerkleTreeProofResolver{}
	resolvers[verifiable.SparseMerkleTreeProof] = &sparseMerkleTreeProofResolver{}
	resolvers[verifiable.Iden3commRevocationStatusV1] = &iden3CommRevocationStatusV1Resolver{}
	resolvers[verifiable.Iden3OnchainSparseMerkleTreeProof2023] = &iden3OnChainSparseMerkleTreeProof2023Resolver{}
	return &RevocationStatusResolver{
		credentialStatusSettings: credentialStatusSettings,
		resolvers:                resolvers,
	}
}

// GetCredentialRevocationStatus - return a way to check credential revocation status.
// If status is not supported, an error is returned.
// If status is supported, a way to check revocation status is returned.
func (rsr *RevocationStatusResolver) GetCredentialRevocationStatus(_ context.Context, issuerDID w3c.DID, nonce uint64, issuerState string, credentialStatusType verifiable.CredentialStatusType) (*verifiable.CredentialStatus, error) {
	if credentialStatusType == "" {
		credentialStatusType = verifiable.Iden3commRevocationStatusV1
	}
	resolver, ok := rsr.resolvers[credentialStatusType]
	if !ok {
		return nil, errors.New("unsupported credential credentialStatusType type")
	}
	return resolver.resolve(rsr.credentialStatusSettings, issuerDID, nonce, issuerState), nil
}
