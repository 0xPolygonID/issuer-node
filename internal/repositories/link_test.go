package repositories

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

func TestSaveLink(t *testing.T) {
	ctx := context.Background()
	didStr := "did:polygonid:polygon:mumbai:2qPtCq1WDpimtqsFPkpbBYzgzDbJ8i3pn9vHDLyF63"
	schemaStore := NewSchema(*storage)

	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	assert.NoError(t, err)

	schemaID := insertSchemaForLink(ctx, didStr, schemaStore, t)

	linkStore := NewLink(*storage)

	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	validUntil := time.Date(2050, 8, 15, 14, 30, 45, 100, time.Local)
	credentialExpiration := time.Date(2050, 8, 15, 14, 30, 45, 100, time.Local)
	linkToSave := domain.NewLink(*did, common.ToPointer(10), &validUntil, schemaID, &credentialExpiration, true, false, domain.CredentialSubject{"birthday": 19790911, "documentType": 1},
		&verifiable.RefreshService{
			ID:   "https://refresh.xyz",
			Type: verifiable.Iden3RefreshService2023,
		},
		&verifiable.DisplayMethod{
			ID:   "https://display.xyz",
			Type: verifiable.Iden3BasicDisplayMethodV1,
		},
	)

	linkID, err := linkStore.Save(ctx, storage.Pgx, linkToSave)
	assert.NoError(t, err)
	assert.NotNil(t, linkID)
	linkFetched, err := linkStore.GetByID(ctx, *did, *linkID)
	require.NoError(t, err)
	assert.Equal(t, linkToSave.Active, linkFetched.Active)
	assert.Equal(t, linkToSave.MaxIssuance, linkFetched.MaxIssuance)
	assert.InDelta(t, linkToSave.ValidUntil.Unix(), linkFetched.ValidUntil.Unix(), 500)
	assert.Equal(t, linkToSave.SchemaID, linkFetched.SchemaID)
	assert.Equal(t, linkToSave.CredentialSignatureProof, linkFetched.CredentialSignatureProof)
	assert.Equal(t, linkToSave.CredentialMTPProof, linkFetched.CredentialMTPProof)
	assert.Equal(t, linkToSave.RefreshService, linkFetched.RefreshService)
	assert.Equal(t, linkToSave.DisplayMethod, linkFetched.DisplayMethod)
	tcCred, err := json.Marshal(linkToSave.CredentialSubject)
	require.NoError(t, err)
	respCred, err := json.Marshal(linkFetched.CredentialSubject)
	require.NoError(t, err)
	assert.Equal(t, tcCred, respCred)

	didStr2 := "did:polygonid:polygon:mumbai:2qFrLQA6R1bfUTxjRnZEN9st77g6ZN2c7Vw1Dq6Vpp"
	_, err = storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr2, "BJJ")
	assert.NoError(t, err)
	did2, err := w3c.ParseDID(didStr2)
	require.NoError(t, err)
	schemaID2 := insertSchemaForLink(ctx, didStr2, schemaStore, t)
	validUntil = time.Date(2055, 8, 15, 14, 30, 45, 100, time.Local)
	linkToSave.Active = false
	linkToSave.MaxIssuance = common.ToPointer[int](20)
	linkToSave.CredentialExpiration = common.ToPointer(time.Date(2055, 8, 15, 14, 30, 45, 100, time.Local))
	linkToSave.CredentialMTPProof = false
	linkToSave.CredentialSignatureProof = false
	linkToSave.ValidUntil = &validUntil
	linkToSave.SchemaID = schemaID2
	linkToSave.IssuerDID = domain.LinkCoreDID(*did2)
	linkToSave.CredentialSubject = domain.CredentialSubject{"birthday": 19791011, "documentTpe": 2}
	linkID, err = linkStore.Save(ctx, storage.Pgx, linkToSave)
	assert.NoError(t, err)
	linkFetched, err = linkStore.GetByID(ctx, *did2, *linkID)
	require.NoError(t, err)
	assert.Equal(t, linkToSave.SchemaID, linkFetched.SchemaID)
	assert.Equal(t, linkToSave.IssuerDID, linkFetched.IssuerDID)
	assert.Equal(t, linkToSave.Active, linkFetched.Active)
	assert.Equal(t, linkToSave.MaxIssuance, linkFetched.MaxIssuance)
	assert.InDelta(t, linkToSave.CredentialExpiration.Unix(), linkFetched.CredentialExpiration.Unix(), 100)
	assert.InDelta(t, linkToSave.ValidUntil.Unix(), linkFetched.ValidUntil.Unix(), 100)
	assert.Equal(t, linkToSave.CredentialMTPProof, linkFetched.CredentialMTPProof)
	assert.Equal(t, linkToSave.CredentialSignatureProof, linkFetched.CredentialSignatureProof)
	tcCred, err = json.Marshal(linkToSave.CredentialSubject)
	require.NoError(t, err)
	respCred, err = json.Marshal(linkFetched.CredentialSubject)
	require.NoError(t, err)
	assert.Equal(t, tcCred, respCred)
}

