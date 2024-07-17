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
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/event"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/credentials/revocation_status"
	"github.com/polygonid/sh-id-platform/pkg/helpers"
	networkPkg "github.com/polygonid/sh-id-platform/pkg/network"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

func TestServer_CreateIdentity(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		ETH        = "ETH"
	)
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()

	reader := helpers.CreateFile(t)

	networkResolver, err := networkPkg.NewResolver(context.Background(), cfg, keyStore, reader)
	require.NoError(t, err)

	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)
	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.ServerUrl, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	accountService := services.NewAccountService(*networkResolver)
	server := NewServer(&cfg, identityService, accountService, claimsService, nil, NewPublisherMock(), NewPackageManagerMock(), *networkResolver, nil)
	handler := getHandler(context.Background(), server)

	type expected struct {
		httpCode int
		message  *string
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		input    CreateIdentityRequest
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
			name: "should create a BJJ identity for amoy network",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: string(core.Amoy), Type: BJJ},
			},
			expected: expected{
				httpCode: 201,
				message:  nil,
			},
		},
		{
			name: "should create a ETH identity for amoy network",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: string(core.Amoy), Type: ETH},
			},
			expected: expected{
				httpCode: 201,
				message:  nil,
			},
		},
		{
			name: "should create a BJJ identity",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: network, Type: BJJ},
			},
			expected: expected{
				httpCode: 201,
				message:  nil,
			},
		},
		{
			name: "should create a ETH identity",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: network, Type: ETH},
			},
			expected: expected{
				httpCode: 201,
				message:  nil,
			},
		},
		{
			name: "should return an error wrong network",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: "mynetwork", Type: BJJ},
			},
			expected: expected{
				httpCode: 400,
				message:  common.ToPointer("error getting reverse hash service settings: rhsSettings not found for polygon:mynetwork"),
			},
		},
		{
			name: "should return an error wrong method",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: "my method", Network: network, Type: BJJ},
			},
			expected: expected{
				httpCode: 400,
				message:  common.ToPointer("cannot create identity: can't add genesis claims to tree: wrong DID Metadata"),
			},
		},
		{
			name: "should return an error wrong blockchain",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: "my blockchain", Method: method, Network: network, Type: BJJ},
			},
			expected: expected{
				httpCode: 400,
				message:  common.ToPointer("error getting reverse hash service settings: rhsSettings not found for my blockchain:amoy"),
			},
		},
		{
			name: "should return an error wrong type",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: "my blockchain", Method: method, Network: network, Type: "a wrong type"},
			},
			expected: expected{
				httpCode: 400,
				message:  common.ToPointer("Type must be BJJ or ETH"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "/v1/identities", tests.JSONBody(t, tc.input))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response CreateIdentityResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				require.NotNil(t, response.Identifier)
				assert.Contains(t, *response.Identifier, tc.input.DidMetadata.Network)
				assert.NotNil(t, response.State.CreatedAt)
				assert.NotNil(t, response.State.ModifiedAt)
				assert.NotNil(t, response.State.State)
				assert.NotNil(t, response.State.Status)
				if tc.input.DidMetadata.Type == BJJ {
					assert.NotNil(t, *response.State.ClaimsTreeRoot)
				}
				if tc.input.DidMetadata.Type == ETH {
					assert.NotNil(t, *response.Address)
				}
			case http.StatusBadRequest:
				var response CreateIdentity400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			}
		})
	}
}

