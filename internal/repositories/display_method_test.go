package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/sqltools"
)

func TestSaveDisplayMethod(t *testing.T) {
	ctx := context.Background()
	displayMethodRepository := NewDisplayMethod(*storage)
	didStr := "did:iden3:privado:main:2SbVmQkU7H2WdKgQ1mMEXgUY3cHxcsVDe216K4XazX"
	issuerDID, err := w3c.ParseDID(didStr)
	require.NoError(t, err)
	_, err = storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	require.NoError(t, err)
	displayMethod := domain.NewDisplayMethod(uuid.New(), *issuerDID, "test", "http://test.com", common.ToPointer("Iden3BasicDisplayMethodV1"))

	t.Run("Save display method", func(t *testing.T) {
		id, err := displayMethodRepository.Save(ctx, displayMethod)
		require.NoError(t, err)
		require.NotNil(t, id)
	})

	t.Run("Save display method with same id", func(t *testing.T) {
		id, err := displayMethodRepository.Save(ctx, displayMethod)
		require.NoError(t, err)
		require.NotNil(t, id)
	})

	t.Run("Save display method with same name", func(t *testing.T) {
		displayMethod.ID = uuid.New()
		id, err := displayMethodRepository.Save(ctx, displayMethod)
		require.Error(t, err)
		require.Nil(t, id)
	})
}

func TestGetDisplayMethod(t *testing.T) {
	ctx := context.Background()
	displayMethodRepository := NewDisplayMethod(*storage)
	didStr := "did:iden3:privado:main:2ShTaDYzzn1qvQVWLetH35FnCWVm3yZGgEd8q6AkM3"
	issuerDID, err := w3c.ParseDID(didStr)
	require.NoError(t, err)
	_, err = storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	require.NoError(t, err)
	displayMethod := domain.NewDisplayMethod(uuid.New(), *issuerDID, "test", "http://test.com", common.ToPointer("Iden3BasicDisplayMethodV1"))

	t.Run("should return a display method", func(t *testing.T) {
		id, err := displayMethodRepository.Save(ctx, displayMethod)
		require.NoError(t, err)
		displayMethodToGet, err := displayMethodRepository.GetByID(ctx, *issuerDID, *id)
		require.NoError(t, err)
		require.NotNil(t, displayMethodToGet)
		assert.Equal(t, displayMethod.ID, displayMethodToGet.ID)
		assert.Equal(t, displayMethod.Name, displayMethodToGet.Name)
		assert.Equal(t, displayMethod.URL, displayMethodToGet.URL)
		assert.Equal(t, displayMethod.IssuerDID, displayMethodToGet.IssuerDID)
		assert.Equal(t, displayMethod.Type, displayMethodToGet.Type)
	})

	t.Run("should return an error", func(t *testing.T) {
		displayMethodToGet, err := displayMethodRepository.GetByID(ctx, *issuerDID, uuid.New())
		require.Error(t, err)
		require.Nil(t, displayMethodToGet)
	})
}

func TestGetAllDisplayMethod(t *testing.T) {
	ctx := context.Background()
	displayMethodRepository := NewDisplayMethod(*storage)
	didStr := "did:iden3:privado:main:2SbYMQFCzjAguQ7uhPXvutDWp9FdNAtvcucN4WrKLZ"
	issuerDID, err := w3c.ParseDID(didStr)
	require.NoError(t, err)
	_, err = storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	require.NoError(t, err)

	didStr2 := "did:iden3:polygon:amoy:x7zcJTBSWaUSgjmf1vnN1o13zbK5HChCD92JcgmSC"
	issuerDID2, err := w3c.ParseDID(didStr2)
	require.NoError(t, err)
	_, err = storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr2, "BJJ")
	require.NoError(t, err)

	displayMethod := domain.NewDisplayMethod(uuid.New(), *issuerDID, "displayMethod1", "http://test1.com", common.ToPointer("Iden3BasicDisplayMethodV1"))
	displayMethod2 := domain.NewDisplayMethod(uuid.New(), *issuerDID, "displayMethod2", "http://test2.com", common.ToPointer("Iden3BasicDisplayMethodV1"))
	displayMethod3 := domain.NewDisplayMethod(uuid.New(), *issuerDID2, "test", "http://test.com", common.ToPointer("Iden3BasicDisplayMethodV1"))

	_, err = displayMethodRepository.Save(ctx, displayMethod)
	require.NoError(t, err)
	_, err = displayMethodRepository.Save(ctx, displayMethod2)
	require.NoError(t, err)
	_, err = displayMethodRepository.Save(ctx, displayMethod3)
	require.NoError(t, err)

	t.Run("should return two display method - created at desc", func(t *testing.T) {
		displayMethods, total, err := displayMethodRepository.GetAll(ctx, *issuerDID, ports.DisplayMethodFilter{
			MaxResults: 10,
			Page:       1,
			OrderBy: sqltools.OrderByFilters{
				{Field: "created_at", Desc: true},
			},
		})
		require.NoError(t, err)
		assert.Len(t, displayMethods, 2)
		assert.Equal(t, uint(2), total)

		assert.Equal(t, displayMethod2.ID, displayMethods[0].ID)
		assert.Equal(t, displayMethod.ID, displayMethods[1].ID)
		assert.Equal(t, displayMethod2.Name, displayMethods[0].Name)
		assert.Equal(t, displayMethod.Name, displayMethods[1].Name)
	})

	t.Run("should return two display method - created at asc", func(t *testing.T) {
		displayMethods, total, err := displayMethodRepository.GetAll(ctx, *issuerDID, ports.DisplayMethodFilter{
			MaxResults: 10,
			Page:       1,
			OrderBy: sqltools.OrderByFilters{
				{Field: "created_at", Desc: false},
			},
		})
		require.NoError(t, err)
		assert.Len(t, displayMethods, 2)
		assert.Equal(t, uint(2), total)

		assert.Equal(t, displayMethod.ID, displayMethods[0].ID)
		assert.Equal(t, displayMethod2.ID, displayMethods[1].ID)
	})

	t.Run("should return two display method - paginated", func(t *testing.T) {
		displayMethods, total, err := displayMethodRepository.GetAll(ctx, *issuerDID, ports.DisplayMethodFilter{
			MaxResults: 1,
			Page:       1,
			OrderBy: sqltools.OrderByFilters{
				{Field: "created_at", Desc: true},
			},
		})
		require.NoError(t, err)
		assert.Len(t, displayMethods, 1)
		assert.Equal(t, uint(2), total)

		assert.Equal(t, displayMethod2.ID, displayMethods[0].ID)
	})
}

