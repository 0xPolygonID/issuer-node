package revocation_status

import (
	"fmt"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/network"
)

type iden3CommRevocationStatusV1Resolver struct{}

func (r *iden3CommRevocationStatusV1Resolver) resolve(credentialStatusSettings network.RhsSettings, _ w3c.DID, nonce uint64, _ string) *verifiable.CredentialStatus {
	return &verifiable.CredentialStatus{
		ID:              fmt.Sprintf("%s/v2/agent", credentialStatusSettings.Iden3CommAgentStatus),
		Type:            verifiable.Iden3commRevocationStatusV1,
		RevocationNonce: nonce,
	}
}
