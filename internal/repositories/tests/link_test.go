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
	credentialExpiration := time.Date(2050, 8, 15, 14, 30, 45, 100, time.Local)
	linkToSave := domain.NewLink(did, common.ToPointer[int](10), &validUntil, schemaID, &credentialExpiration, true, false,
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

	didStr2 := "did:polygonid:polygon:mumbai:2qFrLQA6R1bfUTxjRnZEN9st77g6ZN2c7Vw1Dq6Vpp"
	_, err = storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier) VALUES ($1)", didStr2)
	assert.NoError(t, err)
	did2 := core.DID{}
	require.NoError(t, did2.SetString(didStr2))
	schemaID2 := insertSchemaForLink(ctx, didStr2, schemaStore, t)
	validUntil = time.Date(2055, 8, 15, 14, 30, 45, 100, time.Local)
	linkToSave.Active = false
	linkToSave.MaxIssuance = common.ToPointer[int](20)
	linkToSave.CredentialExpiration = common.ToPointer(time.Date(2055, 8, 15, 14, 30, 45, 100, time.Local))
	linkToSave.CredentialMTPProof = false
	linkToSave.CredentialSignatureProof = false
	linkToSave.ValidUntil = &validUntil
	linkToSave.SchemaID = schemaID2
	linkToSave.IssuerDID = domain.LinkCoreDID(did2)
	linkToSave.CredentialAttributes = []domain.CredentialAttributes{{Name: "birthday", Value: "19791011"}, {Name: "documentTpe", Value: "2"}}

	linkID, err = linkStore.Save(ctx, linkToSave)
	assert.NoError(t, err)
	linkFetched, err = linkStore.GetByID(ctx, *linkID)
	assert.NoError(t, err)
	assert.Equal(t, linkToSave.SchemaID, linkFetched.SchemaID)
	assert.Equal(t, linkToSave.IssuerDID, linkFetched.IssuerDID)
	assert.Equal(t, linkToSave.Active, linkFetched.Active)
	assert.Equal(t, linkToSave.MaxIssuance, linkFetched.MaxIssuance)
	assert.InDelta(t, linkToSave.CredentialExpiration.Unix(), linkFetched.CredentialExpiration.Unix(), 100)
	assert.InDelta(t, linkToSave.ValidUntil.Unix(), linkFetched.ValidUntil.Unix(), 100)
	assert.Equal(t, linkToSave.CredentialMTPProof, linkFetched.CredentialMTPProof)
	assert.Equal(t, linkToSave.CredentialSignatureProof, linkFetched.CredentialSignatureProof)
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

func TestGetLinkById(t *testing.T) {
	ctx := context.Background()
	didStr := "did:polygonid:polygon:mumbai:2qP8C6HFRANi79HDdnak4b2QJeGewKWbQBYakNXJTh"
	schemaStore := repositories.NewSchema(*storage)

	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier) VALUES ($1)", didStr)
	assert.NoError(t, err)

	schemaID := insertSchemaForLink(ctx, didStr, schemaStore, t)

	linkStore := repositories.NewLink(*storage)

	did := core.DID{}
	require.NoError(t, did.SetString(didStr))

	validUntil := time.Date(2050, 8, 15, 14, 30, 45, 100, time.Local)
	credentialExpiration := time.Date(2050, 8, 15, 14, 30, 45, 100, time.Local)
	linkToSave := domain.NewLink(did, common.ToPointer[int](10), &validUntil, schemaID, &credentialExpiration, true, false,
		[]domain.CredentialAttributes{{Name: "birthday", Value: "19790911"}, {Name: "documentTpe", Value: "1"}})

	linkID, err := linkStore.Save(ctx, linkToSave)
	assert.NoError(t, err)
	assert.NotNil(t, linkID)
	linkFetched, err := linkStore.GetByID(ctx, *linkID)
	assert.NoError(t, err)
	assert.NotNil(t, linkFetched)

	_, err = linkStore.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrLinkDoesNotExist, err)
}

func TestDeleteLink(t *testing.T) {
	ctx := context.Background()
	didStr := "did:polygonid:polygon:mumbai:2qJ8RWkEpMtsAwnACo5oUktJSeS1wqPfnXMF99Y4Hj"
	didStr2 := "did:polygonid:polygon:mumbai:2qPKWbeUSqzk6zGx4cv1EspaDMJXu2suprCr6yHfkQ"
	schemaStore := repositories.NewSchema(*storage)

	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier) VALUES ($1)", didStr)
	assert.NoError(t, err)

	_, err = storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier) VALUES ($1)", didStr2)
	assert.NoError(t, err)

	schemaID := insertSchemaForLink(ctx, didStr, schemaStore, t)

	linkStore := repositories.NewLink(*storage)

	did := core.DID{}
	require.NoError(t, did.SetString(didStr))

	did2 := core.DID{}
	require.NoError(t, did2.SetString(didStr2))

	validUntil := time.Date(2050, 8, 15, 14, 30, 45, 100, time.Local)
	credentialExpiration := time.Date(2050, 8, 15, 14, 30, 45, 100, time.Local)
	linkToSave := domain.NewLink(did, common.ToPointer[int](10), &validUntil, schemaID, &credentialExpiration, true, false,
		[]domain.CredentialAttributes{{Name: "birthday", Value: "19790911"}, {Name: "documentTpe", Value: "1"}})

	linkID, err := linkStore.Save(ctx, linkToSave)
	assert.NoError(t, err)
	assert.NotNil(t, linkID)

	err = linkStore.Delete(ctx, *linkID, did2)
	assert.Error(t, err)

	err = linkStore.Delete(ctx, *linkID, did)
	assert.NoError(t, err)

	err = linkStore.Delete(ctx, uuid.New(), did)
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrLinkDoesNotExist, err)
}
