package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

func TestServer_GetQrFromStore(t *testing.T) {
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

	_, err = server.Services.links.CreateQRCode(ctx, *did, link.ID, "https://privado.id")
	require.NoError(t, err)

	linkExpired, err := server.Services.links.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, common.ToPointer(tomorrow), true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	require.NoError(t, err)

	linkMaxIssuance, err := server.Services.links.Save(ctx, *did, common.ToPointer(0), &yesterday, importedSchema.ID, common.ToPointer(tomorrow), true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		httpCode int
	}

	type testConfig struct {
		name     string
		request  GetQrFromStoreRequestObject
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "happy path",
			request: GetQrFromStoreRequestObject{
				Params: GetQrFromStoreParams{
					Id:     common.ToPointer(link.ID),
					Issuer: common.ToPointer(iden.Identifier),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
			},
		},
		{
			name: "link expired",
			request: GetQrFromStoreRequestObject{
				Params: GetQrFromStoreParams{
					Id:     common.ToPointer(linkExpired.ID),
					Issuer: common.ToPointer(iden.Identifier),
				},
			},
			expected: expected{
				httpCode: http.StatusGone,
			},
		},
		{
			name: "link Max Issuance reached",
			request: GetQrFromStoreRequestObject{
				Params: GetQrFromStoreParams{
					Id:     common.ToPointer(linkMaxIssuance.ID),
					Issuer: common.ToPointer(iden.Identifier),
				},
			},
			expected: expected{
				httpCode: http.StatusGone,
			},
		},
		{
			name: "wrong did",
			request: GetQrFromStoreRequestObject{
				Params: GetQrFromStoreParams{
					Id:     common.ToPointer(linkMaxIssuance.ID),
					Issuer: common.ToPointer("123"),
				},
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "link wrong id",
			request: GetQrFromStoreRequestObject{
				Params: GetQrFromStoreParams{
					Id:     common.ToPointer(uuid.New()),
					Issuer: common.ToPointer(iden.Identifier),
				},
			},
			expected: expected{
				httpCode: http.StatusNotFound,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/public/v2/qr-store?id=%s&issuer=%s", tc.request.Params.Id, *tc.request.Params.Issuer)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)
			if tc.expected.httpCode == http.StatusOK {
				var response protocol.AuthorizationRequestMessage
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.NotNil(t, response)
				assert.NotNil(t, response.Body)
				assert.Equal(t, iden.Identifier, response.From)
				assert.NotNil(t, response.ID)
				assert.Equal(t, iden3comm.ProtocolMessage("https://iden3-communication.io/authorization/1.0/request"), response.Type)
				assert.Equal(t, iden3comm.MediaType("application/iden3comm-plain-json"), response.Typ)
				assert.Equal(t, "", response.To)
				assert.NotNil(t, response.Body.Scope)
				assert.NotNil(t, response.Body.Message)
				assert.NotNil(t, response.Body.CallbackURL)
			}
		})
	}
}
