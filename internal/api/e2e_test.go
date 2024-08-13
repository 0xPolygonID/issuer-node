package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/db/tests"

	core "github.com/iden3/go-iden3-core/v2"
)

func Test_UnsupportedMediaType(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		ETH        = "ETH"
	)
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	type AgentBody struct {
		Id string `json:"id"`
	}

	type AgentFetch struct {
		ID   string    `json:"id"`
		ThID string    `json:"thid"`
		Typ  string    `json:"typ"`
		Type string    `json:"type"`
		Body AgentBody `json:"body"`
		From string    `json:"from"`
		To   string    `json:"to"`
	}

	idData := CreateIdentityRequest{
		DidMetadata: struct {
			AuthBJJCredentialStatus *CreateIdentityRequestDidMetadataAuthBJJCredentialStatus `json:"authBJJCredentialStatus,omitempty"`
			Blockchain              string                                                   `json:"blockchain"`
			Method                  string                                                   `json:"method"`
			Network                 string                                                   `json:"network"`
			Type                    CreateIdentityRequestDidMetadataType                     `json:"type"`
		}{
			Blockchain:              blockchain,
			Method:                  method,
			Network:                 string(core.Amoy),
			Type:                    BJJ,
			AuthBJJCredentialStatus: (*CreateIdentityRequestDidMetadataAuthBJJCredentialStatus)(common.ToPointer("Iden3commRevocationStatusV1.0")),
		},
	}

	credData := CreateClaimRequest{
		CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
		Type:             "KYCAgeCredential",
		CredentialSubject: map[string]any{
			"id":           "did:polygonid:polygon:amoy:2qSJrG9cQXBa9A4J1XMRLVunZTz17hhQokNXq7HNrX",
			"birthday":     19960425,
			"documentType": 2,
		},
		Expiration: common.ToPointer(time.Now().Unix()),
	}

	t.Run("Unsupported media type", func(t *testing.T) {
		rr := httptest.NewRecorder()
		idReq, err := http.NewRequest(http.MethodPost, "/v1/identities", tests.JSONBody(t, idData))
		idReq.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, idReq)
		require.Equal(t, http.StatusCreated, rr.Code)
		var idResponse CreateIdentity201JSONResponse
		json.Unmarshal(rr.Body.Bytes(), &idResponse)

		rr = httptest.NewRecorder()
		credReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/v1/%s/credentials", *idResponse.Identifier), tests.JSONBody(t, credData))
		credReq.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, credReq)
		require.Equal(t, http.StatusCreated, rr.Code)
		var credResponse CreateClaimResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &credResponse))

		agentFetch := AgentFetch{
			ID:   "1924af5a-7d63-4850-addf-0177cdc34786",
			ThID: "1924af5a-7d63-4850-addf-0177cdc34786",
			Typ:  "application/iden3comm-plain-json",
			Type: "https://iden3-communication.io/credentials/1.0/fetch-request",
			Body: AgentBody(credResponse),
			From: "did:polygonid:polygon:amoy:2qT7W95xRrMbtrp63tEysxuSbQLjp9DFQe3FbqAQH8",
			To:   "did:polygonid:polygon:amoy:2qZ29oxye2h1tx9a8eSrs1kuweHBcPXMwHSuusCYyj",
		}

		rr = httptest.NewRecorder()
		agentReq, err := http.NewRequest(http.MethodPost, "/v1/agent", tests.JSONBody(t, agentFetch))
		credReq.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, agentReq)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		var agentResponse Agent400JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &agentResponse))
		require.Equal(t, agentResponse.Message, "unsupported media type 'application/iden3comm-plain-json' for message type 'https://iden3-communication.io/credentials/1.0/fetch-request'")
		agentFetchJson, err := json.Marshal(agentFetch)
		require.NoError(t, err)

		bodyJWZ, err := server.packageManager.Pack("application/iden3-zkp-json", agentFetchJson, nil)
		require.NoError(t, err)
		rr = httptest.NewRecorder()
		agentReqJWZ, err := http.NewRequest(http.MethodPost, "/v1/agent", tests.JSONBody(t, bodyJWZ))
		credReq.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, agentReqJWZ)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		var agentJWZResponse Agent200JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &agentJWZResponse))
	})
}

