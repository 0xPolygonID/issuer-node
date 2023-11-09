package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/kms"
)

// DIDCreationOptions represents options for DID creation
type DIDCreationOptions struct {
	Method                  core.DIDMethod                  `json:"method"`
	Blockchain              core.Blockchain                 `json:"blockchain"`
	Network                 core.NetworkID                  `json:"network"`
	KeyType                 kms.KeyType                     `json:"keyType"`
	AuthBJJCredentialStatus verifiable.CredentialStatusType `json:"authBJJCredentialStatus,omitempty"`
}

// IdentityService is the interface implemented by the identity service
type IdentityService interface {
	GetByDID(ctx context.Context, identifier w3c.DID) (*domain.Identity, error)
	Create(ctx context.Context, hostURL string, didOptions *DIDCreationOptions) (*domain.Identity, error)
	SignClaimEntry(ctx context.Context, authClaim *domain.Claim, claimEntry *core.Claim) (*verifiable.BJJSignatureProof2021, error)
	Get(ctx context.Context) (identities []string, err error)
	UpdateState(ctx context.Context, did w3c.DID) (*domain.IdentityState, error)
	Exists(ctx context.Context, identifier w3c.DID) (bool, error)
	GetLatestStateByID(ctx context.Context, identifier w3c.DID) (*domain.IdentityState, error)
	GetKeyIDFromAuthClaim(ctx context.Context, authClaim *domain.Claim) (kms.KeyID, error)
	GetUnprocessedIssuersIDs(ctx context.Context) ([]*w3c.DID, error)
	HasUnprocessedStatesByID(ctx context.Context, identifier w3c.DID) (bool, error)
	HasUnprocessedAndFailedStatesByID(ctx context.Context, identifier w3c.DID) (bool, error)
	GetNonTransactedStates(ctx context.Context) ([]domain.IdentityState, error)
	UpdateIdentityState(ctx context.Context, state *domain.IdentityState) error
	GetTransactedStates(ctx context.Context) ([]domain.IdentityState, error)
	GetStates(ctx context.Context, issuerDID w3c.DID) ([]domain.IdentityState, error)
	CreateAuthenticationQRCode(ctx context.Context, serverURL string, issuerDID w3c.DID) (string, error)
	Authenticate(ctx context.Context, message string, sessionID uuid.UUID, serverURL string, issuerDID w3c.DID) (*protocol.AuthorizationResponseMessage, error)
	GetFailedState(ctx context.Context, identifier w3c.DID) (*domain.IdentityState, error)
	PublishGenesisStateToRHS(ctx context.Context, did *w3c.DID) error
}
