package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// VerificationService interface.
type VerificationService interface {
	CheckVerification(ctx context.Context, issuerID w3c.DID, verificationQueryID uuid.UUID, userDID string) (*domain.VerificationResponse, *domain.VerificationQuery, error)
	SubmitVerificationResponse(ctx context.Context, verificationScopeID uuid.UUID, userDID string, response domain.VerificationResponse) (*domain.VerificationResponse, error)
}
