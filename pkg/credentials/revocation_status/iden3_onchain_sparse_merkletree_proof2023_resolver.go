package revocation_status

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/pkg/network"
)

type iden3OnChainSparseMerkleTreeProof2023Resolver struct{}

func (r *iden3OnChainSparseMerkleTreeProof2023Resolver) resolve(credentialStatusSettings network.RhsSettings, issuerDID w3c.DID, nonce uint64, issuerState string) *verifiable.CredentialStatus {
	contractAddressHex := *credentialStatusSettings.ContractAddress
	return &verifiable.CredentialStatus{
		ID:              buildIden3OnchainSMTProofURL(issuerDID, nonce, ethcommon.HexToAddress(contractAddressHex), *credentialStatusSettings.ChainID, issuerState),
		Type:            verifiable.Iden3OnchainSparseMerkleTreeProof2023,
		RevocationNonce: nonce,
	}
}
