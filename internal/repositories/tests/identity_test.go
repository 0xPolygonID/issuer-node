package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestGetIdentities(t *testing.T) {
	fixture := tests.NewFixture(storage)
	idStr1 := "did:polygonid:polygon:mumbai:2qGqLpDT2VyqFq1NmfRkB9gwLxBhMRuazv2ZgHfjUw"
	idStr2 := "did:polygonid:polygon:mumbai:2qNR5sUiiSt5v6bnKQZyjCu2n9uNbKD34cZkSkgwUq"

	identity1 := &domain.Identity{
		Identifier: idStr1,
		Relay:      "relay_mock",
		Immutable:  false,
	}

	identity2 := &domain.Identity{
		Identifier: idStr2,
		Relay:      "relay_mock",
		Immutable:  false,
	}
	fixture.CreateIdentity(t, identity1)
	fixture.CreateIdentity(t, identity2)

	identityRepo := repositories.NewIdentity()
	t.Run("should get identities", func(t *testing.T) {
		identities, err := identityRepo.Get(context.Background(), storage.Pgx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(identities))
		assert.Equal(t, idStr1, identities[0])
		assert.Equal(t, idStr2, identities[1])
	})
}
