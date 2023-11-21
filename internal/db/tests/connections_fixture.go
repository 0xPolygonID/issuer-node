package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// CreateConnection fixture
func (f *Fixture) CreateConnection(t *testing.T, conn *domain.Connection) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	id, err := f.connectionsRepository.Save(ctx, f.storage.Pgx, conn)
	assert.NoError(t, err)
	return id
}

// CreateUserAuthentication fixture
func (f *Fixture) CreateUserAuthentication(t *testing.T, connID uuid.UUID, sessID uuid.UUID, mTime time.Time) {
	t.Helper()
	ctx := context.Background()
	err := f.connectionsRepository.SaveUserAuthentication(ctx, f.storage.Pgx, connID, sessID, mTime)
	assert.NoError(t, err)
}
