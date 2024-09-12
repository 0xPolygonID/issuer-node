package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
)

func TestServer_CreateLink(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)
	importedSchema, err := server.Services.schema.ImportSchema(ctx, *did, ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription")))
	assert.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response CreateLinkResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		body     CreateLinkRequest
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Happy path",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local)),
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": 12},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "No merkle tree proof or signature proof selected. At least one should be enabled",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local)),
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": 12},
				MtProof:              false,
				SignatureProof:       false,
			},
			expected: expected{
				response: CreateLink400JSONResponse{N400JSONResponse{Message: "at least one proof type should be enabled"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Claim link expiration exceeded",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           common.ToPointer(time.Date(2000, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": 12},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink400JSONResponse{N400JSONResponse{Message: "invalid claimLinkExpiration. Cannot be a date time prior current time."}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Claim link expiration nil",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: nil,
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": 12},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "Claim expiration date nil",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           nil,
				CredentialExpiration: nil,
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": 12},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "Claim link wrong number of attributes",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: common.ToPointer(time.Date(2020, 8, 15, 14, 30, 45, 100, time.Local)),
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink400JSONResponse{N400JSONResponse{Message: "you must provide at least one attribute"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Claim link wrong attribute type",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: common.ToPointer(time.Date(2000, 8, 15, 14, 30, 45, 100, time.Local)),
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": true},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink400JSONResponse{N400JSONResponse{Message: "cannot parse claim"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Claim link wrong schema id",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             uuid.New(),
				Expiration:           common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: common.ToPointer(time.Date(2000, 8, 15, 14, 30, 45, 100, time.Local)),
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": 12},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink400JSONResponse{N400JSONResponse{Message: "schema does not exist"}},
				httpCode: http.StatusBadRequest,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/credentials/links", did)

			req, err := http.NewRequest(http.MethodPost, url, tests.JSONBody(t, tc.body))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response UUIDResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				_, err := uuid.Parse(response.Id)
				assert.NoError(t, err)
			case http.StatusBadRequest:
				var response CreateLink400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_ActivateLink(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)
	importedSchema, err := server.Services.schema.ImportSchema(ctx, *did, ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription")))
	assert.NoError(t, err)

	tomorrow := time.Now().Add(24 * time.Hour)
	link, err := server.Services.links.Save(ctx, *did, common.ToPointer(10), &tomorrow, importedSchema.ID, nil, true, true, CredentialSubject{"birthday": 19790911, "documentType": 12}, nil, nil)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response ActivateLinkResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		id       uuid.UUID
		auth     func() (string, string)
		body     ActivateLinkJSONBody
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			id:   link.ID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Claim link does not exist",
			auth: authOk,
			id:   uuid.New(),
			body: ActivateLinkJSONBody{
				Active: true,
			},
			expected: expected{
				response: ActivateLink400JSONResponse{N400JSONResponse{Message: "link does not exist"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Claim link already activated",
			auth: authOk,
			id:   link.ID,
			body: ActivateLinkJSONBody{
				Active: true,
			},
			expected: expected{
				response: ActivateLink400JSONResponse{N400JSONResponse{Message: "link is already active"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Happy path",
			auth: authOk,
			id:   link.ID,
			body: ActivateLinkJSONBody{
				Active: false,
			},
			expected: expected{
				response: ActivateLink200JSONResponse{Message: "Link updated"},
				httpCode: http.StatusOK,
			},
		},
		{
			name: "Claim link already deactivated",
			auth: authOk,
			id:   link.ID,
			body: ActivateLinkJSONBody{
				Active: false,
			},
			expected: expected{
				response: ActivateLink400JSONResponse{N400JSONResponse{Message: "link is already inactive"}},
				httpCode: http.StatusBadRequest,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/credentials/links/%s", did, tc.id)

			req, err := http.NewRequest(http.MethodPatch, url, tests.JSONBody(t, tc.body))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GenericMessage
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				expected, ok := tc.expected.response.(ActivateLink200JSONResponse)
				assert.True(t, ok)
				assert.Equal(t, expected.Message, response.Message)

			case http.StatusBadRequest:
				var response ActivateLink400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

// TestServer_GetLink does an end 2 end test for the get link endpoint.
// TIP: Link status test is better covered in the internal/repositories/tests/link_test.go unit tests
// as it is really verbose to do it here.
func TestServer_GetLink(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)
	importedSchema, err := server.Services.schema.ImportSchema(ctx, *did, ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription")))
	assert.NoError(t, err)

	tomorrow := time.Now().Add(24 * time.Hour)
	yesterday := time.Now().Add(-24 * time.Hour)

	link, err := server.Services.links.Save(ctx, *did, common.ToPointer(10), &tomorrow, importedSchema.ID, common.ToPointer(tomorrow), true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	require.NoError(t, err)
	hash, _ := link.Schema.Hash.MarshalText()

	linkExpired, err := server.Services.links.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, common.ToPointer(tomorrow), true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response GetLinkResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		id       uuid.UUID
		auth     func() (string, string)
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			id:   link.ID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Claim link does not exist",
			auth: authOk,
			id:   uuid.New(),
			expected: expected{
				response: GetLink404JSONResponse{N404JSONResponse{Message: "link not found"}},
				httpCode: http.StatusNotFound,
			},
		},
		{
			name: "Happy path, link active by date",
			auth: authOk,
			id:   link.ID,
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLink200JSONResponse{
					Active:               link.Active,
					CredentialSubject:    CredentialSubject{"birthday": 19791109, "documentType": 12, "type": schemaType, "id": "did:polygonid:polygon:mumbai:2qDDDKmo436EZGCBAvkqZjADYoNRJszkG7UymZeCHQ"},
					Expiration:           common.ToPointer(TimeUTC(*link.ValidUntil)),
					Id:                   link.ID,
					IssuedClaims:         link.IssuedClaims,
					MaxIssuance:          link.MaxIssuance,
					SchemaType:           link.Schema.Type,
					SchemaUrl:            link.Schema.URL,
					Status:               LinkStatusActive,
					ProofTypes:           []string{"Iden3commRevocationStatusV1", "BJJSignature2021"},
					CreatedAt:            TimeUTC(link.CreatedAt),
					SchemaHash:           string(hash),
					CredentialExpiration: common.ToPointer(TimeUTC(tomorrow)),
				},
			},
		},
		{
			name: "Happy path, link expired by date",
			auth: authOk,
			id:   linkExpired.ID,
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLink200JSONResponse{
					Active:               linkExpired.Active,
					CredentialSubject:    CredentialSubject{"birthday": 19791109, "documentType": 12, "type": schemaType, "id": "did:polygonid:polygon:mumbai:2qDDDKmo436EZGCBAvkqZjADYoNRJszkG7UymZeCHQ"},
					Expiration:           common.ToPointer(TimeUTC(*linkExpired.ValidUntil)),
					Id:                   linkExpired.ID,
					IssuedClaims:         linkExpired.IssuedClaims,
					MaxIssuance:          linkExpired.MaxIssuance,
					SchemaType:           linkExpired.Schema.Type,
					SchemaUrl:            linkExpired.Schema.URL,
					Status:               LinkStatusExceeded,
					ProofTypes:           []string{"Iden3commRevocationStatusV1", "BJJSignature2021"},
					CredentialExpiration: nil,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/credentials/links/%s", did, tc.id)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetLink200JSONResponse
				expected, ok := tc.expected.response.(GetLink200JSONResponse)
				require.True(t, ok)
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, expected.Active, response.Active)
				assert.Equal(t, expected.MaxIssuance, response.MaxIssuance)
				assert.Equal(t, expected.Status, response.Status)
				assert.Equal(t, expected.IssuedClaims, response.IssuedClaims)
				assert.Equal(t, expected.Id, response.Id)
				tcCred, err := json.Marshal(expected.CredentialSubject)
				require.NoError(t, err)
				respCred, err := json.Marshal(response.CredentialSubject)
				require.NoError(t, err)
				assert.Equal(t, tcCred, respCred)
				assert.Equal(t, expected.SchemaType, response.SchemaType)
				assert.Equal(t, expected.SchemaUrl, response.SchemaUrl)
				assert.Equal(t, expected.Active, response.Active)
				assert.InDelta(t, time.Time(*expected.Expiration).UnixMilli(), time.Time(*response.Expiration).UnixMilli(), 1000)
				assert.Equal(t, len(expected.ProofTypes), len(response.ProofTypes))
				if expected.CredentialExpiration != nil {
					tt := time.Time(*expected.CredentialExpiration)
					tt00 := common.ToPointer(TimeUTC(time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.UTC)))
					assert.Equal(t, tt00.String(), response.CredentialExpiration.String())
				}
			case http.StatusNotFound:
				var response GetLink404JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				expected, ok := tc.expected.response.(GetLink404JSONResponse)
				require.True(t, ok)
				assert.EqualValues(t, expected.Message, response.Message)
			}
		})
	}
}

func TestServer_GetAllLinks(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		sUrl       = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)
	importedSchema, err := server.Services.schema.ImportSchema(ctx, *did, ports.NewImportSchemaRequest(sUrl, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription")))
	assert.NoError(t, err)

	tomorrow := time.Now().Add(24 * time.Hour)
	yesterday := time.Now().Add(-24 * time.Hour)

	link1, err := server.Services.links.Save(ctx, *did, common.ToPointer(10), &tomorrow, importedSchema.ID, &tomorrow, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12},
		&verifiable.RefreshService{
			ID:   "https://refresh.xyz",
			Type: verifiable.Iden3RefreshService2023,
		},
		&verifiable.DisplayMethod{
			ID:   "https://display.xyz",
			Type: verifiable.Iden3BasicDisplayMethodV1,
		},
	)
	require.NoError(t, err)
	linkActive := getLinkResponse(*link1)

	time.Sleep(10 * time.Millisecond)

	link2, err := server.Services.links.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, &tomorrow, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12},
		&verifiable.RefreshService{
			ID:   "https://revreshv2.xyz",
			Type: verifiable.Iden3RefreshService2023,
		},
		&verifiable.DisplayMethod{
			ID:   "https://display.xyz",
			Type: verifiable.Iden3BasicDisplayMethodV1,
		},
	)
	require.NoError(t, err)
	linkExpired := getLinkResponse(*link2)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	link3, err := server.Services.links.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, &tomorrow, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	link3.Active = false
	require.NoError(t, err)
	require.NoError(t, server.Services.links.Activate(ctx, *did, link3.ID, false))
	linkInactive := getLinkResponse(*link3)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	handler := getHandler(ctx, server)
	type expected struct {
		response []Link
		httpCode int
	}
	type testConfig struct {
		name     string
		query    *string
		status   *GetLinksParamsStatus
		auth     func() (string, string)
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:   "Wrong type",
			auth:   authOk,
			status: common.ToPointer(GetLinksParamsStatus("unknown-filter-type")),
			expected: expected{
				httpCode: http.StatusBadRequest,
			},
		},

		{
			name: "Happy path. All schemas",
			auth: authOk,
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive, linkExpired, linkActive},
			},
		},
		{
			name:   "Happy path. All schemas, explicit",
			auth:   authOk,
			status: common.ToPointer(GetLinksParamsStatus("all")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive, linkExpired, linkActive},
			},
		},
		{
			name:   "Happy path. All schemas, active",
			auth:   authOk,
			status: common.ToPointer(GetLinksParamsStatus("active")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkActive},
			},
		},
		{
			name:   "Happy path. All schemas, exceeded",
			auth:   authOk,
			status: common.ToPointer(GetLinksParamsStatus("exceeded")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive, linkExpired},
			},
		},
		{
			name:   "Happy path. All schemas, exceeded",
			auth:   authOk,
			status: common.ToPointer(GetLinksParamsStatus("inactive")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive},
			},
		},
		{
			name:   "Exceeded with filter found",
			auth:   authOk,
			query:  common.ToPointer("documentType"),
			status: common.ToPointer(GetLinksParamsStatus("exceeded")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive, linkExpired},
			},
		},
		{
			name:   "Exceeded with filter found, partial match",
			auth:   authOk,
			query:  common.ToPointer("docum"),
			status: common.ToPointer(GetLinksParamsStatus("exceeded")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive, linkExpired},
			},
		},
		{
			name:   "Empty result",
			auth:   authOk,
			query:  common.ToPointer("documentType"),
			status: common.ToPointer(GetLinksParamsStatus("exceeded")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive, linkExpired},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			endpoint := url.URL{Path: fmt.Sprintf("/v2/identities/%s/credentials/links", did)}
			if tc.status != nil {
				endpoint.RawQuery = endpoint.RawQuery + "status=" + string(*tc.status)
			}
			if tc.query != nil {
				endpoint.RawQuery = endpoint.RawQuery + "&query=" + *tc.query
			}

			req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetLinks200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				if assert.Equal(t, len(tc.expected.response), len(response)) {
					for i, resp := range response {
						assert.Equal(t, tc.expected.response[i].Id, resp.Id)
						assert.Equal(t, tc.expected.response[i].Status, resp.Status)
						assert.Equal(t, tc.expected.response[i].IssuedClaims, resp.IssuedClaims)
						assert.Equal(t, tc.expected.response[i].Active, resp.Active)
						assert.Equal(t, tc.expected.response[i].MaxIssuance, resp.MaxIssuance)
						assert.Equal(t, tc.expected.response[i].SchemaUrl, resp.SchemaUrl)
						assert.Equal(t, tc.expected.response[i].SchemaType, resp.SchemaType)
						assert.Equal(t, tc.expected.response[i].RefreshService, resp.RefreshService)
						tcCred, err := json.Marshal(tc.expected.response[i].CredentialSubject)
						require.NoError(t, err)
						respCred, err := json.Marshal(tc.expected.response[i].CredentialSubject)
						require.NoError(t, err)
						assert.Equal(t, tcCred, respCred)
						assert.InDelta(t, time.Time(*tc.expected.response[i].Expiration).UnixMilli(), time.Time(*resp.Expiration).UnixMilli(), 1000)
						expectCredExpiration := common.ToPointer(TimeUTC(time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC)))
						assert.Equal(t, expectCredExpiration.String(), resp.CredentialExpiration.String())
					}
				}
			case http.StatusBadRequest:
				var response GetLinks400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
			}
		})
	}
}

