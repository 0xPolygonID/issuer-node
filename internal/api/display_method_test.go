package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
)

func TestServer_CreateDisplayMethod(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response CreateDisplayMethodResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		body     CreateDisplayMethodRequest
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
			body: CreateDisplayMethodRequest{
				Name:    "test",
				Url:     "http://test.com",
				Default: true,
			},
			expected: expected{
				response: CreateDisplayMethod201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "missing name",
			auth: authOk,
			body: CreateDisplayMethodRequest{
				Url:     "http://test.com",
				Default: true,
			},
			expected: expected{
				response: CreateDisplayMethod400JSONResponse{
					N400JSONResponse: N400JSONResponse{Message: "name is required"},
				},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "missing url",
			auth: authOk,
			body: CreateDisplayMethodRequest{
				Name:    "test",
				Default: true,
			},
			expected: expected{
				response: CreateDisplayMethod400JSONResponse{
					N400JSONResponse: N400JSONResponse{Message: "url is required"},
				},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "second default display method",
			auth: authOk,
			body: CreateDisplayMethodRequest{
				Name:    "test",
				Url:     "http://test.com",
				Default: true,
			},
			expected: expected{
				response: CreateDisplayMethod400JSONResponse{
					N400JSONResponse: N400JSONResponse{Message: "Duplicated default display method"},
				},
				httpCode: http.StatusBadRequest,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/display-method", did)
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

func TestServer_UpdateDisplayMethod(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	displayMethodToUpdateID, err := server.Services.displayMethod.Save(ctx, *did, "test", "http://test.com", true)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response UpdateDisplayMethodResponseObject
		httpCode int
	}

	type testConfig struct {
		name            string
		auth            func() (string, string)
		body            UpdateDisplayMethodJSONRequestBody
		displayMethodID *uuid.UUID
		expected        expected
	}

	for _, tc := range []testConfig{
		{
			name:            "No auth header",
			auth:            authWrong,
			displayMethodID: displayMethodToUpdateID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:            "update name",
			auth:            authOk,
			displayMethodID: displayMethodToUpdateID,
			body: UpdateDisplayMethodJSONRequestBody{
				Name: common.ToPointer("tes2"),
			},
			expected: expected{
				response: UpdateDisplayMethod200JSONResponse{},
				httpCode: http.StatusOK,
			},
		},
		{
			name:            "update empty name",
			auth:            authOk,
			displayMethodID: displayMethodToUpdateID,
			body: UpdateDisplayMethodJSONRequestBody{
				Name: common.ToPointer(""),
			},
			expected: expected{
				response: UpdateDisplayMethod400JSONResponse{
					N400JSONResponse: N400JSONResponse{Message: "name cannot be empty"},
				},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name:            "update empty url",
			auth:            authOk,
			displayMethodID: displayMethodToUpdateID,
			body: UpdateDisplayMethodJSONRequestBody{
				Url: common.ToPointer(""),
			},
			expected: expected{
				response: UpdateDisplayMethod400JSONResponse{
					N400JSONResponse: N400JSONResponse{Message: "url cannot be empty"},
				},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name:            "update default",
			auth:            authOk,
			displayMethodID: displayMethodToUpdateID,
			body: UpdateDisplayMethodJSONRequestBody{
				Default: common.ToPointer(false),
			},
			expected: expected{
				response: UpdateDisplayMethod200JSONResponse{},
				httpCode: http.StatusOK,
			},
		},
		{
			name:            "update invalid id",
			auth:            authOk,
			displayMethodID: common.ToPointer(uuid.New()),
			body: UpdateDisplayMethodJSONRequestBody{
				Default: common.ToPointer(false),
			},
			expected: expected{
				response: UpdateDisplayMethod404JSONResponse{
					N404JSONResponse: N404JSONResponse{Message: "Invalid display method id"},
				},
				httpCode: http.StatusNotFound,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/display-method/%s", did, tc.displayMethodID.String())
			req, err := http.NewRequest(http.MethodPatch, url, tests.JSONBody(t, tc.body))
			require.NoError(t, err)
			req.SetBasicAuth(tc.auth())
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusBadRequest:
				var response UpdateDisplayMethod400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_GetDisplayMethodByID(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	displayMethodToGetID, err := server.Services.displayMethod.Save(ctx, *did, "test", "http://test.com", true)
	require.NoError(t, err)

	displayMethodToGet, err := server.Services.displayMethod.GetByID(ctx, *did, *displayMethodToGetID)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response GetDisplayMethodResponseObject
		httpCode int
	}

	type testConfig struct {
		name            string
		auth            func() (string, string)
		displayMethodID *uuid.UUID
		expected        expected
	}

	for _, tc := range []testConfig{
		{
			name:            "No auth header",
			auth:            authWrong,
			displayMethodID: displayMethodToGetID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:            "happy path",
			auth:            authOk,
			displayMethodID: displayMethodToGetID,
			expected: expected{
				response: GetDisplayMethod200JSONResponse{
					Id:      displayMethodToGet.ID,
					Name:    displayMethodToGet.Name,
					Url:     displayMethodToGet.URL,
					Default: displayMethodToGet.IsDefault,
				},
				httpCode: http.StatusOK,
			},
		},
		{
			name:            "get invalid id",
			auth:            authOk,
			displayMethodID: common.ToPointer(uuid.New()),
			expected: expected{
				response: GetDisplayMethod404JSONResponse{
					N404JSONResponse: N404JSONResponse{Message: "Display method not found"},
				},
				httpCode: http.StatusNotFound,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/display-method/%s", did, tc.displayMethodID.String())
			req, err := http.NewRequest(http.MethodGet, url, tests.JSONBody(t, nil))
			require.NoError(t, err)
			req.SetBasicAuth(tc.auth())
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetDisplayMethod200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			case http.StatusNotFound:
				var response GetDisplayMethod404JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_GetAllDisplayMethod(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	displayMethodID1, err := server.Services.displayMethod.Save(ctx, *did, "test", "http://test.com", false)
	require.NoError(t, err)

	displayMethodID2, err := server.Services.displayMethod.Save(ctx, *did, "test", "http://test.com", true)
	require.NoError(t, err)

	displayMethod1, err := server.Services.displayMethod.GetByID(ctx, *did, *displayMethodID1)
	require.NoError(t, err)

	displayMethod2, err := server.Services.displayMethod.GetByID(ctx, *did, *displayMethodID2)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response GetAllDisplayMethodResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
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
			name: "happy path",
			auth: authOk,
			expected: expected{
				response: GetAllDisplayMethod200JSONResponse{
					{
						Id:      displayMethod1.ID,
						Name:    displayMethod1.Name,
						Url:     displayMethod1.URL,
						Default: displayMethod1.IsDefault,
					},
					{
						Id:      displayMethod2.ID,
						Name:    displayMethod2.Name,
						Url:     displayMethod2.URL,
						Default: displayMethod2.IsDefault,
					},
				},
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/display-method", did)
			req, err := http.NewRequest(http.MethodGet, url, tests.JSONBody(t, nil))
			require.NoError(t, err)
			req.SetBasicAuth(tc.auth())
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetAllDisplayMethod200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			case http.StatusBadRequest:
				var response GetAllDisplayMethod400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_DeleteDisplayMethodByID(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	displayMethodToDeleteID, err := server.Services.displayMethod.Save(ctx, *did, "test", "http://test.com", false)
	require.NoError(t, err)

	defaultDisplayMethodToDeleteID, err := server.Services.displayMethod.Save(ctx, *did, "test", "http://test.com", true)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response DeleteDisplayMethodResponseObject
		httpCode int
	}

	type testConfig struct {
		name            string
		auth            func() (string, string)
		displayMethodID *uuid.UUID
		expected        expected
	}

	for _, tc := range []testConfig{
		{
			name:            "No auth header",
			auth:            authWrong,
			displayMethodID: displayMethodToDeleteID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:            "happy path",
			auth:            authOk,
			displayMethodID: displayMethodToDeleteID,
			expected: expected{
				response: DeleteDisplayMethod200JSONResponse{
					Message: "Display method deleted",
				},
				httpCode: http.StatusOK,
			},
		},
		{
			name:            "delete default display method",
			auth:            authOk,
			displayMethodID: defaultDisplayMethodToDeleteID,
			expected: expected{
				response: DeleteDisplayMethod400JSONResponse{
					N400JSONResponse: N400JSONResponse{Message: "Cannot delete default display method"},
				},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name:            "delete invalid id",
			auth:            authOk,
			displayMethodID: common.ToPointer(uuid.New()),
			expected: expected{
				response: DeleteDisplayMethod404JSONResponse{
					N404JSONResponse: N404JSONResponse{Message: "Display method not found"},
				},
				httpCode: http.StatusNotFound,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/display-method/%s", did, tc.displayMethodID.String())
			req, err := http.NewRequest(http.MethodDelete, url, tests.JSONBody(t, nil))
			require.NoError(t, err)
			req.SetBasicAuth(tc.auth())
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response DeleteDisplayMethod200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			case http.StatusBadRequest:
				var response DeleteDisplayMethod400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			case http.StatusNotFound:
				var response DeleteDisplayMethod404JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}
