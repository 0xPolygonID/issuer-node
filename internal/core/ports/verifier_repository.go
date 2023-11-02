package ports

import (
	// "context"
	"context"

	// "github.com/google/uuid"
	"github.com/iden3/iden3comm/v2/protocol"
)

// VerifierService is the interface implemented by the verifier service
type VerifierRepository interface {
	GetAuthRequest(ctx context.Context,schemaType string,schemaURL string,credSubject map[string]interface{})(protocol.AuthorizationRequestMessage,error)
	Callback(ctx context.Context,sessionId string,tokenString []byte) ([]byte,error)
}