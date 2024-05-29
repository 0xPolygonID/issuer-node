package revocation_status

import (
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/config"
)

type sparseMerkleTreeProofResolver struct{}

func (r *sparseMerkleTreeProofResolver) resolve(credentialStatusSettings config.CredentialStatus, _ w3c.DID, nonce uint64, _ string) *verifiable.CredentialStatus {
	return &verifiable.CredentialStatus{
		ID:              buildDirectRevocationURL(credentialStatusSettings.Iden3CommAgentStatus.GetURL()),
		Type:            verifiable.SparseMerkleTreeProof,
		RevocationNonce: nonce,
	}
}
