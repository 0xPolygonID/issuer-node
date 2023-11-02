package ports

import (
	"context"

	// "github.com/google/uuid"
	"github.com/iden3/iden3comm/v2/protocol"
)


type VerifierService interface {
	GetAuthRequest(ctx context.Context,schemaType string,schemaURL string,credSubject map[string]interface{}) (protocol.AuthorizationRequestMessage,error)
	Callback(ctx context.Context,sessionId string,tokenString []byte) ([]byte,error)
}