func Test_RevocationStatusMessageType(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		ETH        = "ETH"
	)
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	type AgentBody struct {
		RevocationNonce float64 `json:"revocationNonce"`
	}

	type AgentRequestStatus struct {
		ID   string    `json:"id"`
		ThID string    `json:"thid"`
		Typ  string    `json:"typ"`
		Type string    `json:"type"`
		Body AgentBody `json:"body"`
		From string    `json:"from"`
		To   string    `json:"to"`
	}

	idData := CreateIdentityRequest{
		DidMetadata: struct {
			AuthBJJCredentialStatus *CreateIdentityRequestDidMetadataAuthBJJCredentialStatus `json:"authBJJCredentialStatus,omitempty"`
			Blockchain              string                                                   `json:"blockchain"`
			Method                  string                                                   `json:"method"`
			Network                 string                                                   `json:"network"`
			Type                    CreateIdentityRequestDidMetadataType                     `json:"type"`
		}{
			Blockchain:              blockchain,
			Method:                  method,
			Network:                 string(core.Amoy),
			Type:                    BJJ,
			AuthBJJCredentialStatus: (*CreateIdentityRequestDidMetadataAuthBJJCredentialStatus)(common.ToPointer("Iden3commRevocationStatusV1.0")),
		},
	}

	credData := CreateClaimRequest{
		CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
		Type:             "KYCAgeCredential",
		CredentialSubject: map[string]any{
			"id":           "did:polygonid:polygon:amoy:2qSJrG9cQXBa9A4J1XMRLVunZTz17hhQokNXq7HNrX",
			"birthday":     19960425,
			"documentType": 2,
		},
		Expiration: common.ToPointer(time.Now().Unix()),
	}

	t.Run("Revocation status media type", func(t *testing.T) {
		rr := httptest.NewRecorder()
		idReq, err := http.NewRequest(http.MethodPost, "/v1/identities", tests.JSONBody(t, idData))
		idReq.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, idReq)
		require.Equal(t, http.StatusCreated, rr.Code)
		var idResponse CreateIdentity201JSONResponse
		json.Unmarshal(rr.Body.Bytes(), &idResponse)

		rr = httptest.NewRecorder()
		credReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/v1/%s/credentials", *idResponse.Identifier), tests.JSONBody(t, credData))
		credReq.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, credReq)
		require.Equal(t, http.StatusCreated, rr.Code)
		var credResponse CreateClaimResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &credResponse))

		rr = httptest.NewRecorder()
		credGetReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/%s/credentials/%s", *idResponse.Identifier, credResponse.Id), tests.JSONBody(t, credData))
		credGetReq.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, credGetReq)
		require.Equal(t, http.StatusOK, rr.Code)
		var credGetResponse GetCredential200JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &credGetResponse))
		status, ok := credGetResponse.CredentialStatus.(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, status["type"], "Iden3commRevocationStatusV1.0")

		agentRequestStatus := AgentRequestStatus{
			ID:   "1924af5a-7d63-4850-addf-0177cdc34786",
			ThID: "1924af5a-7d63-4850-addf-0177cdc34786",
			Typ:  "application/iden3comm-plain-json",
			Type: "https://iden3-communication.io/credentials/1.0/fetch-request",
			Body: AgentBody{
				RevocationNonce: status["revocationNonce"].(float64),
			},
			From: "did:polygonid:polygon:amoy:2qT7W95xRrMbtrp63tEysxuSbQLjp9DFQe3FbqAQH8",
			To:   "did:polygonid:polygon:amoy:2qZ29oxye2h1tx9a8eSrs1kuweHBcPXMwHSuusCYyj",
		}
		rr = httptest.NewRecorder()
		agentReq, err := http.NewRequest(http.MethodPost, "/v1/agent", tests.JSONBody(t, agentRequestStatus))
		credReq.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, agentReq)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		var agentResponse Agent200JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &agentResponse))
	})
}
