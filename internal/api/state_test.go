package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

	didWithoutTxs, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	didWithTxs, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	didWithTxsW3c, err := w3c.ParseDID(didWithTxs.Identifier)
	require.NoError(t, err)

	query1 := fmt.Sprintf(`INSERT INTO identity_states (identifier, state, root_of_roots, revocation_tree_root, claims_tree_root, block_timestamp, block_number, tx_id, previous_state, status, modified_at, created_at) VALUES ('%s', '246819d3881a911bda3e5b22365e1a62446d3d2a5bdd1f633d325f0c3b7ff50c', null, null, '48f746059439e074c9841198f487f8c531cdd08474fbd593869325b6bc40852f', null, null, null, 'e894bf8cdab1b84d0f2d0ce92dcf67f914ddf110f0ca4dd2d29464b6de2a8414', 'confirmed', '2024-08-06 17:48:54.084857 +00:00', '2024-08-01 12:24:30.694229 +00:00');`, didWithTxs.Identifier)
	query2 := fmt.Sprintf(`INSERT INTO identity_states (identifier, state, root_of_roots, revocation_tree_root, claims_tree_root, block_timestamp, block_number, tx_id, previous_state, status, modified_at, created_at) VALUES ('%s', '4ec17c0090bbdc42ec8f60843c73a363b69fe838aab4e14914f3ef2f2ba7961c', null, null, 'eae5c2d656490c6768216607217a31a4a7eb73a146523cd6c671e69aceee470c', null, null, null, 'e894bf8cdab1b84d0f2d0ce92dcf67f914ddf110f0ca4dd2d29464b6de2a8414', 'confirmed', '2024-08-06 17:48:54.084857 +00:00', '2024-08-01 13:43:31.150943 +00:00');`, didWithTxs.Identifier)
	query3 := fmt.Sprintf(`INSERT INTO identity_states (identifier, state, root_of_roots, revocation_tree_root, claims_tree_root, block_timestamp, block_number, tx_id, previous_state, status, modified_at, created_at) VALUES ('%s', 'e894bf8cdab1b84d0f2d0ce92dcf67f914ddf110f0ca4dd2d29464b6de2a8414', null, null, '1ffd0b9308d50d522a1de9baac4044ef37da67e57bc1145d91c9a95eec97ba1c', null, null, null, 'e894bf8cdab1b84d0f2d0ce92dcf67f914ddf110f0ca4dd2d29464b6de2a8414', 'confirmed', '2024-08-06 17:48:54.084857 +00:00', '2024-08-06 11:31:06.765453 +00:00');`, didWithTxs.Identifier)
	query4 := fmt.Sprintf(`INSERT INTO identity_states (identifier, state, root_of_roots, revocation_tree_root, claims_tree_root, block_timestamp, block_number, tx_id, previous_state, status, modified_at, created_at) VALUES ('%s', '6a62b311d83a406ae21d4e21c84fe44876aa069c17706d90637bf038e8ad0b1c', '2e169a2702182f34483fe6043633985aeb966a973752366af381da5e3307f32c', '0000000000000000000000000000000000000000000000000000000000000000', 'be1617227a4aa5c5b7c50f7bb2fd602d9199c85eba497f1996fc24b99c2e1f2e', 1722947305, 10391855, '0x3b8545366556f55be186f92097f203f85a917b93cd68f431ec292c74df4e2472', 'e894bf8cdab1b84d0f2d0ce92dcf67f914ddf110f0ca4dd2d29464b6de2a8414', 'confirmed', '2024-08-06 17:48:54.084857 +00:00', '2024-08-06 12:28:16.495701 +00:00');`, didWithTxs.Identifier)
	query5 := fmt.Sprintf(`INSERT INTO identity_states (identifier, state, root_of_roots, revocation_tree_root, claims_tree_root, block_timestamp, block_number, tx_id, previous_state, status, modified_at, created_at) VALUES ('%s', '45bb0f2ce3b48b79e8a0b4750400892551afbdbe895d754341e07fb93d80ec26', null, null, '7e2a2c5919c9a071615139d542a7483f1be3fb5c1e240fb1ee9ba5314b00a211', null, null, '0x3b8545366556f55be186f92097f203f85a917b93cd68f431ec292c74df4e2500', 'e894bf8cdab1b84d0f2d0ce92dcf67f914ddf110f0ca4dd2d29464b6de2a8414', 'confirmed', '2024-08-06 17:48:54.084857 +00:00', '2024-08-06 13:34:04.385544 +00:00');`, didWithTxs.Identifier)
	_, err = storage.Pgx.Exec(ctx, query1)
	assert.NoError(t, err)
	_, err = storage.Pgx.Exec(ctx, query2)
	assert.NoError(t, err)
	_, err = storage.Pgx.Exec(ctx, query3)
	assert.NoError(t, err)
	_, err = storage.Pgx.Exec(ctx, query4)
	assert.NoError(t, err)
	_, err = storage.Pgx.Exec(ctx, query5)
	assert.NoError(t, err)

	date1, err := time.Parse(time.RFC3339, "2024-08-06T17:48:54.084857Z")
	assert.NoError(t, err)

	date2, err := time.Parse(time.RFC3339, "2024-08-06T17:48:54.084857Z")
	assert.NoError(t, err)

	type expected struct {
		response GetStateTransactions200JSONResponse
		httpCode int
	}

	type testConfig struct {
		name     string
		did      *w3c.DID
		auth     func() (string, string)
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			did:  didWithoutTxs,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "No states to process",
			auth: authOk,
			did:  didWithoutTxs,
			expected: expected{
				response: GetStateTransactions200JSONResponse{},
				httpCode: http.StatusOK,
			},
		},
		{
			name: "happy path",
			auth: authOk,
			did:  didWithTxsW3c,
			expected: expected{
				response: GetStateTransactions200JSONResponse{
					StateTransaction{
						PublishDate: TimeUTC(date1.UTC()),
						State:       "6a62b311d83a406ae21d4e21c84fe44876aa069c17706d90637bf038e8ad0b1c",
						Status:      "published",
						TxID:        "0x3b8545366556f55be186f92097f203f85a917b93cd68f431ec292c74df4e2472",
					},
					StateTransaction{
						PublishDate: TimeUTC(date2.UTC()),
						State:       "45bb0f2ce3b48b79e8a0b4750400892551afbdbe895d754341e07fb93d80ec26",
						Status:      "published",
						TxID:        "0x3b8545366556f55be186f92097f203f85a917b93cd68f431ec292c74df4e2500",
					},
				},
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/state/transactions?page=2&max_results=3", tc.did)
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
				if len(tc.expected.response) > 0 {
					assert.Equal(t, tc.expected.response[0].PublishDate, response[0].PublishDate)
					assert.Equal(t, tc.expected.response[0].State, response[0].State)
					assert.Equal(t, tc.expected.response[0].Status, response[0].Status)
					assert.Equal(t, tc.expected.response[0].TxID, response[0].TxID)
					assert.Equal(t, tc.expected.response[1].PublishDate, response[1].PublishDate)
					assert.Equal(t, tc.expected.response[1].State, response[1].State)
					assert.Equal(t, tc.expected.response[1].Status, response[1].Status)
					assert.Equal(t, tc.expected.response[1].TxID, response[1].TxID)
				}
			}
		})
	}
}
