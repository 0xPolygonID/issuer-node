package services

import (
	"context"

	// "github.com/google/uuid"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type verifier struct {
	verRepo ports.VerifierRepository
	storage *db.Storage
}
func NewVerifier(reqRepo ports.VerifierRepository, storage *db.Storage) ports.VerifierService {
	return &verifier{
		verRepo: reqRepo,
		storage: storage,
	}
}


func (v *verifier) GetAuthRequest(ctx context.Context,schemaType string,schemaURL string,credSubject map[string]interface{})(protocol.AuthorizationRequestMessage,error){
	return v.verRepo.GetAuthRequest(ctx,schemaType,schemaURL,credSubject);
}

func (v *verifier) Callback(ctx context.Context, sessionId string,tokenString []byte)([]byte,error){
	return v.verRepo.Callback(ctx,sessionId,tokenString)
}