package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestServer_GetSchema(t *testing.T) {
	ctx := context.Background()
	server := newTestServer(t, nil)
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.ServerUrl = "https://testing.env"
	fixture := repositories.NewFixture(storage)

	s := &domain.Schema{
		ID:        uuid.New(),
		IssuerDID: *issuerDID,
		URL:       "https://domain.org/this/is/an/url",
		Type:      "schemaType",
		Words:     domain.SchemaWordsFromString("attr1, attr2, attr3"),
		CreatedAt: time.Now(),
	}
	s.Hash = common.CreateSchemaHash([]byte(s.URL + "#" + s.Type))
	fixture.CreateSchema(t, ctx, s)
	sHash, _ := s.Hash.MarshalText()

	handler := getHandler(ctx, server)
	type expected struct {
		httpCode int
		errorMsg string
		schema   *Schema
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		id       string
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "Not authorized",
			auth: authWrong,
			id:   uuid.NewString(),
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Invalid uuid",
			auth: authOk,
			id:   "someInvalidDID",
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "Invalid format for parameter id: error unmarshaling 'someInvalidDID' text as *uuid.UUID: invalid UUID length: 14",
			},
		},
		{
			name: "Non existing uuid",
			auth: authOk,
			id:   uuid.NewString(),
			expected: expected{
				httpCode: http.StatusNotFound,
				errorMsg: "schema not found",
			},
		},
		{
			name: "Happy path. Existing schema",
			auth: authOk,
			id:   s.ID.String(),
			expected: expected{
				httpCode: http.StatusOK,
				schema: &Schema{
					BigInt:    s.Hash.BigInt().String(),
					CreatedAt: TimeUTC(s.CreatedAt),
					Hash:      string(sHash),
					Id:        s.ID.String(),
					Type:      s.Type,
					Url:       s.URL,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", fmt.Sprintf("/v2/identities/%s/schemas/%s", issuerDID, tc.id), nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetSchema200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.schema.Id, response.Id)
				assert.Equal(t, tc.expected.schema.BigInt, response.BigInt)
				assert.Equal(t, tc.expected.schema.Type, response.Type)
				assert.Equal(t, tc.expected.schema.Url, response.Url)
				assert.Equal(t, tc.expected.schema.Hash, response.Hash)
				assert.InDelta(t, time.Time(tc.expected.schema.CreatedAt).UnixMilli(), time.Time(response.CreatedAt).UnixMilli(), 1000)
			case http.StatusNotFound:
				var response GetSchema404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.errorMsg, response.Message)
			case http.StatusBadRequest:
				var response GetSchema400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.errorMsg, response.Message)
			}
		})
	}
}

