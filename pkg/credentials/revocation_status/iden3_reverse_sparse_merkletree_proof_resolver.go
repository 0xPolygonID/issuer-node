package revocation_status

import (
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/pkg/network"
)

type iden3ReverseSparseMerkleTreeProofResolver struct{}

func (r *iden3ReverseSparseMerkleTreeProofResolver) resolve(credentialStatusSettings network.RhsSettings, issuerDID w3c.DID, nonce uint64, issuerState string) *verifiable.CredentialStatus {
	return &verifiable.CredentialStatus{
		ID:              buildRHSRevocationURL(*credentialStatusSettings.RhsUrl, issuerState),
		Type:            verifiable.Iden3ReverseSparseMerkleTreeProof,
		RevocationNonce: nonce,
		StatusIssuer: &verifiable.CredentialStatus{
			ID:              buildDirectRevocationURL(credentialStatusSettings.DirectUrl, issuerDID.String(), nonce, credentialStatusSettings.SingleIssuer),
			Type:            verifiable.SparseMerkleTreeProof,
			RevocationNonce: nonce,
		},
	}
}
