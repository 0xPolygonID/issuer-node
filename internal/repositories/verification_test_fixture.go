package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// CreateVerificationQuery creates a new query
func (f *Fixture) CreateVerificationQuery(t *testing.T, issuerID w3c.DID, query domain.VerificationQuery) {
	t.Helper()
	_, err := f.verificationRepository.Save(context.Background(), issuerID, query)
	assert.NoError(t, err, "Failed to create verification query")
}

// CreateVerificationResponse creates a new response for a query
func (f *Fixture) CreateVerificationResponse(t *testing.T, queryID uuid.UUID, response domain.VerificationResponse) {
	t.Helper()
	responseID, err := f.verificationRepository.AddResponse(context.Background(), queryID, response)
	assert.NoError(t, err, "Failed to create verification response")
	assert.NotEmpty(t, responseID, "Response ID should not be empty")
}
