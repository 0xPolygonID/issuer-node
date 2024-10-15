package notifications

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

const (
	schemaParts = 2
)

// NewOfferMsg returns an offer message
func NewOfferMsg(fetchURL string, credentials ...*domain.Claim) (*protocol.CredentialsOfferMessage, error) {
	if len(credentials) == 0 {
		return nil, errors.New("no claims provided")
	}

	credentialOffers := toProtocolCredentialOffer(credentials)
	msgID := uuid.NewString()
	credOffer := &protocol.CredentialsOfferMessage{
		ID:       msgID,
		Typ:      packers.MediaTypePlainMessage,
		Type:     protocol.CredentialOfferMessageType,
		ThreadID: msgID,
		Body: protocol.CredentialsOfferMessageBody{
			URL:         fetchURL,
			Credentials: credentialOffers,
		},
		From: credentials[0].Issuer,
		To:   credentials[0].OtherIdentifier,
	}

	return credOffer, nil
}

// NewRevokedMsg returns a revoked message
func NewRevokedMsg(claim *domain.Claim) ([]byte, error) {
	msgID := uuid.NewString()
	statusUpdate := &protocol.CredentialStatusUpdateMessage{
		ID:       msgID,
		Typ:      packers.MediaTypePlainMessage,
		Type:     protocol.CredentialStatusUpdateMessageType,
		ThreadID: msgID,
		Body: protocol.CredentialStatusUpdateMessageBody{
			ID:     claim.ID.String(),
			Reason: "claim was revoked",
		},
		From: claim.Issuer,
		To:   claim.OtherIdentifier,
	}
	return json.Marshal(statusUpdate)
}

func toProtocolCredentialOffer(credentials []*domain.Claim) []protocol.CredentialOffer {
	offers := make([]protocol.CredentialOffer, len(credentials))
	for i := range credentials {
		offers[i] = protocol.CredentialOffer{
			ID:          credentials[i].ID.String(),
			Description: credentialType(credentials[i].SchemaType),
			Status:      protocol.CredentialOfferStatusCompleted,
		}
	}

	return offers
}

func credentialType(fullSchema string) string {
	rawSchema := strings.Split(fullSchema, "#")
	if len(rawSchema) == schemaParts {
		return rawSchema[1]
	}
	return fullSchema
}
