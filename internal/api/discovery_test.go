package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_DiscoverQuery(t *testing.T) {
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	type expected struct {
		httpCode     int
		responseBody string
	}
	type testConfig struct {
		name      string
		expected  expected
		queryJSON string
	}

	for _, tc := range []testConfig{
		{
			name: "accept query",
			queryJSON: `{
				"id": "4391deb9-9d76-4b97-9b57-2a0f7f6c883e",
				"thid": "4391deb9-9d76-4b97-9b57-2a0f7f6c883e",
				"typ": "application/iden3comm-plain-json",
				"type": "https://didcomm.org/discover-features/2.0/queries",
				"body": {
				  "queries": [
					{
					  "feature-type": "accept"
					}
				  ]
				},
				"created_time": 1738071909
			  }`,
			expected: expected{
				httpCode: http.StatusOK,
				responseBody: `{
					  "disclosures": [
						{
						  "feature-type": "accept",
						  "id": "iden3comm/v1;env=application/iden3comm-plain-json"
						}
					  ]
				  }`,
			},
		},
		{
			name: "protocol query with match revocation*",
			queryJSON: `{
				"id": "4391deb9-9d76-4b97-9b57-2a0f7f6c883e",
				"thid": "4391deb9-9d76-4b97-9b57-2a0f7f6c883e",
				"typ": "application/iden3comm-plain-json",
				"type": "https://didcomm.org/discover-features/2.0/queries",
				"body": {
				  "queries": [
					{
					  "feature-type": "protocol",
					  "match": "https://iden3-communication.io/revocation/*"
					}
				  ]
				},
				"created_time": 1738071909
			  }`,
			expected: expected{
				httpCode: http.StatusOK,
				responseBody: `{
					"disclosures": [
					  {
						"feature-type": "protocol",
						"id": "https://iden3-communication.io/revocation/1.0/request-status"
					  }
					]
				  }`,
			},
		},
		{
			name: "header query with match `typ`",
			queryJSON: `{
				"id": "4391deb9-9d76-4b97-9b57-2a0f7f6c883e",
				"thid": "4391deb9-9d76-4b97-9b57-2a0f7f6c883e",
				"typ": "application/iden3comm-plain-json",
				"type": "https://didcomm.org/discover-features/2.0/queries",
				"body": {
				  "queries": [
					{
					  "feature-type": "header",
					  "match": "typ"
					}
				  ]
				},
				"created_time": 1738071909
			  }`,
			expected: expected{
				httpCode: http.StatusOK,
				responseBody: `{
					"disclosures": [
					  {
						"feature-type": "header",
						"id": "typ"
					  }
					]
				  }`,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := "/v2/agent"

			req, err := http.NewRequest("POST", url, strings.NewReader(tc.queryJSON))
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response protocol.DiscoverFeatureDiscloseMessage
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				responseBody, err := json.Marshal(response.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tc.expected.responseBody, string(responseBody))
			default:
				t.Fail()
			}
		})
	}
}
