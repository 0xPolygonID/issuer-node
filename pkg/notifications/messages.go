package notifications

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/iden3/iden3comm/packers"
	"github.com/iden3/iden3comm/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

const (
	schemaParts = 2
)

// NewOfferMsg returns an offer message
func NewOfferMsg(fetchURL string, claim *domain.Claim) ([]byte, error) {
	msgID := uuid.NewString()
	credOffer := &protocol.CredentialsOfferMessage{
		ID:       msgID,
		Typ:      packers.MediaTypePlainMessage,
		Type:     protocol.CredentialOfferMessageType,
		ThreadID: msgID,
		Body: protocol.CredentialsOfferMessageBody{
			URL: fetchURL,
			Credentials: []protocol.CredentialOffer{
				{
					ID:          claim.ID.String(),
					Description: credentialType(claim.SchemaType),
				},
			},
		},
		From: claim.Issuer,
		To:   claim.OtherIdentifier,
	}
	return json.Marshal(credOffer)
}

func credentialType(fullSchema string) string {
	rawSchema := strings.Split(fullSchema, "#")
	if len(rawSchema) == schemaParts {
		return rawSchema[1]
	}
	return fullSchema
}