func insertSchemaForLink(ctx context.Context, didStr string, store ports.SchemaRepository, t *testing.T) uuid.UUID {
	t.Helper()
	SchemaStore := NewSchema(*storage)
	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)
	insertSchemaGetAllData(t, ctx, *did, SchemaStore)

	data := struct {
		typ        string
		attributes domain.SchemaWords
	}{typ: "age", attributes: domain.SchemaWords{"birthday", "documentType"}}

	s := &domain.Schema{
		ID:        uuid.New(),
		IssuerDID: *did,
		URL:       "url is not important in this test but need to be unique",
		Type:      data.typ,
		Words:     data.attributes,
		CreatedAt: time.Now(),
	}
	require.NoError(t, store.Save(ctx, s))
	return s.ID
}

func TestGetLinkById(t *testing.T) {
	ctx := context.Background()
	didStr := "did:polygonid:polygon:mumbai:2qP8C6HFRANi79HDdnak4b2QJeGewKWbQBYakNXJTh"
	schemaStore := NewSchema(*storage)
	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	require.NoError(t, err)

	schemaID := insertSchemaForLink(ctx, didStr, schemaStore, t)

	linkStore := NewLink(*storage)

	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	validUntil := time.Date(2050, 8, 15, 14, 30, 45, 100, time.Local)
	credentialExpiration := time.Date(2050, 8, 15, 14, 30, 45, 100, time.Local)
	linkToSave := domain.NewLink(*did, common.ToPointer[int](10), &validUntil, schemaID, &credentialExpiration, true, false, domain.CredentialSubject{}, nil, nil)
	linkID, err := linkStore.Save(ctx, storage.Pgx, linkToSave)
	assert.NoError(t, err)
	assert.NotNil(t, linkID)

	linkFetched, err := linkStore.GetByID(ctx, *did, *linkID)
	assert.NoError(t, err)
	assert.NotNil(t, linkFetched)

	_, err = linkStore.GetByID(ctx, *did, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, ErrLinkDoesNotExist, err)
}

