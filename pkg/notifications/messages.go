package notifications

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/iden3/iden3comm/packers"
	"github.com/iden3/iden3comm/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// NewOfferMsg returns an offer message
func NewOfferMsg(fetchURL string, credentials ...*domain.Claim) ([]byte, error) {
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

	return json.Marshal(credOffer)
}

func toProtocolCredentialOffer(credentials []*domain.Claim) []protocol.CredentialOffer {
	offers := make([]protocol.CredentialOffer, len(credentials))
	for i := range credentials {
		offers[i] = protocol.CredentialOffer{
			ID:          credentials[i].ID.String(),
			Description: "PolygonIDEarlyAdopter",
		}
	}

	return offers
}
