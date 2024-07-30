package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

func TestServer_GetStateStatus(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()

	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"

	serverWithSignatureClaim := newTestServer(t, nil)

	idenWithSignatureClaim, err := serverWithSignatureClaim.Services.identity.Create(ctx, "amoy-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	didSignatureClaim, err := w3c.ParseDID(idenWithSignatureClaim.Identifier)
	require.NoError(t, err)
	_, err = serverWithSignatureClaim.Services.credentials.Save(ctx, ports.NewCreateClaimRequest(didSignatureClaim, nil, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: false}, nil, true, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	handlerWithSignatureClaim := getHandler(ctx, serverWithSignatureClaim)

	serverWithMTPClaim := newTestServer(t, nil)

	idenWithMTPClaim, err := serverWithMTPClaim.Services.identity.Create(ctx, "amoy-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	didWithMTPClaim, err := w3c.ParseDID(idenWithMTPClaim.Identifier)
	require.NoError(t, err)
	_, err = serverWithMTPClaim.Services.credentials.Save(ctx, ports.NewCreateClaimRequest(didWithMTPClaim, nil, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: true}, nil, true, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)
	handlerWithMTPClaim := getHandler(ctx, serverWithMTPClaim)

	serverWithRevokedClaim := newTestServer(t, nil)
	idenWithRevokedClaim, err := serverWithRevokedClaim.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	didWithRevokedClaim, err := w3c.ParseDID(idenWithRevokedClaim.Identifier)
	require.NoError(t, err)

	cfgWithRevokedClaim := &config.Configuration{
		APIUI: config.APIUI{
			IssuerDID: *didWithRevokedClaim,
		},
	}

	cred, err := serverWithRevokedClaim.Services.credentials.Save(ctx, ports.NewCreateClaimRequest(didWithRevokedClaim, nil, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: false}, nil, true, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)
	require.NoError(t, serverWithRevokedClaim.Services.credentials.Revoke(ctx, cfgWithRevokedClaim.APIUI.IssuerDID, uint64(cred.RevNonce), "not valid"))
	handlerWithRevokedClaim := getHandler(ctx, serverWithRevokedClaim)

	type expected struct {
		response GetStateStatus200JSONResponse
		httpCode int
	}

	type testConfig struct {
		name      string
		handler   http.Handler
		issuerDID *w3c.DID
		auth      func() (string, string)
		expected  expected
	}
	for _, tc := range []testConfig{
		{
			name:      "No auth header",
			handler:   handlerWithSignatureClaim,
			issuerDID: didSignatureClaim,
			auth:      authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:      "No states to process",
			auth:      authOk,
			handler:   handlerWithSignatureClaim,
			issuerDID: didSignatureClaim,
			expected: expected{
				response: GetStateStatus200JSONResponse{PendingActions: false},
				httpCode: http.StatusOK,
			},
		},
		{
			name:      "New state to process because there is a new credential with mtp proof",
			handler:   handlerWithMTPClaim,
			issuerDID: didWithMTPClaim,
			auth:      authOk,
			expected: expected{
				response: GetStateStatus200JSONResponse{PendingActions: true},
				httpCode: http.StatusOK,
			},
		},
		{
			name:      "New state to process because there is a revoked credential",
			handler:   handlerWithRevokedClaim,
			issuerDID: didWithRevokedClaim,
			auth:      authOk,
			expected: expected{
				response: GetStateStatus200JSONResponse{PendingActions: true},
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/state/status", tc.issuerDID.String())

			req, err := http.NewRequest(http.MethodGet, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			tc.handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetStateStatus200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.response.PendingActions, response.PendingActions)
			}
		})
	}
}

func TestServer_GetStateTransactions(t *testing.T) {
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
		response GetStateTransactions200JSONResponse
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
			name: "No states to process",
			auth: authOk,
			expected: expected{
				response: GetStateTransactions200JSONResponse{},
				httpCode: http.StatusOK,
			},
		},
		{
			name: "No state transactions after revoking/creating credentials",
			auth: authOk,
			expected: expected{
				response: GetStateTransactions200JSONResponse{},
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/state/transactions", did)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetStateTransactions200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, len(tc.expected.response), len(response))
			}
		})
	}
}
