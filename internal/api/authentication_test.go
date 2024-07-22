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
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
)

func TestServer_AuthCallback(t *testing.T) {
	server := newTestServer(t)
	handler := getHandler(context.Background(), server)

	type expected struct {
		httpCode int
		message  string
	}
	type testConfig struct {
		name      string
		expected  expected
		sessionID *uuid.UUID
	}

	for _, tc := range []testConfig{
		{
			name:      "should get an error no body",
			sessionID: common.ToPointer(uuid.New()),
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  "Cannot proceed with empty body",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := "/v1/authentication/callback"
			if tc.sessionID != nil {
				url += "?sessionID=" + tc.sessionID.String()
			}

			req, err := http.NewRequest("POST", url, strings.NewReader(``))
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusBadRequest:
				var response AuthCallback400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.message, response.Message)
			default:
				t.Fail()
			}
		})
	}
}

func TestServer_GetAuthenticationConnection(t *testing.T) {
	server := newTestServer(t)
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qKDJmySKNi4GD4vYdqfLb37MSTSijg77NoRZaKfDX")
	require.NoError(t, err)
	server.cfg.APIUI.IssuerDID = *issuerDID
	server.cfg.APIUI.ServerURL = "https://testing.env"
	handler := getHandler(context.Background(), server)

	fixture := tests.NewFixture(storage)
	conn := &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}
	fixture.CreateConnection(t, conn)
	require.NoError(t, err)

	sessionID := uuid.New()
	fixture.CreateUserAuthentication(t, conn.ID, sessionID, conn.CreatedAt)

	type expected struct {
		httpCode int
		message  string
		response GetAuthenticationConnection200JSONResponse
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		id       uuid.UUID
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "Not authorized",
			auth: authWrong,
			id:   uuid.New(),
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Session Not found",
			auth: authOk,
			id:   uuid.New(),
			expected: expected{
				httpCode: http.StatusNotFound,
				message:  services.ErrConnectionDoesNotExist.Error(),
			},
		},
		{
			name: "Happy path. Existing connection",
			auth: authOk,
			id:   sessionID,
			expected: expected{
				httpCode: http.StatusOK,
				response: GetAuthenticationConnection200JSONResponse{
					Connection: AuthenticationConnection{
						Id:         conn.ID.String(),
						IssuerID:   conn.IssuerDID.String(),
						CreatedAt:  TimeUTC(conn.CreatedAt),
						ModifiedAt: TimeUTC(conn.ModifiedAt),
						UserID:     conn.UserDID.String(),
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/authentication/sessions/%s", tc.id.String())
			req, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)
			req.SetBasicAuth(tc.auth())

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetAuthenticationConnection200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.response.Connection.Id, response.Connection.Id)
				assert.Equal(t, tc.expected.response.Connection.IssuerID, response.Connection.IssuerID)
				assert.InDelta(t, time.Time(tc.expected.response.Connection.CreatedAt).Unix(), time.Time(response.Connection.CreatedAt).Unix(), 100)
				assert.InDelta(t, time.Time(tc.expected.response.Connection.ModifiedAt).Unix(), time.Time(response.Connection.ModifiedAt).Unix(), 100)
				assert.Equal(t, tc.expected.response.Connection.UserID, response.Connection.UserID)
			case http.StatusNotFound:
				var response GetAuthenticationConnection404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.message, response.Message)
			}
		})
	}
}

