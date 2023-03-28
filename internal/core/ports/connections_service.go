package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// DeleteRequest struct
type DeleteRequest struct {
	ConnID            uuid.UUID
	DeleteCredentials bool
	RevokeCredentials bool
}

// NewDeleteRequest creates a new DeleteRequest
func NewDeleteRequest(connID uuid.UUID, deleteCredentials *bool, revokeCredentials *bool) *DeleteRequest {
	return &DeleteRequest{
		ConnID:            connID,
		DeleteCredentials: deleteCredentials != nil && *deleteCredentials,
		RevokeCredentials: revokeCredentials != nil && *revokeCredentials,
	}
}

// ConnectionsService  is the interface implemented by the Connections service
type ConnectionsService interface {
	Delete(ctx context.Context, id uuid.UUID, deleteCredentials bool, issuerDID core.DID) error
	DeleteCredentials(ctx context.Context, id uuid.UUID, issuerID core.DID) error
	GetByIDAndIssuerID(ctx context.Context, id uuid.UUID, issuerDID core.DID) (*domain.Connection, error)
	GetAllByIssuerID(ctx context.Context, issuerDID core.DID, query *string) ([]*domain.Connection, error)
}
