package domain

import "github.com/google/uuid"
// "github.com/iden3/iden3comm"

type Request struct{
	ID uuid.UUID
	Schema_id string
	User_id uuid.UUID
	Issuer_id string
	Active bool
}

// type VCRequest struct {
// 	SchemaID string `json:"schemaID"`
// 	UserDID uuid.UUID `json:"userDID"`
// }