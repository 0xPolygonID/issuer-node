package config

import (
	"strings"

	"github.com/iden3/go-schema-processor/v2/verifiable"
)

const (
	iden3commRevocationStatusV1           = verifiable.Iden3commRevocationStatusV1
	iden3ReverseSparseMerkleTreeProof     = "Iden3ReverseSparseMerkleTreeProof"
	iden3OnchainSparseMerkleTreeProof2023 = "Iden3OnchainSparseMerkleTreeProof2023"
	onChain                               = "OnChain"
	offChain                              = "OffChain"
	none                                  = "None"
)

// RHSMode is a mode of RHS
type RHSMode string

// CredentialStatus is the type of credential status
type CredentialStatus struct {
	Iden3CommAgentStatus Iden3CommAgentStatus
	RHS                  RHS
	OnchainTreeStore     OnchainTreeStore `mapstructure:"OnchainTreeStore"`
	RHSMode              RHSMode          `tip:"Reverse hash service mode (OffChain, OnChain, None)"`
	SingleIssuer         bool
	CredentialStatusType verifiable.CredentialStatusType `mapstructure:"CredentialStatusType" default:"Iden3commRevocationStatusV1"`
}

// Iden3CommAgentStatus is the type of direct status
type Iden3CommAgentStatus struct {
	URL string `mapstructure:"URL"`
}

// GetURL returns the URL of the direct status
func (r *Iden3CommAgentStatus) GetURL() string {
	return strings.TrimSuffix(r.URL, "/")
}

// RHS is the type of RHS
type RHS struct {
	URL string `mapstructure:"URL"`
}

// GetURL returns the URL of the RHS
func (r *RHS) GetURL() string {
	return strings.TrimSuffix(r.URL, "/")
}

// DIDResolver is the type of DID resolver
type DIDResolver struct {
	URL string `mapstructure:"URL"`
}

// GetURL returns the URL of the DID resolver
func (r *DIDResolver) GetURL() string {
	return strings.TrimSuffix(r.URL, "/")
}

// OnchainTreeStore is the type of onchain tree store
type OnchainTreeStore struct {
	SupportedTreeStoreContract string `mapstructure:"SupportedTreeStoreContract"`
	PublishingKeyPath          string `mapstructure:"PublishingKeyPath" default:"pbkey"`
	ChainID                    string `mapstructure:"ChainID"`
}