func TestServer_DeleteLink(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)
	importedSchema, err := server.Services.schema.ImportSchema(ctx, *did, ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription")))
	assert.NoError(t, err)

	validUntil := common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local))
	credentialExpiration := common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local))
	link, err := server.Services.links.Save(ctx, *did, common.ToPointer(10), validUntil, importedSchema.ID, credentialExpiration, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	assert.NoError(t, err)
	handler := getHandler(ctx, server)

	type expected struct {
		response DeleteLinkResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		id       uuid.UUID
		auth     func() (string, string)
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			id:   link.ID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Claim link does not exist",
			auth: authOk,
			id:   uuid.New(),
			expected: expected{
				response: DeleteLink400JSONResponse{N400JSONResponse{Message: "link does not exist"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Happy path",
			auth: authOk,
			id:   link.ID,
			expected: expected{
				response: DeleteLink200JSONResponse{Message: "link deleted"},
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/credentials/links/%s", did, tc.id)

			req, err := http.NewRequest(http.MethodDelete, url, tests.JSONBody(t, nil))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GenericMessage
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				expected, ok := tc.expected.response.(DeleteLink200JSONResponse)
				assert.True(t, ok)
				assert.Equal(t, expected.Message, response.Message)

			case http.StatusBadRequest:
				var response DeleteLink400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_DeleteLinkForDifferentDID(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	iden2, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)
	did2, err := w3c.ParseDID(iden2.Identifier)
	require.NoError(t, err)
	importedSchema, err := server.Services.schema.ImportSchema(ctx, *did, ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription")))
	assert.NoError(t, err)

	validUntil := common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local))
	credentialExpiration := common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local))
	link, err := server.Services.links.Save(ctx, *did, common.ToPointer(10), validUntil, importedSchema.ID, credentialExpiration, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	assert.NoError(t, err)
	handler := getHandler(ctx, server)

	type expected struct {
		response DeleteLinkResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		id       uuid.UUID
		auth     func() (string, string)
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			id:   link.ID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Claim link does not exist for a did",
			auth: authOk,
			id:   link.ID,
			expected: expected{
				response: DeleteLink400JSONResponse{N400JSONResponse{Message: "link does not exist"}},
				httpCode: http.StatusBadRequest,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/credentials/links/%s", did2, tc.id)

			req, err := http.NewRequest(http.MethodDelete, url, tests.JSONBody(t, nil))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GenericMessage
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				expected, ok := tc.expected.response.(DeleteLink200JSONResponse)
				assert.True(t, ok)
				assert.Equal(t, expected.Message, response.Message)

			case http.StatusBadRequest:
				var response DeleteLink400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_CreateLinkQRCode(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)
	importedSchema, err := server.Services.schema.ImportSchema(ctx, *did, ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription")))
	assert.NoError(t, err)

	validUntil := common.ToPointer(time.Now().Add(365 * 24 * time.Hour))
	credentialExpiration := common.ToPointer(validUntil.Add(365 * 24 * time.Hour))

	link, err := server.Services.links.Save(ctx, *did, common.ToPointer(10), validUntil, importedSchema.ID, credentialExpiration, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	assert.NoError(t, err)

	yesterday := time.Now().Add(-24 * time.Hour)
	linkExpired, err := server.Services.links.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, nil, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	linkDetail := getLinkResponse(*link)

	type expected struct {
		linkDetail Link
		httpCode   int
		message    string
	}

	type testConfig struct {
		name     string
		request  CreateLinkQrCodeRequestObject
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:    "Wrong link id",
			request: CreateLinkQrCodeRequestObject{Id: uuid.New()},
			expected: expected{
				httpCode: http.StatusNotFound,
				message:  "error: link not found",
			},
		},
		{
			name: "Expired link",
			request: CreateLinkQrCodeRequestObject{
				Id: linkExpired.ID,
			},
			expected: expected{
				httpCode: http.StatusNotFound,
				message:  "error: cannot issue a credential for an expired link",
			},
		},
		{
			name: "Happy path",
			request: CreateLinkQrCodeRequestObject{
				Id: link.ID,
			},
			expected: expected{
				linkDetail: linkDetail,
				httpCode:   http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			apiURL := fmt.Sprintf("/v2/identities/%s/credentials/links/%s/qrcode", did, tc.request.Id.String())

			req, err := http.NewRequest(http.MethodPost, apiURL, tests.JSONBody(t, nil))
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				callBack := cfg.ServerUrl + fmt.Sprintf("/v2/identities/%s/credentials/links/callback?", iden.Identifier)
				var response CreateLinkQrCode200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))

				realQR := protocol.AuthorizationRequestMessage{}

				qrLink := checkQRFetchURLForLinks(t, response.DeepLink)

				// Let's see that universal link is correct
				assert.Equal(t, server.cfg.UniversalLinks.BaseUrl+"#request_uri="+qrLink, response.UniversalLink)

				// Now let's fetch the original QR using the url
				rr := httptest.NewRecorder()
				req, err := http.NewRequest(http.MethodGet, qrLink, nil)
				require.NoError(t, err)
				handler.ServeHTTP(rr, req)
				require.Equal(t, http.StatusOK, rr.Code)

				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &realQR))

				assert.NotNil(t, realQR.Body)
				assert.Equal(t, "authentication", realQR.Body.Reason)
				callbackArr := strings.Split(realQR.Body.CallbackURL, "linkID")
				assert.True(t, len(callbackArr) == 2)
				assert.Equal(t, callBack, callbackArr[0])
				assert.Equal(t, "https://iden3-communication.io/authorization/1.0/request", string(realQR.Type))
				assert.Equal(t, "application/iden3comm-plain-json", string(realQR.Typ))
				assert.Equal(t, did.String(), realQR.From)
				assert.NotNil(t, realQR.ThreadID)
				assert.NotNil(t, response.SessionID)
				assert.Equal(t, tc.expected.linkDetail.Id, response.LinkDetail.Id)
				assert.Equal(t, tc.expected.linkDetail.SchemaType, response.LinkDetail.SchemaType)

			case http.StatusNotFound:
				var response CreateLinkQrCode404JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.message, response.Message)
			}
		})
	}
}