func TestServer_RevokeClaim(t *testing.T) {
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()

	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(context.Background(), cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)

	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)
	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.ServerUrl, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	accountService := services.NewAccountService(*networkResolver)
	server := NewServer(&cfg, identityService, accountService, claimsService, nil, NewPublisherMock(), NewPackageManagerMock(), *networkResolver, nil)

	idStr := "did:polygonid:polygon:mumbai:2qM77fA6NGGWL9QEeb1dv2VA6wz5svcohgv61LZ7wB"
	identity := &domain.Identity{
		Identifier: idStr,
	}
	fixture := tests.NewFixture(storage)
	fixture.CreateIdentity(t, identity)

	idClaim, err := uuid.NewUUID()
	require.NoError(t, err)
	nonce := int64(123)
	revNonce := domain.RevNonceUint64(nonce)
	fixture.CreateClaim(t, &domain.Claim{
		ID:              idClaim,
		Identifier:      &idStr,
		Issuer:          idStr,
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: "",
		Expiration:      0,
		Version:         0,
		RevNonce:        revNonce,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
	})

	query := tests.ExecQueryParams{
		Query: `INSERT INTO identity_mts (identifier, type) VALUES 
                                                    ($1, 0),
                                                    ($1, 1),
                                                    ($1, 2),
                                                    ($1, 3)`,
		Arguments: []interface{}{idStr},
	}

	fixture.ExecQuery(t, query)

	handler := getHandler(context.Background(), server)

	type expected struct {
		response RevokeClaimResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		did      string
		nonce    int64
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:  "No auth header",
			auth:  authWrong,
			did:   idStr,
			nonce: nonce,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:  "should revoke the claim",
			auth:  authOk,
			did:   idStr,
			nonce: nonce,
			expected: expected{
				httpCode: 202,
				response: RevokeClaim202JSONResponse{
					Message: "credential revocation request sent",
				},
			},
		},
		{
			name:  "should get an error wrong nonce",
			auth:  authOk,
			did:   idStr,
			nonce: int64(1231323),
			expected: expected{
				httpCode: 404,
				response: RevokeClaim404JSONResponse{N404JSONResponse{
					Message: "the credential does not exist",
				}},
			},
		},
		{
			name:  "should get an error",
			auth:  authOk,
			did:   "did:polygonid:polygon:mumbai:2qPUUYXa98tQWZKSaRidf2QTDyZicFFxkTWNWjk2HJ",
			nonce: nonce,
			expected: expected{
				httpCode: 500,
				response: RevokeClaim500JSONResponse{N500JSONResponse{
					Message: "error getting merkle trees: not found",
				}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/credentials/revoke/%d", tc.did, tc.nonce)
			req, err := http.NewRequest(http.MethodPost, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch v := tc.expected.response.(type) {
			case RevokeClaim202JSONResponse:
				var response RevokeClaim202JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			case RevokeClaim404JSONResponse:
				var response RevokeClaim404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			case RevokeClaim500JSONResponse:
				var response RevokeClaim500JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			}
		})
	}
}