// Refer to the schema repository tests for more deep test related to Postgres Full Text Search
func TestServer_GetSchemas(t *testing.T) {
	ctx := context.Background()
	// Need an isolated DB here to avoid other tests side effects
	conn := lookupPostgresURL()
	if conn == "" {
		conn = "postgres://postgres:postgres@localhost:5435"
	}
	storage, teardown, err := tests.NewTestStorage(&config.Configuration{Database: config.Database{URL: conn}})
	require.NoError(t, err)
	defer teardown()

	server := newTestServer(t, storage)

	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.ServerUrl = "https://testing.env"
	fixture := repositories.NewFixture(storage)

	for i := 0; i < 20; i++ {
		s := &domain.Schema{
			ID:        uuid.New(),
			IssuerDID: *issuerDID,
			URL:       fmt.Sprintf("https://domain.org/this/is/an/url/%d", i),
			Type:      fmt.Sprintf("schemaType-%d", i),
			Words:     domain.SchemaWordsFromString("attr1, attr2, attr3"),
			CreatedAt: time.Now(),
		}
		s.Hash = common.CreateSchemaHash([]byte(s.URL + "#" + s.Type))
		fixture.CreateSchema(t, ctx, s)
	}
	s := &domain.Schema{
		ID:        uuid.New(),
		IssuerDID: *issuerDID,
		URL:       "https://domain.org/this/is/an/url/ubiprogram",
		Type:      "UbiProgram",
		Words:     domain.SchemaWordsFromString("attr1, attr2, attr3"),
		CreatedAt: time.Now(),
	}
	s.Hash = common.CreateSchemaHash([]byte(s.URL + "#" + s.Type))
	fixture.CreateSchema(t, ctx, s)

	handler := getHandler(ctx, server)
	type expected struct {
		httpCode int
		count    int
		response GetSchemasResponseObject
	}

	type pagination struct {
		page       *int
		maxResults *int
	}

	type testConfig struct {
		name       string
		auth       func() (string, string)
		query      *string
		pagination *pagination
		expected   expected
	}
	for _, tc := range []testConfig{
		{
			name: "Not authorized",
			auth: authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:  "Happy path. All schemas, no query",
			auth:  authOk,
			query: nil,
			expected: expected{
				httpCode: http.StatusOK,
				count:    21,
				response: GetSchemas200JSONResponse{
					Meta: PaginatedMetadata{
						Total:      21,
						Page:       1,
						MaxResults: 50,
					},
				},
			},
		},
		{
			name:  "Happy path. All schemas, no query - maxResults 10 - page 1",
			auth:  authOk,
			query: nil,
			pagination: &pagination{
				page:       common.ToPointer(1),
				maxResults: common.ToPointer(10),
			},
			expected: expected{
				httpCode: http.StatusOK,
				count:    10,
				response: GetSchemas200JSONResponse{
					Meta: PaginatedMetadata{
						Total:      21,
						Page:       1,
						MaxResults: 10,
					},
				},
			},
		},
		{
			name:  "Happy path. All schemas, no query - maxResults 10 - default page 1",
			auth:  authOk,
			query: nil,
			pagination: &pagination{
				maxResults: common.ToPointer(10),
			},
			expected: expected{
				httpCode: http.StatusOK,
				count:    10,
				response: GetSchemas200JSONResponse{
					Meta: PaginatedMetadata{
						Total:      21,
						Page:       1,
						MaxResults: 10,
					},
				},
			},
		},
		{
			name:  "Happy path. All schemas, no query - default maxResults 50 - page 1",
			auth:  authOk,
			query: nil,
			pagination: &pagination{
				page: common.ToPointer(1),
			},
			expected: expected{
				httpCode: http.StatusOK,
				count:    21,
				response: GetSchemas200JSONResponse{
					Meta: PaginatedMetadata{
						Total:      21,
						Page:       1,
						MaxResults: 50,
					},
				},
			},
		},
		{
			name:  "Happy path. All schemas, query=''",
			auth:  authOk,
			query: common.ToPointer(""),
			expected: expected{
				httpCode: http.StatusOK,
				count:    21,
				response: GetSchemas200JSONResponse{
					Meta: PaginatedMetadata{
						Total:      21,
						Page:       1,
						MaxResults: 50,
					},
				},
			},
		},
		{
			name:  "Happy path. Search for schema type. All",
			auth:  authOk,
			query: common.ToPointer("schemaType"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    20,
				response: GetSchemas200JSONResponse{
					Meta: PaginatedMetadata{
						Total:      20,
						Page:       1,
						MaxResults: 50,
					},
				},
			},
		},
		{
			name:  "Happy path. Search for one schema but many attr. Return all",
			auth:  authOk,
			query: common.ToPointer("schemaType-11 attr1"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    21,
				response: GetSchemas200JSONResponse{
					Meta: PaginatedMetadata{
						Total:      21,
						Page:       1,
						MaxResults: 50,
					},
				},
			},
		},
		{
			name:  "Exact search, return 1",
			auth:  authOk,
			query: common.ToPointer("UbiProgram"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    1,
				response: GetSchemas200JSONResponse{
					Meta: PaginatedMetadata{
						Total:      1,
						Page:       1,
						MaxResults: 50,
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			endpoint := fmt.Sprintf("/v2/identities/%s/schemas", issuerDID)
			if tc.query != nil {
				endpoint = endpoint + "?query=" + url.QueryEscape(*tc.query)
			}
			if tc.pagination != nil {
				if tc.pagination.page != nil {
					if tc.query != nil {
						endpoint = endpoint + fmt.Sprintf("&page=%d", *tc.pagination.page)
					} else {
						endpoint = endpoint + fmt.Sprintf("?page=%d", *tc.pagination.page)
					}
					if tc.pagination.maxResults != nil {
						endpoint = endpoint + fmt.Sprintf("&max_results=%d", *tc.pagination.maxResults)
					}
				} else {
					if tc.query != nil {
						endpoint = endpoint + fmt.Sprintf("&max_results=%d", *tc.pagination.maxResults)
					} else {
						endpoint = endpoint + fmt.Sprintf("?max_results=%d", *tc.pagination.maxResults)
					}
				}
			}
			req, err := http.NewRequest("GET", endpoint, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetSchemas200JSONResponse
				expectedResponse, ok := tc.expected.response.(GetSchemas200JSONResponse)
				require.True(t, ok)
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.count, len(response.Items))
				assert.Equal(t, expectedResponse.Meta.Total, response.Meta.Total)
				assert.Equal(t, expectedResponse.Meta.Page, response.Meta.Page)
				assert.Equal(t, expectedResponse.Meta.MaxResults, response.Meta.MaxResults)
			}
		})
	}
}

func TestServer_ImportSchema(t *testing.T) {
	const url = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	const schemaType = "KYCCountryOfResidenceCredential"

	ctx := context.Background()

	server := newTestServer(t, nil)
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.ServerUrl = "https://testing.env"

	handler := getHandler(ctx, server)

	type expected struct {
		httpCode int
		errorMsg string
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		request  *ImportSchemaJSONRequestBody
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name:    "Not authorized",
			auth:    authWrong,
			request: nil,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:    "Empty request",
			auth:    authOk,
			request: nil,
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "bad request: empty url",
			},
		},
		{
			name: "Wrong url",
			auth: authOk,
			request: &ImportSchemaRequest{
				SchemaType:  "lala",
				Url:         "wrong/url",
				Title:       common.ToPointer("some Title"),
				Description: common.ToPointer("some Description"),
				Version:     uuid.NewString(),
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "bad request: parsing url: parse \"wrong/url\": invalid URI for request",
			},
		},
		{
			name: "Valid request",
			auth: authOk,
			request: &ImportSchemaRequest{
				SchemaType:  schemaType,
				Url:         url,
				Title:       common.ToPointer("some Title"),
				Description: common.ToPointer("some Description"),
				Version:     uuid.NewString(),
			},
			expected: expected{
				httpCode: http.StatusCreated,
				errorMsg: "bad request: parsing url: parse \"wrong/url\": invalid URI for request",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("POST", fmt.Sprintf("/v2/identities/%s/schemas", issuerDID), tests.JSONBody(t, tc.request))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response ImportSchema201JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				_, err := uuid.Parse(response.Id)
				assert.NoError(t, err)
			case http.StatusBadRequest:
				var response ImportSchema400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.errorMsg, response.Message)
			}
		})
	}
}

