package domain

import "github.com/google/uuid"
// "github.com/iden3/iden3comm"

type Request struct{
	ID uuid.UUID
	Schema_id string
	User_id string
	Issuer_id string
	Active bool
}

type Responce struct{
	Id uuid.UUID	`json:"id"`
	SchemaID string `json:"schemaID"` 
	UserDID string `json:"userDID"`
	Issuer_id string `json:"issuer_id"`
	Active bool 	`json:"active"`
}
// type VCRequest struct {
// 	SchemaID string `json:"schemaID"`
// 	UserDID uuid.UUID `json:"userDID"`
// }