func TestUpdateDisplayMethod(t *testing.T) {
	ctx := context.Background()
	displayMethodRepository := NewDisplayMethod(*storage)
	didStr := "did:iden3:privado:main:2SdurMNuWKbRYKnYntNYHJFiEFtyFWXm2kr3iaKTq6"
	issuerDID, err := w3c.ParseDID(didStr)
	require.NoError(t, err)
	_, err = storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	require.NoError(t, err)
	displayMethod := domain.NewDisplayMethod(uuid.New(), *issuerDID, "test", "http://test.com", common.ToPointer("Iden3BasicDisplayMethodV1"))

	t.Run("should update a display method", func(t *testing.T) {
		id, err := displayMethodRepository.Save(ctx, displayMethod)
		require.NoError(t, err)
		displayMethodToGet, err := displayMethodRepository.GetByID(ctx, *issuerDID, *id)
		require.NoError(t, err)
		require.NotNil(t, displayMethodToGet)
		assert.Equal(t, displayMethod.ID, displayMethodToGet.ID)
		assert.Equal(t, displayMethod.Name, displayMethodToGet.Name)
		assert.Equal(t, displayMethod.URL, displayMethodToGet.URL)
		assert.Equal(t, displayMethod.IssuerDID, displayMethodToGet.IssuerDID)

		displayMethod.URL = "http://test2.com"
		displayMethod.Name = "test2"
		displayMethod.Type = "Iden3BasicDisplayMethodV2"
		id, err = displayMethodRepository.Save(ctx, displayMethod)
		require.NoError(t, err)

		displayMethodAfterUpdate, err := displayMethodRepository.GetByID(ctx, *issuerDID, *id)
		require.NoError(t, err)
		require.NotNil(t, displayMethodAfterUpdate)
		assert.Equal(t, displayMethod.ID, displayMethodAfterUpdate.ID)
		assert.Equal(t, displayMethod.Name, displayMethodAfterUpdate.Name)
		assert.Equal(t, displayMethod.URL, displayMethodAfterUpdate.URL)
		assert.Equal(t, displayMethod.IssuerDID, displayMethodAfterUpdate.IssuerDID)
		assert.Equal(t, displayMethod.Type, displayMethodAfterUpdate.Type)
	})
}

func TestDeleteDisplayMethod(t *testing.T) {
	ctx := context.Background()
	displayMethodRepository := NewDisplayMethod(*storage)
	didStr := "did:iden3:privado:main:2SgSDjqTnxgv6JGhyxVKchxqa2Mbj62bSZAXQG7hKS"
	issuerDID, err := w3c.ParseDID(didStr)
	require.NoError(t, err)
	_, err = storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	require.NoError(t, err)
	displayMethod := domain.NewDisplayMethod(uuid.New(), *issuerDID, "test", "http://test.com", common.ToPointer("Iden3BasicDisplayMethodV1"))

	t.Run("should return a display method", func(t *testing.T) {
		id, err := displayMethodRepository.Save(ctx, displayMethod)
		require.NoError(t, err)
		_, err = displayMethodRepository.GetByID(ctx, *issuerDID, *id)
		require.NoError(t, err)

		err = displayMethodRepository.Delete(ctx, *issuerDID, *id)
		require.NoError(t, err)
	})
}
