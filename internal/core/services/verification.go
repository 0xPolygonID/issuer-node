package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	auth "github.com/iden3/go-iden3-auth/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/jackc/pgtype"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

type VerificationService struct {
	verifier *auth.Verifier
	repo     ports.VerificationRepository
}

// NewVerificationService creates a new instance of VerificationService
func NewVerificationService(verifier *auth.Verifier, repo ports.VerificationRepository) *VerificationService {
	return &VerificationService{verifier: verifier, repo: repo}
}

// CheckVerification checks if a verification response already exists for a given verification query ID and userDID.
// If no response exists, it returns the verification query.
func (vs *VerificationService) CheckVerification(ctx context.Context, issuerID w3c.DID, verificationQueryID uuid.UUID, userDID string) (*domain.VerificationResponse, *domain.VerificationQuery, error) {
	query, err := vs.repo.Get(ctx, issuerID, verificationQueryID)
	if err != nil {
		if err == repositories.VerificationQueryNotFoundError {
			return nil, nil, fmt.Errorf("verification query not found: %w", err)
		}
		return nil, nil, fmt.Errorf("failed to get verification query: %w", err)
	}

	for _, scope := range query.Scopes {
		response, err := vs.repo.GetVerificationResponse(ctx, scope.ID, userDID)
		if err == nil && response != nil {
			return response, nil, nil
		}
	}

	return nil, query, nil
}

// SubmitVerificationResponse submits a verification response for a given verification scope ID and userDID
func (vs *VerificationService) SubmitVerificationResponse(ctx context.Context, verificationScopeID uuid.UUID, userDID string, responseData domain.VerificationResponse) (*domain.VerificationResponse, error) {
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data to JSON: %w", err)
	}

	var jsonbResponse pgtype.JSONB
	err = jsonbResponse.Set(responseJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to set JSONB value: %w", err)
	}

	response := domain.VerificationResponse{
		ID:                  uuid.New(),
		VerificationScopeID: verificationScopeID,
		UserDID:             userDID,
		Response:            jsonbResponse,
		Pass:                true,
		CreatedAt:           time.Now(),
	}

	responseID, err := vs.repo.AddResponse(ctx, verificationScopeID, response)
	if err != nil {
		return nil, fmt.Errorf("failed to add verification response: %w", err)
	}

	response.ID = responseID
	return &response, nil
}
