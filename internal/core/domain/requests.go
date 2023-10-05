package domain

import (
	"time"

	"github.com/google/uuid"
)

// "github.com/iden3/iden3comm"

type Request struct{
	ID uuid.UUID
	Schema_id uuid.UUID
	User_id string
	Issuer_id string
	CredentialType string
	RequestType string `json:"requestType"`
	RoleType 	string `json:"roleType"`
	ProofType 	string `json:"proofType"`
	ProofId 	string `json:"proofId"`
	Age	string	`json:"proof"`
	Source 	string	`json:"source"`
	Active bool
	Status string
	Type string
	Verify_Status string
	Wallet_Status string
}

type VCRequest struct {
	SchemaID uuid.UUID `json:"schemaID"` 
	UserDID string `json:"userDID"`
	CredentialType string `json:"credentialType"`
	RequestType string `json:"requestType"`
	RoleType 	string `json:"roleType"`
	ProofType 	string `json:"proofType"`
	ProofId 	string `json:"proofId"`
	Age	string	`json:"proof"`
	Source 	string	`json:"source"`
}

type Responce struct{
	Id uuid.UUID	`json:"id"`
	SchemaID uuid.UUID `json:"schemaID"` 
	UserDID string `json:"userDID"`
	Issuer_id string `json:"issuer_id"`
	CredentialType string `json:"credentialType"`
	RequestType string `json:"requestType"`
	RoleType 	string `json:"roleType"`
	ProofType 	string `json:"proofType"`
	ProofId 	string `json:"proofId"`
	Age	string	`json:"proof"`
	Active bool 	`json:"active"`
	RequestStatus  string  `json:"request_status"`
	VerifyStatus string		`json:"verifier_status"`
	WalletStatus string		`json:"wallet_status"`
	Source 	string	`json:"source"`
	CreatedAt	   time.Time  `json:"created_at"`
	ModifiedAt     time.Time	`json:"modified_at"`
}
// type VCRequest struct {
// 	SchemaID string `json:"schemaID"`
// 	UserDID uuid.UUID `json:"userDID"`
// }