func TestServer_AuthQRCode(t *testing.T) {
	server := newTestServer(t)
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.APIUI.IssuerDID = *issuerDID
	server.cfg.APIUI.ServerURL = "https://testing.env"
	handler := getHandler(context.Background(), server)

	did, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)

	type expected struct {
		httpCode   int
		qrWithLink bool
		response   protocol.AuthorizationRequestMessage
	}
	type testConfig struct {
		name     string
		request  AuthQRCodeRequestObject
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "should get a qr code with a link by default",
			request: AuthQRCodeRequestObject{
				Body: &AuthQRCodeJSONRequestBody{
					IssuerDID: did.String(),
				},
				Params: AuthQRCodeParams{Type: nil},
			},
			expected: expected{
				httpCode:   http.StatusOK,
				qrWithLink: true,
				response: protocol.AuthorizationRequestMessage{
					Body: protocol.AuthorizationRequestMessageBody{
						CallbackURL: "https://testing.env/v1/authentication/callback?sessionID=",
						Reason:      "authentication",
						Scope:       make([]protocol.ZeroKnowledgeProofRequest, 0),
					},
					From: issuerDID.String(),
					Typ:  "application/iden3comm-plain-json",
					Type: "https://iden3-communication.io/authorization/1.0/request",
				},
			},
		},
		{
			name: "should get a qr code with a link as requested",
			request: AuthQRCodeRequestObject{
				Body: &AuthQRCodeJSONRequestBody{
					IssuerDID: did.String(),
				},
				Params: AuthQRCodeParams{
					Type: common.ToPointer(Link),
				},
			},
			expected: expected{
				httpCode:   http.StatusOK,
				qrWithLink: true,
				response: protocol.AuthorizationRequestMessage{
					Body: protocol.AuthorizationRequestMessageBody{
						CallbackURL: "https://testing.env/v1/authentication/callback?sessionID=",
						Reason:      "authentication",
						Scope:       make([]protocol.ZeroKnowledgeProofRequest, 0),
					},
					From: issuerDID.String(),
					Typ:  "application/iden3comm-plain-json",
					Type: "https://iden3-communication.io/authorization/1.0/request",
				},
			},
		},
		{
			name: "should get a RAW qr code as requested",
			request: AuthQRCodeRequestObject{
				Body: &AuthQRCodeJSONRequestBody{
					IssuerDID: did.String(),
				},
				Params: AuthQRCodeParams{
					Type: common.ToPointer(Raw),
				},
			},
			expected: expected{
				httpCode:   http.StatusOK,
				qrWithLink: false,
				response: protocol.AuthorizationRequestMessage{
					Body: protocol.AuthorizationRequestMessageBody{
						CallbackURL: "https://testing.env/v1/authentication/callback?sessionID=",
						Reason:      "authentication",
						Scope:       make([]protocol.ZeroKnowledgeProofRequest, 0),
					},
					From: issuerDID.String(),
					Typ:  "application/iden3comm-plain-json",
					Type: "https://iden3-communication.io/authorization/1.0/request",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			apiURL := "/v1/authentication/qrcode"
			if tc.request.Params.Type != nil {
				apiURL += fmt.Sprintf("?type=%s", *tc.request.Params.Type)
			}
			req, err := http.NewRequest(http.MethodPost, apiURL, tests.JSONBody(t, tc.request.Body))
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var resp AuthQRCode200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				require.NotEmpty(t, resp.QrCodeLink)
				require.NotEmpty(t, resp.SessionID)

				realQR := protocol.AuthorizationRequestMessage{}
				if tc.expected.qrWithLink {
					qrLink := checkQRfetchURL(t, resp.QrCodeLink)

					// Now let's fetch the original QR using the url
					rr := httptest.NewRecorder()
					req, err := http.NewRequest(http.MethodGet, qrLink, nil)
					require.NoError(t, err)
					handler.ServeHTTP(rr, req)
					require.Equal(t, http.StatusOK, rr.Code)
					require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &realQR))
				} else {
					require.NoError(t, json.Unmarshal([]byte(resp.QrCodeLink), &realQR))
				}

				// Let's verify the QR body

				v := tc.expected.response

				assert.Equal(t, v.Typ, realQR.Typ)
				assert.Equal(t, v.Type, realQR.Type)
				assert.Equal(t, v.From, realQR.From)
				assert.Equal(t, v.Body.Scope, realQR.Body.Scope)
				assert.Equal(t, v.Body.Reason, realQR.Body.Reason)
				assert.True(t, strings.Contains(realQR.Body.CallbackURL, v.Body.CallbackURL))
			}
		})
	}
}

func checkQRfetchURL(t *testing.T, qrLink string) string {
	t.Helper()
	qrURL, err := url.Parse(qrLink)
	require.NoError(t, err)
	assert.Equal(t, "iden3comm", qrURL.Scheme)
	vals, err := url.ParseQuery(qrURL.RawQuery)
	require.NoError(t, err)
	val, found := vals["request_uri"]
	require.True(t, found)
	fetchURL, err := url.QueryUnescape(val[0])
	require.NoError(t, err)
	return fetchURL
}