func TestGetAll(t *testing.T) {
	ctx := context.Background()
	fixture := NewFixture(storage)
	didStr := "did:iden3:tLZ7NJdCek9j79a1Pmxci3seELHctfGibcrnjjftQ"
	schemaStore := NewSchema(*storage)
	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	require.NoError(t, err)
	linkStore := NewLink(*storage)

	schemaID := insertSchemaForLink(ctx, didStr, schemaStore, t)
	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	tomorrow := time.Now().Add(24 * time.Hour)
	nextWeek := time.Now().Add(7 * 24 * time.Hour)
	past := time.Now().Add(-100 * 24 * time.Hour)
	// 10  not expired links and no max issuance
	for i := 0; i < 10; i++ {
		linkToSave := domain.NewLink(*did, nil, &tomorrow, schemaID, &nextWeek, true, false, domain.CredentialSubject{}, nil, nil)
		linkID, err := linkStore.Save(ctx, storage.Pgx, linkToSave)
		require.NoError(t, err)
		assert.NotNil(t, linkID)
	}
	// 10  not expired links
	for i := 0; i < 10; i++ {
		linkToSave := domain.NewLink(*did, common.ToPointer[int](10), &tomorrow, schemaID, &nextWeek, true, false, domain.CredentialSubject{}, nil, nil)
		linkID, err := linkStore.Save(ctx, storage.Pgx, linkToSave)
		require.NoError(t, err)
		assert.NotNil(t, linkID)
	}
	// 10 expired ones
	for i := 0; i < 10; i++ {
		linkToSave := domain.NewLink(*did, common.ToPointer[int](10), &past, schemaID, &nextWeek, true, false, domain.CredentialSubject{}, nil, nil)
		linkID, err := linkStore.Save(ctx, storage.Pgx, linkToSave)
		require.NoError(t, err)
		assert.NotNil(t, linkID)
	}
	// 10 valid but over used
	for i := 0; i < 10; i++ {
		linkToSave := domain.NewLink(*did, common.ToPointer[int](10), &tomorrow, schemaID, &nextWeek, true, false, domain.CredentialSubject{}, nil, nil)
		linkToSave.MaxIssuance = common.ToPointer(100)

		linkID, err := linkStore.Save(ctx, storage.Pgx, linkToSave)
		require.NoError(t, err)
		assert.NotNil(t, linkID)

		for j := 0; j <= 200; j++ {
			idClaim, _ := uuid.NewUUID()
			HIndex := uuid.New().String()
			fixture.CreateClaim(t, &domain.Claim{
				ID:              idClaim,
				Identifier:      &didStr,
				Issuer:          didStr,
				SchemaHash:      "ca938857241db9451ea329256b9c06e5",
				SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
				SchemaType:      "AuthBJJCredential",
				OtherIdentifier: "did:polygonid:polygon:mumbai:2qP8KN3KRwBi37jB2ENXrWxhTo3pefaU5u5BFPbjYo",
				Expiration:      0,
				Version:         0,
				RevNonce:        0,
				CoreClaim:       domain.CoreClaim{},
				Status:          nil,
				HIndex:          HIndex,
				LinkID:          linkID,
			})
		}
	}
	// 10 inactive
	for i := 0; i < 10; i++ {
		linkToSave := domain.NewLink(*did, common.ToPointer[int](10), &tomorrow, schemaID, &nextWeek, true, false, domain.CredentialSubject{}, nil, nil)
		linkToSave.Active = false
		linkID, err := linkStore.Save(ctx, storage.Pgx, linkToSave)
		require.NoError(t, err)
		assert.NotNil(t, linkID)
	}
	type expected struct {
		count  int
		active *string
	}
	type testConfig struct {
		name     string
		filter   ports.LinkStatus
		query    *string
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name:   "all",
			filter: ports.LinkAll,
			query:  nil,
			expected: expected{
				count: 50,
			},
		},
		{
			name:   "excedeed",
			filter: ports.LinkExceeded,
			query:  nil,
			expected: expected{
				count:  20, // 10 expired + 10 over used
				active: common.ToPointer(string(ports.LinkExceeded)),
			},
		},
		{
			name:   "inactive",
			filter: ports.LinkInactive,
			query:  nil,
			expected: expected{
				count:  10,
				active: common.ToPointer(string(ports.LinkInactive)),
			},
		},
		{
			name:   "active",
			filter: ports.LinkActive,
			query:  nil,
			expected: expected{
				count:  20,
				active: common.ToPointer(string(ports.LinkActive)),
			},
		},
		{
			name:   "active, with query that should not match",
			filter: ports.LinkActive,
			query:  common.ToPointer("NOOOOT MATCH"),
			expected: expected{
				count: 0,
			},
		},
		{
			name:   "active, with query that should match",
			filter: ports.LinkActive,
			query:  common.ToPointer("birthday"),
			expected: expected{
				count:  20,
				active: common.ToPointer(string(ports.LinkActive)),
			},
		},
		{
			name:   "active, with query that should match because of the beginning of a term",
			filter: ports.LinkActive,
			query:  common.ToPointer("birth"),
			expected: expected{
				count:  20,
				active: common.ToPointer(string(ports.LinkActive)),
			},
		},
		{
			name:   "active, with query that should match because in the middle of a term",
			filter: ports.LinkActive,
			query:  common.ToPointer("thday"),
			expected: expected{
				count:  20,
				active: common.ToPointer(string(ports.LinkActive)),
			},
		},
		{
			name:   "inactive, with query that should match",
			filter: ports.LinkInactive,
			query:  common.ToPointer("birthday"),
			expected: expected{
				count:  10,
				active: common.ToPointer(string(ports.LinkInactive)),
			},
		},
		{
			name:     "inactive, with query that should NOT match",
			filter:   ports.LinkInactive,
			query:    common.ToPointer("NORRR"),
			expected: expected{count: 0},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			all, err := linkStore.GetAll(ctx, *did, tc.filter, tc.query)
			require.NoError(t, err)
			require.Len(t, all, tc.expected.count)
			for _, one := range all {
				if tc.expected.active != nil {
					assert.Equal(t, one.Status(), *tc.expected.active)
				}
			}
		})
	}
}

