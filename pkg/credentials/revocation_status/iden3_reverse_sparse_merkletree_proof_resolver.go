package revocation_status

import (
	"fmt"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/pkg/network"
)

type iden3ReverseSparseMerkleTreeProofResolver struct{}

func (r *iden3ReverseSparseMerkleTreeProofResolver) resolve(credentialStatusSettings network.RhsSettings, _ w3c.DID, nonce uint64, issuerState string) *verifiable.CredentialStatus {
	return &verifiable.CredentialStatus{
		ID:              buildRHSRevocationURL(*credentialStatusSettings.RhsUrl, issuerState),
		Type:            verifiable.Iden3ReverseSparseMerkleTreeProof,
		RevocationNonce: nonce,
		StatusIssuer: &verifiable.CredentialStatus{
			ID:              fmt.Sprintf("%s/v2/agent", credentialStatusSettings.Iden3CommAgentStatus),
			Type:            verifiable.Iden3commRevocationStatusV1,
			RevocationNonce: nonce,
		},
	}
}
