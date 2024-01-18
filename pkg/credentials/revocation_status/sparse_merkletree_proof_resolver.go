package revocation_status

import (
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/pkg/network"
)

type sparseMerkleTreeProofResolver struct{}

func (r *sparseMerkleTreeProofResolver) resolve(credentialStatusSettings network.RhsSettings, issuerDID w3c.DID, nonce uint64, issuerState string) *verifiable.CredentialStatus {
	return &verifiable.CredentialStatus{
		ID:              buildDirectRevocationURL(credentialStatusSettings.DirectUrl, issuerDID.String(), nonce, credentialStatusSettings.SingleIssuer),
		Type:            verifiable.SparseMerkleTreeProof,
		RevocationNonce: nonce,
	}
}
