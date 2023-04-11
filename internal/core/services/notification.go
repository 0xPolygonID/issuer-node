package services

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/verifiable"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/notifications"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
)

type notification struct {
	notificationGateway ports.NotificationGateway
	connService         ports.ConnectionsService
	credService         ports.ClaimsService
}

// NewNotification returns a Notification Service
func NewNotification(notificationGateway ports.NotificationGateway, connService ports.ConnectionsService, credService ports.ClaimsService) ports.NotificationService {
	return &notification{
		notificationGateway: notificationGateway,
		connService:         connService,
		credService:         credService,
	}
}

func (n *notification) SendCreateCredentialNotification(ctx context.Context, e pubsub.Event) error {
	var cEvent pubsub.CreateCredentialEvent
	if err := pubsub.UnmarshalEvent(e, &cEvent); err != nil {
		return errors.New("sendCredentialNotification unexpected data type")
	}

	return n.sendCreateCredentialNotification(ctx, cEvent.IssuerID, cEvent.CredentialID)
}

func (n *notification) sendCreateCredentialNotification(ctx context.Context, issuerID string, credID string) error {
	issuerDID, err := core.ParseDID(issuerID)
	if err != nil {
		return err
	}

	credUUID, err := uuid.Parse(credID)
	if err != nil {
		return err
	}

	credential, err := n.credService.GetByID(ctx, issuerDID, credUUID)
	if err != nil {
		return err
	}

	userDID, err := core.ParseDID(credential.OtherIdentifier)
	if err != nil {
		return err
	}

	conn, err := n.connService.GetByUserID(ctx, *issuerDID, *userDID)
	if err != nil {
		return err
	}

	var managedDIDDoc verifiable.DIDDocument
	err = json.Unmarshal(conn.IssuerDoc, &managedDIDDoc)
	if err != nil {
		return err
	}

	managedService, err := notifications.FindServiceByType(managedDIDDoc, verifiable.Iden3CommServiceType)
	if err != nil {
		return err
	}

	credOfferBytes, err := notifications.NewOfferMsg(managedService.ServiceEndpoint, credential)
	if err != nil {
		return err
	}

	var subjectDIDDoc verifiable.DIDDocument
	err = json.Unmarshal(conn.UserDoc, &subjectDIDDoc)
	if err != nil {
		return err
	}

	res, err := n.notificationGateway.Notify(ctx, credOfferBytes, subjectDIDDoc)
	if err != nil {
		return err
	}

	for _, nr := range res.Devices {
		if nr.Status != domain.DeviceNotificationStatusSuccess {
			log.Error(ctx, "failed to send push notification to certain user device",
				"device encrypted info", nr.Device.Ciphertext, "reason", nr.Reason)
			return errors.New("failed to send push notification")
		}
	}

	return nil
}
