package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestSaveLink(t *testing.T) {
	ctx := context.Background()
	didStr := "did:polygonid:polygon:mumbai:2qPtCq1WDpimtqsFPkpbBYzgzDbJ8i3pn9vHDLyF63"
	schemaStore := repositories.NewSchema(*storage)

	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier) VALUES ($1)", didStr)
	assert.NoError(t, err)

	schemaID := insertSchemaForLink(ctx, didStr, schemaStore, t)

	linkStore := repositories.NewLink(*storage)

	did := core.DID{}
	require.NoError(t, did.SetString(didStr))

	validUntil := time.Date(2050, 8, 15, 14, 30, 45, 100, time.Local)
	linkToSave := domain.NewLink(did, common.ToPointer[int32](10), &validUntil, schemaID, true, false,
		[]domain.CredentialAttributes{{Name: "birthday", Value: "19790911"}, {Name: "documentTpe", Value: "1"}})

	linkID, err := linkStore.Save(ctx, linkToSave)
	assert.NoError(t, err)
	assert.NotNil(t, linkID)
	linkFetched, err := linkStore.GetByID(ctx, *linkID)
	assert.NoError(t, err)
	assert.Equal(t, linkToSave.Active, linkFetched.Active)
	assert.Equal(t, linkToSave.MaxIssuance, linkFetched.MaxIssuance)
	assert.InDelta(t, linkToSave.ValidUntil.Unix(), linkFetched.ValidUntil.Unix(), 500)
	assert.Equal(t, linkToSave.SchemaID, linkFetched.SchemaID)
	assert.Equal(t, linkToSave.CredentialSignatureProof, linkFetched.CredentialSignatureProof)
	assert.Equal(t, linkToSave.CredentialMTPProof, linkFetched.CredentialMTPProof)
	assert.Equal(t, linkToSave.CredentialAttributes[0], linkFetched.CredentialAttributes[0])
	assert.Equal(t, linkToSave.CredentialAttributes[1], linkFetched.CredentialAttributes[1])
}

func insertSchemaForLink(ctx context.Context, didStr string, store ports.SchemaRepository, t *testing.T) uuid.UUID {
	t.Helper()
	SchemaStore := repositories.NewSchema(*storage)
	did := core.DID{}
	require.NoError(t, did.SetString(didStr))
	insertSchemaGetAllData(t, ctx, did, SchemaStore)

	data := struct {
		typ        string
		attributes domain.SchemaAttrs
	}{typ: "age", attributes: domain.SchemaAttrs{"birthday", "documentTpe"}}

	s := &domain.Schema{
		ID:         uuid.New(),
		IssuerDID:  did,
		URL:        "url is not important in this test but need to be unique",
		Type:       data.typ,
		Attributes: data.attributes,
		CreatedAt:  time.Now(),
	}
	require.NoError(t, store.Save(ctx, s))
	return s.ID
}
