package repositories

import (
	"context"
	"fmt"
	// "io"
	"log"
	// "net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/iden3/go-circuits/v2"
	auth "github.com/iden3/go-iden3-auth/v2"
	"github.com/iden3/go-iden3-auth/v2/loaders"
	"github.com/iden3/go-iden3-auth/v2/pubsignals"
	"github.com/iden3/go-iden3-auth/v2/state"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
)


type verifier struct{}

func NewVerifier() ports.VerifierRepository {
	return &verifier{}
}
var requestMap = make(map[string]interface{})
var sessionID = 0
func (v *verifier) GetAuthRequest(ctx context.Context,schemaType string,schemaURL string,credSubject map[string]interface{})(protocol.AuthorizationRequestMessage,error){
	// Audience is verifier id
	rURL := "localhost:3002"
	sessionID++ 
	CallbackURL := "/api/callback"
	Audience := "did:polygonid:polygon:mumbai:2qDyy1kEo2AYcP3RT4XGea7BtxsY285szg6yP9SPrs"

	uri := fmt.Sprintf("%s%s?sessionId=%s", rURL, CallbackURL, strconv.Itoa(sessionID))

	// Generate request for basic authentication
	var request protocol.AuthorizationRequestMessage = auth.CreateAuthorizationRequest("test flow", Audience, uri)

	request.ID = uuid.New().String()
	request.ThreadID = request.ID
	// Add request for a specific proof
	var mtpProofRequest protocol.ZeroKnowledgeProofRequest
	mtpProofRequest.ID = 1
	mtpProofRequest.CircuitID = string(circuits.AtomicQuerySigV2CircuitID)
	mtpProofRequest.Query = map[string]interface{}{
		"allowedIssuers": []string{"*"},
		"credentialSubject":credSubject,
		"context": schemaURL,
		"type":    schemaType,
	}
	request.Body.Scope = append(request.Body.Scope, mtpProofRequest)

	// Store auth request in map associated with session ID
	requestMap[strconv.Itoa(sessionID)] = request

	// print request
	fmt.Println("Request",request)
	return request,nil;
}

// // Callback works with sign-in callbacks
func (v *verifier) Callback(ctx context.Context,sessionId string,tokenBytes []byte)(messageBytes []byte, err error) {

	// Get session ID from request
	// sessionID := r.URL.Query().Get("sessionId")

	// // get JWZ token params from the post request
	// tokenBytes, _ := io.ReadAll(r.Body)

	// Add Polygon Mumbai RPC node endpoint - needed to read on-chain state
	ethURL := "https://polygon-mumbai.g.alchemy.com/v2/YSO_NsiNTjiA-6thPC2RXS9NoBbjjDKC"

	// Add identity state contract address
	contractAddress := "0x134B1BE34911E39A8397ec6289782989729807a4"

	resolverPrefix := "polygon:mumbai"

	// Locate the directory that contains circuit's verification keys
	keyDIR := "../keys"

	// fetch authRequest from sessionID
	authRequest := requestMap[sessionId]

	// print authRequest
	fmt.Println(authRequest)

	// load the verification key
	var verificationKeyloader = &loaders.FSKeyLoader{Dir: keyDIR}
	resolver := state.ETHResolver{
		RPCUrl:          ethURL,
		ContractAddress: common.HexToAddress(contractAddress),
	}

	resolvers := map[string]pubsignals.StateResolver{
		resolverPrefix: resolver,
	}

	// EXECUTE VERIFICATION
	verifier, err := auth.NewVerifier(verificationKeyloader, resolvers, auth.WithIPFSGateway("https://ipfs.io"))
	if err != nil {
		log.Println(err.Error())
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	authResponse, err := verifier.FullVerify(
		ctx,
		string(tokenBytes),
		authRequest.(protocol.AuthorizationRequestMessage),
		pubsignals.WithAcceptedStateTransitionDelay(time.Minute*5))
	if err != nil {
		log.Println(err.Error())
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil,err
	}
	userID := authResponse.From
	messageBytes = []byte("User with ID " + userID + " Successfully authenticated")

	return messageBytes, nil
}