func TestServer_ImportSchemaIPFS(t *testing.T) {
	// TIP: A copy of the files here internal/api_ui/testdata/ipfs-schema-1.json
	const url = "ipfs://QmQVeb5dkz5ekDqBrYVVxBFQZoCbzamnmMUn9B8twCEgDL"
	const schemaType = "testNewType"
	ctx := context.Background()

	server := newTestServer(t, nil)
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.ServerUrl = "https://testing.env"

	handler := getHandler(ctx, server)

	type expected struct {
		httpCode int
		errorMsg string
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		request  *ImportSchemaJSONRequestBody
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name:    "Not authorized",
			auth:    authWrong,
			request: nil,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:    "Empty request",
			auth:    authOk,
			request: nil,
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "bad request: empty url",
			},
		},
		{
			name: "Wrong url",
			auth: authOk,
			request: &ImportSchemaRequest{
				SchemaType:  "lala",
				Url:         "wrong/url",
				Title:       common.ToPointer("title"),
				Description: common.ToPointer("description"),
				Version:     "1.0.0",
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "bad request: parsing url: parse \"wrong/url\": invalid URI for request",
			},
		},
		{
			name: "Valid request",
			auth: authOk,
			request: &ImportSchemaRequest{
				SchemaType:  schemaType,
				Url:         url,
				Title:       common.ToPointer("title"),
				Description: common.ToPointer("description"),
				Version:     "1.0.0",
			},
			expected: expected{
				httpCode: http.StatusCreated,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("POST", fmt.Sprintf("/v2/identities/%s/schemas", issuerDID), tests.JSONBody(t, tc.request))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response ImportSchema201JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				_, err := uuid.Parse(response.Id)
				assert.NoError(t, err)
			case http.StatusBadRequest:
				var response ImportSchema400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.errorMsg, response.Message)
			}
		})
	}
}

