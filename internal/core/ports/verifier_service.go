package ports

import (
	"context"

	// "github.com/google/uuid"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

type VerifierService interface {
	GetAuthRequest(ctx context.Context, schemaType string, schemaURL string, credSubject map[string]interface{}) (protocol.AuthorizationRequestMessage, error)
	Callback(ctx context.Context, sessionId string, tokenString []byte) ([]byte, error)
	AccessDigiLocker(ctx context.Context, patronId string, requestId string, accessToken string, Adhar bool, PAN bool) (string, error)
	GetDigiLockerURL(ctx context.Context) (*domain.DigilockerURLResponse, error)
	VerifyPAN(ctx context.Context,PAN string,Name string) (*domain.VerifyPANResponse, error)
	VerifyAdhar(ctx context.Context,AdahrNumber string) (*domain.VerifyAadhaarResponse, error)
	VerifyGSTIN(ctx context.Context, gstin string) (*domain.VerifyGSTINResponseNew, error)
}