func TestServer_CreateCredential(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	qrService := services.NewQrStoreService(cachex)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()

	reader := helpers.CreateFile(t)

	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, qrService, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)
	pubSub := pubsub.NewMock()
	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.ServerUrl, pubSub, ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	accountService := services.NewAccountService(*networkResolver)
	server := NewServer(&cfg, identityService, accountService, claimsService, nil, NewPublisherMock(), NewPackageManagerMock(), *networkResolver, nil)
	handler := getHandler(ctx, server)

	iden, err := identityService.Create(ctx, "http://polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did := iden.Identifier

	claimID, err := uuid.NewUUID()
	require.NoError(t, err)

	type expected struct {
		response                    CreateCredentialResponseObject
		httpCode                    int
		createCredentialEventsCount int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		did      string
		body     CreateClaimRequest
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "No auth header",
			did:  did,
			auth: authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Happy path",
			auth: authOk,
			did:  did,
			body: CreateClaimRequest{
				CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"birthday":     19960425,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Happy path with two proofs",
			auth: authOk,
			did:  did,
			body: CreateClaimRequest{
				CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"birthday":     19960425,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
				Proofs: &[]CreateClaimRequestProofs{
					"BJJSignature2021",
					"Iden3SparseMerkleTreeProof",
				},
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Happy path with bjjSignature proof",
			auth: authOk,
			did:  did,
			body: CreateClaimRequest{
				CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"birthday":     19960425,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
				Proofs: &[]CreateClaimRequestProofs{
					"BJJSignature2021",
				},
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Happy path with Iden3SparseMerkleTreeProof proof",
			auth: authOk,
			did:  did,
			body: CreateClaimRequest{
				CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"birthday":     19960425,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
				Proofs: &[]CreateClaimRequestProofs{
					"Iden3SparseMerkleTreeProof",
				},
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 0,
			},
		},
		{
			name: "Happy path with ipfs schema",
			auth: authOk,
			did:  did,
			body: CreateClaimRequest{
				CredentialSchema: "ipfs://QmQVeb5dkz5ekDqBrYVVxBFQZoCbzamnmMUn9B8twCEgDL",
				Type:             "testNewType",
				CredentialSubject: map[string]any{
					"id":             "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"testNewTypeInt": 1234,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Happy path with refresh service",
			auth: authOk,
			did:  did,
			body: CreateClaimRequest{
				CredentialSchema: "ipfs://QmQVeb5dkz5ekDqBrYVVxBFQZoCbzamnmMUn9B8twCEgDL",
				Type:             "testNewType",
				CredentialSubject: map[string]any{
					"id":             "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"testNewTypeInt": 1234,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
				RefreshService: &RefreshService{
					Id:   "http://localhost:8080",
					Type: RefreshServiceType(verifiable.Iden3RefreshService2023),
				},
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Happy path with claim id",
			auth: authOk,
			did:  did,
			body: CreateClaimRequest{
				ClaimID:          common.ToPointer(claimID),
				CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"birthday":     19960425,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
				Proofs: &[]CreateClaimRequestProofs{
					"BJJSignature2021",
					"Iden3SparseMerkleTreeProof",
				},
			},
			expected: expected{
				response: CreateCredential201JSONResponse{
					Id: claimID.String(),
				},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Wrong credential url",
			auth: authOk,
			did:  did,
			body: CreateClaimRequest{
				CredentialSchema: "wrong url",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
					"birthday":     19960424,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
			},
			expected: expected{
				response: CreateCredential400JSONResponse{N400JSONResponse{Message: "malformed url"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Unreachable well formed credential url",
			auth: authOk,
			did:  did,
			body: CreateClaimRequest{
				CredentialSchema: "http://www.wrong.url/cannot/get/the/credential",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
					"birthday":     19960424,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
			},
			expected: expected{
				response: CreateCredential422JSONResponse{N422JSONResponse{Message: "cannot load schema"}},
				httpCode: http.StatusUnprocessableEntity,
			},
		},
		{
			name: "Wrong proof type",
			auth: authOk,
			did:  did,
			body: CreateClaimRequest{
				CredentialSchema: "http://www.wrong.url/cannot/get/the/credential",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
					"birthday":     19960424,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
				Proofs:     &[]CreateClaimRequestProofs{"wrong proof"},
			},
			expected: expected{
				response: CreateCredential400JSONResponse{N400JSONResponse{Message: "unsupported proof type: wrong proof"}},
				httpCode: http.StatusBadRequest,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pubSub.Clear(event.CreateCredentialEvent)
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/credentials", tc.did)

			req, err := http.NewRequest(http.MethodPost, url, tests.JSONBody(t, tc.body))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			assert.Equal(t, tc.expected.createCredentialEventsCount, len(pubSub.AllPublishedEvents(event.CreateCredentialEvent)))

			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response CreateClaimResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				_, err := uuid.Parse(response.Id)
				assert.NoError(t, err)
				if tc.body.ClaimID != nil {
					assert.Equal(t, tc.body.ClaimID.String(), response.Id)
				}
			case http.StatusBadRequest:
				var response CreateClaim400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			case http.StatusUnprocessableEntity:
				var response CreateClaim422JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_GetIdentities(t *testing.T) {
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()

	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.ServerUrl, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	accountService := services.NewAccountService(*networkResolver)
	server := NewServer(&cfg, identityService, accountService, claimsService, nil, NewPublisherMock(), NewPackageManagerMock(), *networkResolver, nil)
	handler := getHandler(context.Background(), server)

	idStr1 := "did:polygonid:polygon:mumbai:2qE1ZT16aqEWhh9mX9aqM2pe2ZwV995dTkReeKwCaQ"
	idStr2 := "did:polygonid:polygon:mumbai:2qMHFTHn2SC3XkBEJrR4eH4Yk8jRGg5bzYYG1ZGECa"
	identity1 := &domain.Identity{
		Identifier: idStr1,
	}
	identity2 := &domain.Identity{
		Identifier: idStr2,
	}
	fixture := tests.NewFixture(storage)
	fixture.CreateIdentity(t, identity1)
	fixture.CreateIdentity(t, identity2)

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
			name: "should return all the entities",
			auth: authOk,
			expected: expected{
				httpCode: 200,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/v1/identities", nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			if tc.expected.httpCode == http.StatusOK {
				var response GetIdentities200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.httpCode, rr.Code)
				assert.True(t, len(response) >= 2)
			}
		})
	}
}

func TestServer_GetCredentialQrCode(t *testing.T) {
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	reader := helpers.CreateFile(t)

	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	idStr := "did:polygonid:polygon:mumbai:2qPrv5Yx8s1qAmEnPym68LfT7gTbASGampiGU7TseL"
	idNoClaims := "did:polygonid:polygon:mumbai:2qGjTUuxZKqKS4Q8UmxHUPw55g15QgEVGnj6Wkq8Vk"
	accountService := services.NewAccountService(*networkResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.ServerUrl, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)

	identity := &domain.Identity{
		Identifier: idStr,
	}

	fixture := tests.NewFixture(storage)
	fixture.CreateIdentity(t, identity)

	claim := fixture.NewClaim(t, identity.Identifier)
	fixture.CreateClaim(t, claim)

	server := NewServer(&cfg, identityService, accountService, claimsService, nil, NewPublisherMock(), NewPackageManagerMock(), *networkResolver, nil)
	handler := getHandler(context.Background(), server)

	type expected struct {
		response GetCredentialQrCodeResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		did      string
		claim    uuid.UUID
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name:  "No auth",
			auth:  authWrong,
			did:   idStr,
			claim: claim.ID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:  "should get an error non existing claimID",
			auth:  authOk,
			did:   idStr,
			claim: uuid.New(),
			expected: expected{
				response: GetCredentialQrCode404JSONResponse{N404JSONResponse{
					Message: "claim not found",
				}},
				httpCode: http.StatusNotFound,
			},
		},
		{
			name:  "should get an error the given did has no entry for claimID",
			auth:  authOk,
			did:   idNoClaims,
			claim: claim.ID,
			expected: expected{
				response: GetCredentialQrCode404JSONResponse{N404JSONResponse{
					Message: "claim not found",
				}},
				httpCode: http.StatusNotFound,
			},
		},
		{
			name:  "should get an error wrong did invalid format",
			auth:  authOk,
			did:   ":polygon:mumbai:2qPUUYXa98tQWZKSaRidf2QTDyZicFFxkTWNWjk2HJ",
			claim: claim.ID,
			expected: expected{
				response: GetCredentialQrCode400JSONResponse{N400JSONResponse{
					Message: "invalid did",
				}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name:  "should get a json QR",
			auth:  authOk,
			did:   idStr,
			claim: claim.ID,
			expected: expected{
				response: GetCredentialQrCode200JSONResponse{},
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/credentials/%s/qrcode", tc.did, tc.claim)
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch v := tc.expected.response.(type) {
			case GetCredentialQrCode200JSONResponse:
				var response GetClaimQrCode200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, string(protocol.CredentialOfferMessageType), response.Type)
				assert.Equal(t, string(packers.MediaTypePlainMessage), response.Typ)
				_, err := uuid.Parse(response.Id)
				assert.NoError(t, err)
				assert.Equal(t, response.Id, response.Thid)
				assert.Equal(t, idStr, response.From)
				assert.Equal(t, claim.OtherIdentifier, response.To)
				assert.Equal(t, cfg.ServerUrl+"v1/agent", response.Body.Url)
				require.Len(t, response.Body.Credentials, 1)
				_, err = uuid.Parse(response.Body.Credentials[0].Id)
				assert.NoError(t, err)
				assert.Equal(t, claim.SchemaType, response.Body.Credentials[0].Description)

			case GetCredentialQrCode400JSONResponse:
				var response GetClaimQrCode400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			case GetCredentialQrCode404JSONResponse:
				var response GetClaimQrCode400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			case GetCredentialQrCode500JSONResponse:
				var response GetClaimQrCode500JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			}
		})
	}
}

func TestServer_GetCredential(t *testing.T) {
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	reader := helpers.CreateFile(t)

	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.ServerUrl, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)

	accountService := services.NewAccountService(*networkResolver)
	server := NewServer(&cfg, identityService, accountService, claimsService, nil, NewPublisherMock(), NewPackageManagerMock(), *networkResolver, nil)

	idStr := "did:polygonid:polygon:mumbai:2qLduMv2z7hnuhzkcTWesCUuJKpRVDEThztM4tsJUj"
	idStrWithoutClaims := "did:polygonid:polygon:mumbai:2qGjTUuxZKqKS4Q8UmxHUPw55g15QgEVGnj6Wkq8Vk"
	identity := &domain.Identity{
		Identifier: idStr,
	}
	fixture := tests.NewFixture(storage)
	fixture.CreateIdentity(t, identity)

	claim := fixture.NewClaim(t, identity.Identifier)
	fixture.CreateClaim(t, claim)

	query := tests.ExecQueryParams{
		Query: `INSERT INTO identity_mts (identifier, type) VALUES 
                                                    ($1, 0),
                                                    ($1, 1),
                                                    ($1, 2),
                                                    ($1, 3)`,
		Arguments: []interface{}{idStr},
	}

	fixture.ExecQuery(t, query)

	handler := getHandler(context.Background(), server)

	type expected struct {
		response GetCredentialResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		did      string
		claimID  uuid.UUID
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			did:  idStr,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:    "should get an error non existing claimID",
			auth:    authOk,
			did:     idStr,
			claimID: uuid.New(),
			expected: expected{
				httpCode: http.StatusNotFound,
				response: GetCredential404JSONResponse{N404JSONResponse{
					Message: "claim not found",
				}},
			},
		},
		{
			name:    "should get an error the given did has no entry for claimID",
			auth:    authOk,
			did:     idStrWithoutClaims,
			claimID: claim.ID,
			expected: expected{
				httpCode: http.StatusNotFound,
				response: GetCredential404JSONResponse{N404JSONResponse{
					Message: "claim not found",
				}},
			},
		},
		{
			name:    "should get an error wrong did invalid format",
			auth:    authOk,
			did:     ":polygon:mumbai:2qPUUYXa98tQWZKSaRidf2QTDyZicFFxkTWNWjk2HJ",
			claimID: claim.ID,
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: GetCredential400JSONResponse{N400JSONResponse{
					Message: "invalid did",
				}},
			},
		},
		{
			name:    "should get the claim",
			auth:    authOk,
			did:     idStr,
			claimID: claim.ID,
			expected: expected{
				httpCode: http.StatusOK,
				response: GetCredential200JSONResponse{
					Context: []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
					CredentialSchema: CredentialSchema{
						"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
						"JsonSchemaValidator2018",
					},
					CredentialStatus: verifiable.CredentialStatus{
						ID:              fmt.Sprintf("http://localhost/v1/%s/credentials/revocation/status/%d", idStr, claim.RevNonce),
						Type:            "SparseMerkleTreeProof",
						RevocationNonce: uint64(claim.RevNonce),
					},
					CredentialSubject: map[string]interface{}{
						"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
						"birthday":     float64(19960424),
						"documentType": float64(2),
						"type":         "KYCAgeCredential",
					},
					Id:           fmt.Sprintf("http://localhost/api/v1/credentials/%s", claim.ID),
					IssuanceDate: common.ToPointer(TimeUTC(time.Now())),
					Issuer:       idStr,
					Type:         []string{"VerifiableCredential", "KYCAgeCredential"},
					RefreshService: &RefreshService{
						Id:   "https://refresh-service.xyz",
						Type: RefreshServiceType(verifiable.Iden3RefreshService2023),
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/credentials/%s", tc.did, tc.claimID.String())
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch v := tc.expected.response.(type) {
			case GetCredential200JSONResponse:
				var response GetClaimResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				validateClaim(t, response, GetClaimResponse(v))

			case GetCredential400JSONResponse:
				var response GetClaim404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			case GetCredential404JSONResponse:
				var response GetClaim404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			case GetCredential500JSONResponse:
				var response GetClaim500JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			}
		})
	}
}

func TestServer_GetCredentials(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	connectionsRepository := repositories.NewConnections()
	reader := helpers.CreateFile(t)

	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.ServerUrl, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)

	fixture := tests.NewFixture(storage)

	accountService := services.NewAccountService(*networkResolver)
	server := NewServer(&cfg, identityService, accountService, claimsService, nil, NewPublisherMock(), NewPackageManagerMock(), *networkResolver, nil)
	identityMultipleClaims, err := server.identityService.Create(ctx, "https://localhost.com", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	claim := fixture.NewClaim(t, identityMultipleClaims.Identifier)
	_ = fixture.CreateClaim(t, claim)

	emptyIdentityStr := "did:polygonid:polygon:mumbai:2qLQGgjpP5Yq7r7jbRrQZbWy8ikADvxamSLB7CqR4F"

	handler := getHandler(context.Background(), server)

	type expected struct {
		response GetCredentialsResponseObject
		len      int
		httpCode int
	}

	type filter struct {
		schemaHash *string
		schemaType *string
		subject    *string
		revoked    *string
		self       *string
		queryField *string
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		did      string
		expected expected
		filter   filter
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			did:  ":polygon:mumbai:2qPUUYXa98tQWZKSaRidf2QTDyZicFFxkTWNWjk2HJ",
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "should get an error wrong did invalid format",
			auth: authOk,
			did:  ":polygon:mumbai:2qPUUYXa98tQWZKSaRidf2QTDyZicFFxkTWNWjk2HJ",
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: GetCredentials400JSONResponse{N400JSONResponse{
					Message: "invalid did",
				}},
			},
		},
		{
			name: "should get an error self and subject filter cannot be used together",
			auth: authOk,
			did:  identityMultipleClaims.Identifier,
			filter: filter{
				self:    common.ToPointer("true"),
				subject: common.ToPointer("some subject"),
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: GetCredentials400JSONResponse{N400JSONResponse{"self and subject filter cannot be used together"}},
			},
		},
		{
			name: "should get 0 claims",
			auth: authOk,
			did:  emptyIdentityStr,
			expected: expected{
				httpCode: http.StatusOK,
				len:      0,
				response: GetCredentials200JSONResponse{},
			},
		},
		{
			name: "should get the default claim plus another one that has been created",
			auth: authOk,
			did:  identityMultipleClaims.Identifier,
			expected: expected{
				httpCode: http.StatusOK,
				len:      1,
				response: GetCredentials200JSONResponse{
					GetClaimResponse{
						Context: []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
						CredentialSchema: CredentialSchema{
							"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
							"JsonSchemaValidator2018",
						},
						CredentialStatus: verifiable.CredentialStatus{
							ID:              fmt.Sprintf("http://localhost/v1/%s/credentials/revocation/status/%d", identityMultipleClaims.Identifier, claim.RevNonce),
							Type:            "SparseMerkleTreeProof",
							RevocationNonce: uint64(claim.RevNonce),
						},
						CredentialSubject: map[string]interface{}{
							"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
							"birthday":     float64(19960424),
							"documentType": float64(2),
							"type":         "KYCAgeCredential",
						},
						Id:           fmt.Sprintf("http://localhost/api/v1/credentials/%s", claim.ID),
						IssuanceDate: common.ToPointer(TimeUTC(time.Now())),
						Issuer:       identityMultipleClaims.Identifier,
						Type:         []string{"VerifiableCredential", "KYCAgeCredential"},
						RefreshService: &RefreshService{
							Id:   "https://refresh-service.xyz",
							Type: RefreshServiceType(verifiable.Iden3RefreshService2023),
						},
					},
				},
			},
		},
		{
			name: "should get 1 credential with the given schemaHash filter",
			auth: authOk,
			did:  identityMultipleClaims.Identifier,
			filter: filter{
				schemaHash: common.ToPointer(claim.SchemaHash),
			},
			expected: expected{
				httpCode: http.StatusOK,
				len:      1,
				response: GetCredentials200JSONResponse{
					GetClaimResponse{
						Context: []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
						CredentialSchema: CredentialSchema{
							"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
							"JsonSchemaValidator2018",
						},
						CredentialStatus: verifiable.CredentialStatus{
							ID:              fmt.Sprintf("http://localhost/v1/%s/credentials/revocation/status/%d", identityMultipleClaims.Identifier, claim.RevNonce),
							Type:            "SparseMerkleTreeProof",
							RevocationNonce: uint64(claim.RevNonce),
						},
						CredentialSubject: map[string]interface{}{
							"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
							"birthday":     float64(19960424),
							"documentType": float64(2),
							"type":         "KYCAgeCredential",
						},
						Id:           fmt.Sprintf("http://localhost/api/v1/credentials/%s", claim.ID),
						IssuanceDate: common.ToPointer(TimeUTC(time.Now())),
						Issuer:       identityMultipleClaims.Identifier,
						Type:         []string{"VerifiableCredential", "KYCAgeCredential"},
						RefreshService: &RefreshService{
							Id:   "https://refresh-service.xyz",
							Type: RefreshServiceType(verifiable.Iden3RefreshService2023),
						},
					},
				},
			},
		},
		{
			name: "should get 1 credential with multiple filters",
			auth: authOk,
			did:  identityMultipleClaims.Identifier,
			filter: filter{
				schemaHash: common.ToPointer(claim.SchemaHash),
				revoked:    common.ToPointer("false"),
			},
			expected: expected{
				httpCode: http.StatusOK,
				len:      1,
				response: GetCredentials200JSONResponse{
					GetClaimResponse{
						Context: []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
						CredentialSchema: CredentialSchema{
							"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
							"JsonSchemaValidator2018",
						},
						CredentialStatus: verifiable.CredentialStatus{
							ID:              fmt.Sprintf("http://localhost/v1/%s/credentials/revocation/status/%d", identityMultipleClaims.Identifier, claim.RevNonce),
							Type:            "SparseMerkleTreeProof",
							RevocationNonce: uint64(claim.RevNonce),
						},
						CredentialSubject: map[string]interface{}{
							"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
							"birthday":     float64(19960424),
							"documentType": float64(2),
							"type":         "KYCAgeCredential",
						},
						Id:           fmt.Sprintf("http://localhost/api/v1/credentials/%s", claim.ID),
						IssuanceDate: common.ToPointer(TimeUTC(time.Now())),
						Issuer:       identityMultipleClaims.Identifier,
						Type:         []string{"VerifiableCredential", "KYCAgeCredential"},
						RefreshService: &RefreshService{
							Id:   "https://refresh-service.xyz",
							Type: RefreshServiceType(verifiable.Iden3RefreshService2023),
						},
					},
				},
			},
		},
		{
			name: "should get 0 revoked credentials",
			auth: authOk,
			did:  identityMultipleClaims.Identifier,
			filter: filter{
				revoked: common.ToPointer("true"),
			},
			expected: expected{
				httpCode: http.StatusOK,
				len:      0,
				response: GetCredentials200JSONResponse{},
			},
		},
		{
			name: "should get two non revoked credentials",
			auth: authOk,
			did:  identityMultipleClaims.Identifier,
			filter: filter{
				revoked: common.ToPointer("false"),
			},
			expected: expected{
				httpCode: http.StatusOK,
				len:      1,
				response: GetCredentials200JSONResponse{
					GetClaimResponse{
						Context: []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
						CredentialSchema: CredentialSchema{
							"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
							"JsonSchemaValidator2018",
						},
						CredentialStatus: verifiable.CredentialStatus{
							ID:              fmt.Sprintf("http://localhost/v1/%s/credentials/revocation/status/%d", identityMultipleClaims.Identifier, claim.RevNonce),
							Type:            "SparseMerkleTreeProof",
							RevocationNonce: uint64(claim.RevNonce),
						},
						CredentialSubject: map[string]interface{}{
							"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
							"birthday":     float64(19960424),
							"documentType": float64(2),
							"type":         "KYCAgeCredential",
						},
						Id:           fmt.Sprintf("http://localhost/api/v1/credentials/%s", claim.ID),
						IssuanceDate: common.ToPointer(TimeUTC(time.Now())),
						Issuer:       identityMultipleClaims.Identifier,
						Type:         []string{"VerifiableCredential", "KYCAgeCredential"},
						RefreshService: &RefreshService{
							Id:   "https://refresh-service.xyz",
							Type: RefreshServiceType(verifiable.Iden3RefreshService2023),
						},
					},
				},
			},
		},
		{
			name: "should get one credential with the given schemaType filter",
			auth: authOk,
			did:  identityMultipleClaims.Identifier,
			filter: filter{
				schemaType: common.ToPointer("AuthBJJCredential"),
			},
			expected: expected{
				httpCode: http.StatusOK,
				len:      1,
				response: GetCredentials200JSONResponse{
					GetClaimResponse{
						Context: []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
						CredentialSchema: CredentialSchema{
							"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
							"JsonSchemaValidator2018",
						},
						CredentialStatus: verifiable.CredentialStatus{
							ID:              fmt.Sprintf("http://localhost/v1/%s/credentials/revocation/status/%d", identityMultipleClaims.Identifier, claim.RevNonce),
							Type:            "SparseMerkleTreeProof",
							RevocationNonce: uint64(claim.RevNonce),
						},
						CredentialSubject: map[string]interface{}{
							"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
							"birthday":     float64(19960424),
							"documentType": float64(2),
							"type":         "KYCAgeCredential",
						},
						Id:           fmt.Sprintf("http://localhost/api/v1/credentials/%s", claim.ID),
						IssuanceDate: common.ToPointer(TimeUTC(time.Now())),
						Issuer:       identityMultipleClaims.Identifier,
						Type:         []string{"VerifiableCredential", "KYCAgeCredential"},
						RefreshService: &RefreshService{
							Id:   "https://refresh-service.xyz",
							Type: RefreshServiceType(verifiable.Iden3RefreshService2023),
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			tURL := createGetClaimsURL(tc.did, tc.filter.schemaHash, tc.filter.schemaType, tc.filter.subject, tc.filter.revoked, tc.filter.self, tc.filter.queryField)
			req, err := http.NewRequest("GET", tURL, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch v := tc.expected.response.(type) {
			case GetCredentials200JSONResponse:
				var response GetClaims200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.len, len(response))
				for i := range response {
					validateClaim(t, response[i], v[i])
				}
			case GetCredentials400JSONResponse:
				var response GetClaims400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			case GetCredentials500JSONResponse:
				var response GetClaims500JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			}
		})
	}
}

func TestServer_GetRevocationStatus(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()

	reader := helpers.CreateFile(t)

	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	identity, err := identityService.Create(ctx, "http://localhost:3001", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	assert.NoError(t, err)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.ServerUrl, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	accountService := services.NewAccountService(*networkResolver)
	server := NewServer(&cfg, identityService, accountService, claimsService, nil, NewPublisherMock(), NewPackageManagerMock(), *networkResolver, nil)
	handler := getHandler(context.Background(), server)

	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	did, _ := w3c.ParseDID(identity.Identifier)
	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"

	merklizedRootPosition := "value"
	claimRequestProofs := ports.ClaimRequestProofs{
		BJJSignatureProof2021:      true,
		Iden3SparseMerkleTreeProof: true,
	}
	claim, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schema, credentialSubject, common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition, claimRequestProofs, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	assert.NoError(t, err)

	type expected struct {
		httpCode int
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		nonce    int64
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:  "should get revocation status",
			auth:  authOk,
			nonce: int64(claim.RevNonce),
			expected: expected{
				httpCode: http.StatusOK,
			},
		},

		{
			name:  "should get revocation status wrong nonce",
			auth:  authOk,
			nonce: 123456,
			expected: expected{
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/credentials/revocation/status/%d", identity.Identifier, tc.nonce)
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			if tc.expected.httpCode == http.StatusOK {
				var response GetRevocationStatus200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.NotNil(t, response.Issuer.ClaimsTreeRoot)
				assert.NotNil(t, response.Issuer.State)
				assert.NotNil(t, response.Mtp.Existence)
				assert.NotNil(t, response.Mtp.Siblings)
			}
		})
	}
}

func validateClaim(t *testing.T, resp, tc GetClaimResponse) {
	t.Helper()
	var responseCredentialStatus verifiable.CredentialStatus

	credentialSubjectTypes := []string{"AuthBJJCredential", "KYCAgeCredential"}

	type credentialKYCSubject struct {
		Id           string `json:"id"`
		Birthday     uint64 `json:"birthday"`
		DocumentType uint64 `json:"documentType"`
		Type         string `json:"type"`
	}

	type credentialBJJSubject struct {
		Type string `json:"type"`
		X    string `json:"x"`
		Y    string `json:"y"`
	}

	assert.Equal(t, resp.Id, tc.Id)
	assert.Equal(t, len(resp.Context), len(tc.Context))
	assert.EqualValues(t, resp.Context, tc.Context)
	assert.EqualValues(t, resp.CredentialSchema, tc.CredentialSchema)
	assert.InDelta(t, time.Time(*resp.IssuanceDate).UnixMilli(), time.Time(*tc.IssuanceDate).UnixMilli(), 1000)
	assert.Equal(t, resp.Type, tc.Type)
	assert.Equal(t, resp.ExpirationDate, tc.ExpirationDate)
	assert.Equal(t, resp.Issuer, tc.Issuer)
	assert.Equal(t, resp.RefreshService, tc.RefreshService)
	credentialSubjectType, ok := tc.CredentialSubject["type"]
	require.True(t, ok)
	assert.Contains(t, credentialSubjectTypes, credentialSubjectType)
	if credentialSubjectType == "AuthBJJCredential" {
		var responseCredentialSubject, tcCredentialSubject credentialBJJSubject
		assert.NoError(t, mapstructure.Decode(resp.CredentialSubject, &responseCredentialSubject))
		assert.NoError(t, mapstructure.Decode(tc.CredentialSubject, &tcCredentialSubject))
		assert.EqualValues(t, responseCredentialSubject, tcCredentialSubject)
	} else {
		var responseCredentialSubject, tcCredentialSubject credentialKYCSubject
		assert.NoError(t, mapstructure.Decode(resp.CredentialSubject, &responseCredentialSubject))
		assert.NoError(t, mapstructure.Decode(tc.CredentialSubject, &tcCredentialSubject))
		assert.EqualValues(t, responseCredentialSubject, tcCredentialSubject)
	}

	assert.NoError(t, mapstructure.Decode(resp.CredentialStatus, &responseCredentialStatus))
	responseCredentialStatus.ID = strings.Replace(responseCredentialStatus.ID, "%3A", ":", -1)
	credentialStatusTC, ok := tc.CredentialStatus.(verifiable.CredentialStatus)
	require.True(t, ok)
	assert.EqualValues(t, responseCredentialStatus, credentialStatusTC)
}

func createGetClaimsURL(did string, schemaHash *string, schemaType *string, subject *string, revoked *string, self *string, queryField *string) string {
	tURL := &url.URL{Path: fmt.Sprintf("/v1/%s/credentials", did)}
	q := tURL.Query()

	if self != nil {
		q.Add("self", *self)
	}

	if schemaHash != nil {
		q.Add("schemaHash", *schemaHash)
	}

	if schemaType != nil {
		q.Add("schemaType", *schemaType)
	}

	if revoked != nil {
		q.Add("revoked", *revoked)
	}

	if subject != nil {
		q.Add("subject", *subject)
	}

	if queryField != nil {
		q.Add("queryField", *queryField)
	}

	tURL.RawQuery = q.Encode()
	return tURL.String()
}
