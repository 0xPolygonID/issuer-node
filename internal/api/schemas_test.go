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
			req, err := http.NewRequest("GET", fmt.Sprintf("/v1/identities/%s/schemas/%s", issuerDID, tc.id), nil)
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
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		query    *string
		expected expected
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
			},
		},
		{
			name:  "Happy path. All schemas, query=''",
			auth:  authOk,
			query: common.ToPointer(""),
			expected: expected{
				httpCode: http.StatusOK,
				count:    21,
			},
		},
		{
			name:  "Happy path. Search for schema type. All",
			auth:  authOk,
			query: common.ToPointer("schemaType"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    20,
			},
		},
		{
			name:  "Happy path. Search for one schema but many attr. Return all",
			auth:  authOk,
			query: common.ToPointer("schemaType-11 attr1"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    21,
			},
		},
		{
			name:  "Exact search, return 1",
			auth:  authOk,
			query: common.ToPointer("UbiProgram"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    1,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			endpoint := fmt.Sprintf("/v1/identities/%s/schemas", issuerDID)
			if tc.query != nil {
				endpoint = endpoint + "?query=" + url.QueryEscape(*tc.query)
			}
			req, err := http.NewRequest("GET", endpoint, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetSchemas200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.count, len(response))
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
			req, err := http.NewRequest("POST", fmt.Sprintf("/v1/identities/%s/schemas", issuerDID), tests.JSONBody(t, tc.request))
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
			req, err := http.NewRequest("POST", fmt.Sprintf("/v1/identities/%s/schemas", issuerDID), tests.JSONBody(t, tc.request))
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
