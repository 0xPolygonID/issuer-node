package revocation_status

import (
	"context"
	"testing"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/pkg/helpers"
	"github.com/polygonid/sh-id-platform/pkg/network"
)

func TestRevocationStatusResolver_GetCredentialRevocationStatus(t *testing.T) {
	const did = "did:polygonid:polygon:amoy:2qSuD8ZDpsAG3s8WJjwzqhMsqGLz8RUG1BHVUe3Gwu"
	didW3c, err := w3c.ParseDID(did)
	require.NoError(t, err)

	type expected struct {
		err error
		*verifiable.CredentialStatus
	}

	type testConfig struct {
		name                     string
		credentialStatusSettings network.RhsSettings
		credentialStatusType     verifiable.CredentialStatusType
		nonce                    uint64
		issuerState              string
		expected                 expected
	}

	for _, tc := range []testConfig{
		{
			name: "Iden3ReverseSparseMerkleTreeProof for single issuer",
			credentialStatusSettings: network.RhsSettings{
				Mode:                 network.OffChain,
				Iden3CommAgentStatus: "https://issuernode",
				SingleIssuer:         true,
			},
			credentialStatusType: verifiable.Iden3ReverseSparseMerkleTreeProof,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.Iden3ReverseSparseMerkleTreeProof,
					ID:              "https://rhs-staging.polygonid.me/node?state=issuer-state",
					RevocationNonce: 12345,
					StatusIssuer: &verifiable.CredentialStatus{
						Type:            verifiable.Iden3commRevocationStatusV1,
						ID:              "https://issuer-node.privado.id/v1/agent",
						RevocationNonce: 12345,
					},
				},
			},
		},
		{
			name: "Iden3ReverseSparseMerkleTreeProof for multiples issuers",
			credentialStatusSettings: network.RhsSettings{
				Mode:                 network.OffChain,
				Iden3CommAgentStatus: "https://issuernode",
				SingleIssuer:         true,
			},
			credentialStatusType: verifiable.Iden3ReverseSparseMerkleTreeProof,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.Iden3ReverseSparseMerkleTreeProof,
					ID:              "https://rhs-staging.polygonid.me/node?state=issuer-state",
					RevocationNonce: 12345,
					StatusIssuer: &verifiable.CredentialStatus{
						Type:            verifiable.Iden3commRevocationStatusV1,
						ID:              "https://issuer-node.privado.id/v1/agent",
						RevocationNonce: 12345,
					},
				},
			},
		},
		{
			name: "Iden3OnchainSparseMerkleTreeProof2023 for single issuer",
			credentialStatusSettings: network.RhsSettings{
				Mode:                 network.OnChain,
				Iden3CommAgentStatus: "https://issuernode",
				SingleIssuer:         true,
				RhsUrl:               common.ToPointer("https://rhs"),
				ContractAddress:      common.ToPointer("0x1234567890"),
				PublishingKey:        "pbkey",
				ChainID:              common.ToPointer("80002"),
			},
			credentialStatusType: verifiable.Iden3OnchainSparseMerkleTreeProof2023,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.Iden3OnchainSparseMerkleTreeProof2023,
					ID:              "did:polygonid:polygon:amoy:2qSuD8ZDpsAG3s8WJjwzqhMsqGLz8RUG1BHVUe3Gwu/credentialStatus?revocationNonce=12345&contractAddress=80002:0x3d3763eC0a50CE1AdF83d0b5D99FBE0e3fEB43fb&state=issuer-state",
					RevocationNonce: 12345,
				},
			},
		},
		{
			name: "Iden3OnchainSparseMerkleTreeProof2023 for multiples issuers",
			credentialStatusSettings: network.RhsSettings{
				Mode:                 network.OnChain,
				Iden3CommAgentStatus: "https://issuernode",
				SingleIssuer:         true,
				RhsUrl:               common.ToPointer("https://rhs"),
				ContractAddress:      common.ToPointer("0x1234567890"),
				PublishingKey:        "pbkey",
				ChainID:              common.ToPointer("80002"),
			},
			credentialStatusType: verifiable.Iden3OnchainSparseMerkleTreeProof2023,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.Iden3OnchainSparseMerkleTreeProof2023,
					ID:              "did:polygonid:polygon:amoy:2qSuD8ZDpsAG3s8WJjwzqhMsqGLz8RUG1BHVUe3Gwu/credentialStatus?revocationNonce=12345&contractAddress=80002:0x3d3763eC0a50CE1AdF83d0b5D99FBE0e3fEB43fb&state=issuer-state",
					RevocationNonce: 12345,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Configuration{
				ServerUrl:           "https://issuer-node.privado.id",
				NetworkResolverPath: "",
				Ethereum: config.Ethereum{
					URL:            "https://polygon-mumbai.g.alchemy.com/v2/xaP2_",
					ResolverPrefix: "polygon:mumbai",
				},
			}
			networkResolver, err := network.NewResolver(context.Background(), *cfg, nil, helpers.CreateFile(t))
			require.NoError(t, err)
			rsr := NewRevocationStatusResolver(*networkResolver)
			credentialStatus, err := rsr.GetCredentialRevocationStatus(context.Background(), *didW3c, tc.nonce, tc.issuerState, tc.credentialStatusType)
			require.Equal(t, tc.expected.CredentialStatus, credentialStatus)
			require.NoError(t, err)
		})
	}
}
