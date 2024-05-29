package revocation_status

import (
	"fmt"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/config"
)

type iden3CommRevocationStatusV1Resolver struct{}

func (r *iden3CommRevocationStatusV1Resolver) resolve(credentialStatusSettings config.CredentialStatus, issuerDID w3c.DID, nonce uint64, issuerState string) *verifiable.CredentialStatus {
	return &verifiable.CredentialStatus{
		ID:              fmt.Sprintf("%s/v1/agent", credentialStatusSettings.Iden3CommAgentStatus.GetURL()),
		Type:            verifiable.Iden3commRevocationStatusV1,
		RevocationNonce: nonce,
	}
}
