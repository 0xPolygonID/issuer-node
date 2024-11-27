package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// VerificationRepository is a repository for verification queries
type VerificationRepository interface {
	Save(ctx context.Context, issuerID w3c.DID, query domain.VerificationQuery) (uuid.UUID, error)
	Get(ctx context.Context, issuerID w3c.DID, id uuid.UUID) (*domain.VerificationQuery, error)
	GetAll(ctx context.Context, issuerID w3c.DID) ([]domain.VerificationQuery, error)
	AddResponse(ctx context.Context, scopeID uuid.UUID, response domain.VerificationResponse) (uuid.UUID, error)
	GetVerificationResponse(ctx context.Context, scopeID uuid.UUID, userDID string) (*domain.VerificationResponse, error)
}