func TestServer_GetSchemasWithPaginationAndSort(t *testing.T) {
	ctx := context.Background()
	conn := lookupPostgresURL()
	if conn == "" {
		conn = "postgres://postgres:postgres@localhost:5435"
	}
	storage, teardown, err := tests.NewTestStorage(&config.Configuration{Database: config.Database{URL: conn}})
	require.NoError(t, err)
	defer teardown()

	server := newTestServer(t, storage)

	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.ServerUrl = "https://testing.env"
	fixture := repositories.NewFixture(storage)
	handler := getHandler(ctx, server)

	s1 := &domain.Schema{
		ID:        uuid.New(),
		IssuerDID: *issuerDID,
		URL:       "https://domain.org/this/is/an/SchemaA",
		Type:      "SchemaA",
		Words:     domain.SchemaWordsFromString("attr1, attr2, attr3"),
		CreatedAt: time.Now(),
		Version:   "1",
	}
	s1.Hash = common.CreateSchemaHash([]byte(s1.URL + "#" + s1.Type))
	fixture.CreateSchema(t, ctx, s1)

	s2 := &domain.Schema{
		ID:        uuid.New(),
		IssuerDID: *issuerDID,
		URL:       "https://domain.org/this/is/an/SchemaB",
		Type:      "SchemaB",
		Words:     domain.SchemaWordsFromString("attr1, attr2, attr3"),
		CreatedAt: time.Now(),
		Version:   "2",
	}
	s2.Hash = common.CreateSchemaHash([]byte(s2.URL + "#" + s2.Type))
	fixture.CreateSchema(t, ctx, s2)

	t.Run("Happy path. All schemas, no query - maxResults 10 - page 1 - sort by created_at desc default", func(t *testing.T) {
		rr := httptest.NewRecorder()
		endpoint := fmt.Sprintf("/v2/identities/%s/schemas", issuerDID)
		req, err := http.NewRequest("GET", endpoint, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetSchemas200JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Len(t, response.Items, 2)
		assert.Equal(t, s2.ID.String(), response.Items[0].Id)
		assert.Equal(t, s1.ID.String(), response.Items[1].Id)
	})

	t.Run("Happy path. All schemas, no query - maxResults 10 - page 1 - sort by created_at desc", func(t *testing.T) {
		rr := httptest.NewRecorder()
		endpoint := fmt.Sprintf("/v2/identities/%s/schemas?sort=-importDate", issuerDID)
		req, err := http.NewRequest("GET", endpoint, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetSchemas200JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Len(t, response.Items, 2)
		assert.Equal(t, s2.ID.String(), response.Items[0].Id)
		assert.Equal(t, s1.ID.String(), response.Items[1].Id)
	})

	t.Run("Happy path. All schemas, no query - maxResults 10 - page 1 - sort by created_at asc", func(t *testing.T) {
		rr := httptest.NewRecorder()
		endpoint := fmt.Sprintf("/v2/identities/%s/schemas?sort=importDate", issuerDID)
		req, err := http.NewRequest("GET", endpoint, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetSchemas200JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Len(t, response.Items, 2)
		assert.Equal(t, s2.ID.String(), response.Items[1].Id)
		assert.Equal(t, s1.ID.String(), response.Items[0].Id)
	})

	t.Run("Happy path. All schemas, no query - maxResults 10 - page 1 - sort by schemaType desc", func(t *testing.T) {
		rr := httptest.NewRecorder()
		endpoint := fmt.Sprintf("/v2/identities/%s/schemas?sort=-schemaType", issuerDID)
		req, err := http.NewRequest("GET", endpoint, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetSchemas200JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Len(t, response.Items, 2)
		assert.Equal(t, s2.ID.String(), response.Items[0].Id)
		assert.Equal(t, s1.ID.String(), response.Items[1].Id)
	})

	t.Run("Happy path. All schemas, no query - maxResults 10 - page 1 - sort by schemaType asc", func(t *testing.T) {
		rr := httptest.NewRecorder()
		endpoint := fmt.Sprintf("/v2/identities/%s/schemas?sort=schemaType", issuerDID)
		req, err := http.NewRequest("GET", endpoint, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetSchemas200JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Len(t, response.Items, 2)
		assert.Equal(t, s2.ID.String(), response.Items[1].Id)
		assert.Equal(t, s1.ID.String(), response.Items[0].Id)
	})

	t.Run("Happy path. All schemas, no query - maxResults 10 - page 1 - sort by version desc", func(t *testing.T) {
		rr := httptest.NewRecorder()
		endpoint := fmt.Sprintf("/v2/identities/%s/schemas?sort=-schemaVersion", issuerDID)
		req, err := http.NewRequest("GET", endpoint, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetSchemas200JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Len(t, response.Items, 2)
		assert.Equal(t, s2.ID.String(), response.Items[0].Id)
		assert.Equal(t, s1.ID.String(), response.Items[1].Id)
	})

	t.Run("Happy path. All schemas, no query - maxResults 10 - page 1 - sort by version asc", func(t *testing.T) {
		rr := httptest.NewRecorder()
		endpoint := fmt.Sprintf("/v2/identities/%s/schemas?sort=schemaVersion", issuerDID)
		req, err := http.NewRequest("GET", endpoint, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetSchemas200JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Len(t, response.Items, 2)
		assert.Equal(t, s2.ID.String(), response.Items[1].Id)
		assert.Equal(t, s1.ID.String(), response.Items[0].Id)
	})

	t.Run("Happy path. All schemas, with query - maxResults 10 - page 1 - sort by schemaType desc", func(t *testing.T) {
		rr := httptest.NewRecorder()
		endpoint := fmt.Sprintf("/v2/identities/%s/schemas?query=SchemaA&sort=-schemaType", issuerDID)
		req, err := http.NewRequest("GET", endpoint, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetSchemas200JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Len(t, response.Items, 1)
		assert.Equal(t, s1.ID.String(), response.Items[0].Id)
	})

	t.Run("Happy path. All schemas, with query2 - maxResults 10 - page 1 - sort by schemaType desc", func(t *testing.T) {
		rr := httptest.NewRecorder()
		endpoint := fmt.Sprintf("/v2/identities/%s/schemas?query=Schema&sort=-schemaType", issuerDID)
		req, err := http.NewRequest("GET", endpoint, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetSchemas200JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Len(t, response.Items, 2)
		assert.Equal(t, s2.ID.String(), response.Items[0].Id)
		assert.Equal(t, s1.ID.String(), response.Items[1].Id)
	})
}
