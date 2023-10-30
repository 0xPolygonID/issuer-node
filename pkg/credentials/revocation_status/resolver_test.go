package revocation_status

import (
	"context"
	"testing"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/config"
)

func TestRevocationStatusResolver_GetCredentialRevocationStatus(t *testing.T) {
	const did = "did:polygonid:polygon:mumbai:2qFbNk3Vz7Uy3ryq6zjwkC7p7RbLTfRpMsy6axjxeG"
	didW3c, err := w3c.ParseDID(did)
	require.NoError(t, err)

	type expected struct {
		err error
		*verifiable.CredentialStatus
	}

	type testConfig struct {
		name                     string
		credentialStatusSettings config.CredentialStatus
		credentialStatusType     verifiable.CredentialStatusType
		nonce                    uint64
		issuerState              string
		expected                 expected
	}

	for _, tc := range []testConfig{
		{
			name: "SparseMerkleTreeProof for single issuer",
			credentialStatusSettings: config.CredentialStatus{
				RHSMode: config.RHSMode("none"),
				DirectStatus: config.DirectStatus{
					URL: "https://issuernode",
				},
				SingleIssuer: true,
			},
			credentialStatusType: verifiable.SparseMerkleTreeProof,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.SparseMerkleTreeProof,
					ID:              "https://issuernode/v1/credentials/revocation/status/12345",
					RevocationNonce: 12345,
				},
			},
		},
		{
			name: "SparseMerkleTreeProof for multiples issuers",
			credentialStatusSettings: config.CredentialStatus{
				RHSMode: config.RHSMode("none"),
				DirectStatus: config.DirectStatus{
					URL: "https://issuernode",
				},
				SingleIssuer: false,
			},
			credentialStatusType: verifiable.SparseMerkleTreeProof,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.SparseMerkleTreeProof,
					ID:              "https://issuernode/v1/did%3Apolygonid%3Apolygon%3Amumbai%3A2qFbNk3Vz7Uy3ryq6zjwkC7p7RbLTfRpMsy6axjxeG/claims/revocation/status/12345",
					RevocationNonce: 12345,
				},
			},
		},
		{
			name: "Iden3ReverseSparseMerkleTreeProof for single issuer",
			credentialStatusSettings: config.CredentialStatus{
				RHSMode: config.RHSMode("OffChain"),
				RHS: config.RHS{
					URL: "https://rhs",
				},
				DirectStatus: config.DirectStatus{
					URL: "https://issuernode",
				},
				SingleIssuer: true,
			},
			credentialStatusType: verifiable.Iden3ReverseSparseMerkleTreeProof,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.Iden3ReverseSparseMerkleTreeProof,
					ID:              "https://rhs/node?state=issuer-state",
					RevocationNonce: 12345,
					StatusIssuer: &verifiable.CredentialStatus{
						Type:            verifiable.SparseMerkleTreeProof,
						ID:              "https://issuernode/v1/credentials/revocation/status/12345",
						RevocationNonce: 12345,
					},
				},
			},
		},
		{
			name: "Iden3ReverseSparseMerkleTreeProof for multiples issuers",
			credentialStatusSettings: config.CredentialStatus{
				RHSMode: config.RHSMode("OffChain"),
				RHS: config.RHS{
					URL: "https://rhs",
				},
				DirectStatus: config.DirectStatus{
					URL: "https://issuernode",
				},
				SingleIssuer: false,
			},
			credentialStatusType: verifiable.Iden3ReverseSparseMerkleTreeProof,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.Iden3ReverseSparseMerkleTreeProof,
					ID:              "https://rhs/node?state=issuer-state",
					RevocationNonce: 12345,
					StatusIssuer: &verifiable.CredentialStatus{
						Type:            verifiable.SparseMerkleTreeProof,
						ID:              "https://issuernode/v1/did%3Apolygonid%3Apolygon%3Amumbai%3A2qFbNk3Vz7Uy3ryq6zjwkC7p7RbLTfRpMsy6axjxeG/claims/revocation/status/12345",
						RevocationNonce: 12345,
					},
				},
			},
		},
		{
			name: "Iden3OnchainSparseMerkleTreeProof2023 for single issuer",
			credentialStatusSettings: config.CredentialStatus{
				RHSMode: config.RHSMode("OffChain"),
				RHS: config.RHS{
					URL: "https://rhs",
				},
				DirectStatus: config.DirectStatus{
					URL: "https://issuernode",
				},
				SingleIssuer: false,
				OnchainTreeStore: config.OnchainTreeStore{
					SupportedTreeStoreContract: "0x1234567890",
					PublishingKeyPath:          "pbkey",
					ChainID:                    "80001",
				},
			},
			credentialStatusType: verifiable.Iden3OnchainSparseMerkleTreeProof2023,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.Iden3OnchainSparseMerkleTreeProof2023,
					ID:              "did:polygonid:polygon:mumbai:2qFbNk3Vz7Uy3ryq6zjwkC7p7RbLTfRpMsy6axjxeG/credentialStatus?revocationNonce=12345&contractAddress=80001:0x0000000000000000000000000000001234567890&state=issuer-state",
					RevocationNonce: 12345,
				},
			},
		},
		{
			name: "Iden3OnchainSparseMerkleTreeProof2023 for multiples issuers",
			credentialStatusSettings: config.CredentialStatus{
				RHSMode: config.RHSMode("OffChain"),
				RHS: config.RHS{
					URL: "https://rhs",
				},
				DirectStatus: config.DirectStatus{
					URL: "https://issuernode",
				},
				SingleIssuer: false,
				OnchainTreeStore: config.OnchainTreeStore{
					SupportedTreeStoreContract: "0x1234567890",
					PublishingKeyPath:          "pbkey",
					ChainID:                    "80001",
				},
			},
			credentialStatusType: verifiable.Iden3OnchainSparseMerkleTreeProof2023,
			nonce:                12345,
			issuerState:          "issuer-state",
			expected: expected{
				err: nil,
				CredentialStatus: &verifiable.CredentialStatus{
					Type:            verifiable.Iden3OnchainSparseMerkleTreeProof2023,
					ID:              "did:polygonid:polygon:mumbai:2qFbNk3Vz7Uy3ryq6zjwkC7p7RbLTfRpMsy6axjxeG/credentialStatus?revocationNonce=12345&contractAddress=80001:0x0000000000000000000000000000001234567890&state=issuer-state",
					RevocationNonce: 12345,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rsr := NewRevocationStatusResolver(tc.credentialStatusSettings)
			credentialStatus, err := rsr.GetCredentialRevocationStatus(context.Background(), *didW3c, tc.nonce, tc.issuerState, tc.credentialStatusType)
			require.Equal(t, tc.expected.CredentialStatus, credentialStatus)
			require.NoError(t, err)
		})
	}
}
