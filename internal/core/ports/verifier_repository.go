package ports

import (
	// "context"
	"context"

	// "github.com/google/uuid"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// VerifierService is the interface implemented by the verifier service
type VerifierRepository interface {
	GetAuthRequest(ctx context.Context, schemaType string, schemaURL string, credSubject map[string]interface{}) (protocol.AuthorizationRequestMessage, error)
	Callback(ctx context.Context, sessionId string, tokenString []byte) ([]byte, error)
	Login(ctx context.Context, username string, password string) (*domain.SinzyLoginResponse, error)
	Logout(ctx context.Context, accessToken string)
	VerifyAccount(ctx context.Context, patronid string, parameterType string, parameterValue string)
	GetDigilockerURL(ctx context.Context, patronid string, accessToken string) (*domain.DigilockerURLResponse, error)
	GetDigilockerEAdharData(ctx context.Context, patronid string)
	GetRequestID(ctx context.Context, patronid string)
	GetListOfDocuments(ctx context.Context, patronid string, accessToken string, Adhar bool, PAN bool)
	GetPANDoc(ctx context.Context, patronid string)
	GetIdentity(ctx context.Context, patronid string, _type string, accessToken string) (*domain.VerificationIdentity, error)
	// UploadToIPFS(ctx context.Context, filePath string) (string, error)
	PullDocuments(ctx context.Context, patronid string, requestId string, accessToken string) (*domain.DigilockerDocumentList, error)
	VerifyPAN(ctx context.Context, itemId string, accessToken string, Authorization string, panNumber string, Name string, fuzzy bool, panStatus bool) (*domain.VerifyPANResponse, error)
	VerifyAdhar(ctx context.Context, itemId string, accessToken string, Authorization string, adharNumber string) (*domain.VerifyAdharResponse, error)
	GetDetails(ctx context.Context, partonId string, requestId string, accessToken string)
	VerifyGSTIN(ctx context.Context, partonId string,Authorization string, gstin string) (*domain.VerifyGSTINResponse, error)
}
