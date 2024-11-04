package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_GetSupportedNetworks(t *testing.T) {
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	type expected struct {
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
			name: "should return supported networks",
			auth: authOk,
			expected: expected{
				httpCode: 200,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			req, err := http.NewRequest("GET", "/v2/supported-networks", nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			if tc.expected.httpCode == http.StatusOK {
				var response GetSupportedNetworks200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, 1, len(response))
				assert.Equal(t, "polygon", response[0].Blockchain)
				assert.Equal(t, []NetworkData{
					{
						Name:             "amoy",
						CredentialStatus: []string{"Iden3commRevocationStatusV1.0"},
					},
				}, response[0].Networks)
			}
		})
	}
}
