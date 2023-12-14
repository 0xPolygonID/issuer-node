package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/pkg/pagination"
)

// NewGetAllConnectionsRequest struct
type NewGetAllConnectionsRequest struct {
	WithCredentials bool
	Query           string
	Pagination      *pagination.Filter
}

// DeleteRequest struct
type DeleteRequest struct {
	ConnID            uuid.UUID
	DeleteCredentials bool
	RevokeCredentials bool
}

// NewGetAllRequest returns the request object for obtaining all connections
func NewGetAllRequest(withCredentials *bool, query *string, page *uint, maxResults *uint) *NewGetAllConnectionsRequest {
	var (
		connQuery string
		pagFilter *pagination.Filter
	)
	if query != nil {
		connQuery = *query
	}

	if page != nil {
		pagFilter = pagination.NewFilter(maxResults, page)
	}

	return &NewGetAllConnectionsRequest{
		WithCredentials: withCredentials != nil && *withCredentials,
		Query:           connQuery,
		Pagination:      pagFilter,
	}
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
	Delete(ctx context.Context, id uuid.UUID, deleteCredentials bool, issuerDID w3c.DID) error
	DeleteCredentials(ctx context.Context, id uuid.UUID, issuerID w3c.DID) error
	GetByIDAndIssuerID(ctx context.Context, id uuid.UUID, issuerDID w3c.DID) (*domain.Connection, error)
	GetByUserID(ctx context.Context, issuerDID w3c.DID, userID w3c.DID) (*domain.Connection, error)
	GetAllByIssuerID(ctx context.Context, issuerDID w3c.DID, request *NewGetAllConnectionsRequest) ([]*domain.Connection, uint, error)
	GetByUserSessionID(ctx context.Context, sessionID uuid.UUID) (*domain.Connection, error)
}