func TestDeleteLink(t *testing.T) {
	ctx := context.Background()
	didStr := "did:polygonid:polygon:mumbai:2qJ8RWkEpMtsAwnACo5oUktJSeS1wqPfnXMF99Y4Hj"
	didStr2 := "did:polygonid:polygon:mumbai:2qPKWbeUSqzk6zGx4cv1EspaDMJXu2suprCr6yHfkQ"
	schemaStore := NewSchema(*storage)

	fixture := NewFixture(storage)
	identity := &domain.Identity{
		Identifier: didStr,
	}
	fixture.CreateIdentity(t, identity)

	schemaID := insertSchemaForLink(ctx, didStr, schemaStore, t)
	linkStore := NewLink(*storage)

	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	did2, err := w3c.ParseDID(didStr2)
	require.NoError(t, err)

	validUntil := time.Date(2050, 8, 15, 14, 30, 45, 100, time.Local)
	credentialExpiration := time.Date(2050, 8, 15, 14, 30, 45, 100, time.Local)
	linkToSave := domain.NewLink(*did, common.ToPointer[int](10), &validUntil, schemaID, &credentialExpiration, true, false, domain.CredentialSubject{}, nil, nil)

	linkID, err := linkStore.Save(ctx, storage.Pgx, linkToSave)
	assert.NoError(t, err)
	assert.NotNil(t, linkID)

	idClaim, _ := uuid.NewUUID()

	fixture.CreateClaim(t, &domain.Claim{
		ID:              idClaim,
		Identifier:      &didStr,
		Issuer:          didStr,
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: "",
		Expiration:      0,
		Version:         0,
		RevNonce:        0,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
		HIndex:          "123",
		LinkID:          linkID,
	})

	err = linkStore.Delete(ctx, *linkID, *did2)
	assert.Error(t, err)

	err = linkStore.Delete(ctx, *linkID, *did)
	assert.NoError(t, err)

	claimStorage := NewClaims()
	claim, err := claimStorage.GetByIdAndIssuer(ctx, storage.Pgx, did, idClaim)
	assert.NoError(t, err)
	assert.NotNil(t, claim)

	err = linkStore.Delete(ctx, uuid.New(), *did)
	assert.Error(t, err)
	assert.Equal(t, ErrLinkDoesNotExist, err)
}
