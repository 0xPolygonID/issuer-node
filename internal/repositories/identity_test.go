package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

func TestGetIdentities(t *testing.T) {
	fixture := NewFixture(storage)
	did1 := randomDID(t)
	idStr1 := did1.String()
	did2 := randomDID(t)
	idStr2 := did2.String()

	identity1 := &domain.Identity{
		Identifier: idStr1,
	}

	identity2 := &domain.Identity{
		Identifier: idStr2,
	}
	fixture.CreateIdentity(t, identity1)
	fixture.CreateIdentity(t, identity2)

	identityRepo := NewIdentity()
	t.Run("should get identities", func(t *testing.T) {
		identities, err := identityRepo.Get(context.Background(), storage.Pgx)
		assert.NoError(t, err)
		assert.True(t, len(identities) >= 2)
	})
}
