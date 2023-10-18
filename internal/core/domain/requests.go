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


type NotificationData struct{
	ID uuid.UUID `json:"id"`
	User_id string `json:"user_id"`
	Module string `json:"module"`
	NotificationType string `json:"notification_type"`
	NotificationTitle string `json:"notification_title"`
	NotificationMessage string `json:"notification_message"`
}

type NotificationReponse struct{
	ID uuid.UUID `json:"id"`
	User_id string `json:"user_id"`
	Module string `json:"module"`
	NotificationType string `json:"notification_type"`
	NotificationTitle string `json:"notification_title"`
	NotificationMessage string `json:"notification_message"`
	CreatedAt	   time.Time  `json:"created_at"`
}


type UserRequest struct{
	ID string `json:"id"`
	Name string `json:"name"`
	DOB string `json:"dob"`
	Owner string `json:"owner"`
	Username string `json:"username"`
	Password string `json:"password"`
	Gmail string `json:"gmail"`
	Phone string `json:"phone"`
	Gstin string `json:"gstin"`
	UserType string `json:"userType"`
	Address string `json:"address"`
	Adhar string `json:"adhar"`
	PAN string `json:"PAN"`
	DocumentationSource string `json:"documentationSource"`
}


type SignUpRequest struct {
	UserDID string `json:"userDID"`
	Email string `json:"email"`
	UserName string `json:"userName"`
	Password string `json:"password"`
	FullName string `json:"firstName"`
	Role string `json:"role"`
}

type LoginResponse struct {
	UserDID string `json:"userDID"`
	Email string `json:"email"`
	UserName string `json:"userName"`
	Password string `json:"password"`
	FullName string `json:"firstName"`
	Role string `json:"role"`
	Iscompleted bool `json:"iscompleted"`
}



type UserResponse struct{
	ID string `json:"id"`
	Name string `json:"name"`
	Owner string `json:"owner"`
	DOB string `json:"dob"`
	Phone string `json:"phone"`
	Username string `json:"username"`
	Gmail string `json:"gmail"`
	Gstin string `json:"gstin"`
	UserType string `json:"userType"`
	Address string `json:"address"`
	Adhar string `json:"adhar"`
	PAN string `json:"PAN"`
	DocumentationSource string `json:"documentationSource"`
	Iscompleted bool `json:"iscompleted"`
	CreatedAt	   time.Time  `json:"created_at"`
}

type DeleteNotificationResponse struct{
	Status bool `json:"status"`
	Msg string `json:"msg"`
}