package ports

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/verifiable"
	comm "github.com/iden3/iden3comm"
	"github.com/iden3/iden3comm/protocol"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// CreateClaimRequest struct
type CreateClaimRequest struct {
	DID                   *core.DID
	Schema                string
	CredentialSubject     map[string]any
	Expiration            *time.Time
	Type                  string
	Version               uint32
	SubjectPos            string
	MerklizedRootPosition string
	SignatureProof        bool
	MTProof               bool
}

// AgentRequest struct
type AgentRequest struct {
	Body      json.RawMessage
	ThreadID  string
	IssuerDID *core.DID
	UserDID   *core.DID
	ClaimID   uuid.UUID
	Typ       comm.MediaType
	Type      comm.ProtocolMessage
}

// Filter struct
type Filter struct {
	Self       *bool
	Revoked    *bool
	SchemaHash string
	SchemaType string
	Subject    string
	QueryField string
}

// NewClaimsFilter returns a valid claims filter
func NewClaimsFilter(schemaHash, schemaType, subject, queryField *string, self, revoked *bool) (*Filter, error) {
	var filter Filter

	if self != nil && *self {
		if subject != nil && *subject != "" {
			return nil, fmt.Errorf("self and subject filter can not be used together")
		}
		filter.Self = self
	}

	if schemaHash != nil && *schemaHash != "" {
		filter.SchemaHash = *schemaHash
	}

	if schemaType != nil && *schemaType != "" {
		filter.SchemaType = *schemaType
	}

	if revoked != nil {
		filter.Revoked = revoked
	}

	if subject != nil && *subject != "" {
		filter.Subject = *subject
	}

	if queryField != nil && *queryField != "" {
		filter.QueryField = *queryField
	}

	return &filter, nil
}

// NewCreateClaimRequest returns a new claim object with the given parameters
func NewCreateClaimRequest(did *core.DID, credentialSchema string, credentialSubject map[string]any, expiration *int64, typ string, cVersion *uint32, subjectPos *string, merklizedRootPosition *string, sigProof *bool, mtProof *bool) *CreateClaimRequest {
	if sigProof == nil {
		sigProof = common.ToPointer(false)
	}

	if mtProof == nil {
		mtProof = common.ToPointer(false)
	}

	req := &CreateClaimRequest{
		DID:               did,
		Schema:            credentialSchema,
		CredentialSubject: credentialSubject,
		Type:              typ,
		SignatureProof:    *sigProof,
		MTProof:           *mtProof,
	}
	if expiration != nil {
		t := time.Unix(*expiration, 0)
		req.Expiration = &t
	}
	if cVersion != nil {
		req.Version = *cVersion
	}
	if subjectPos != nil {
		req.SubjectPos = *subjectPos
	}
	if merklizedRootPosition != nil {
		req.MerklizedRootPosition = *merklizedRootPosition
	}
	return req
}

// NewAgentRequest validates the inputs and returns a new AgentRequest
func NewAgentRequest(basicMessage *comm.BasicMessage) (*AgentRequest, error) {
	if basicMessage.To == "" {
		return nil, fmt.Errorf("'to' field can not be empty")
	}

	toDID, err := core.ParseDID(basicMessage.To)
	if err != nil {
		return nil, err
	}

	if basicMessage.From == "" {
		return nil, fmt.Errorf("'from' field can not be empty")
	}

	fromDID, err := core.ParseDID(basicMessage.From)
	if err != nil {
		return nil, err
	}

	if basicMessage.ID == "" {
		return nil, fmt.Errorf("'id' field can not be empty")
	}

	claimID, err := uuid.Parse(basicMessage.ID)
	if err != nil {
		return nil, err
	}

	if basicMessage.Type != protocol.CredentialFetchRequestMessageType && basicMessage.Type != protocol.RevocationStatusRequestMessageType {
		return nil, fmt.Errorf("invalid type")
	}

	if basicMessage.ID == "" {
		return nil, fmt.Errorf("'id' field can not be empty")
	}

	return &AgentRequest{
		Body:      basicMessage.Body,
		UserDID:   fromDID,
		IssuerDID: toDID,
		ThreadID:  basicMessage.ThreadID,
		ClaimID:   claimID,
		Typ:       basicMessage.Typ,
		Type:      basicMessage.Type,
	}, nil
}

// ClaimsService is the interface implemented by the claim service
type ClaimsService interface {
	CreateClaim(ctx context.Context, claimReq *CreateClaimRequest) (*domain.Claim, error)
	Revoke(ctx context.Context, id core.DID, nonce uint64, description string) error
	RevokeAllFromConnection(ctx context.Context, connID uuid.UUID, issuerID core.DID) error
	GetAll(ctx context.Context, did *core.DID, filter *Filter) ([]*domain.Claim, error)
	GetRevocationStatus(ctx context.Context, id string, nonce uint64) (*verifiable.RevocationStatus, error)
	GetByID(ctx context.Context, issID *core.DID, id uuid.UUID) (*domain.Claim, error)
	Agent(ctx context.Context, req *AgentRequest) (*domain.Agent, error)
	GetAuthClaim(ctx context.Context, did *core.DID) (*domain.Claim, error)
	GetAuthClaimForPublishing(ctx context.Context, did *core.DID, state string) (*domain.Claim, error)
	UpdateClaimsMTPAndState(ctx context.Context, currentState *domain.IdentityState) error
	Delete(ctx context.Context, id uuid.UUID) error
}
