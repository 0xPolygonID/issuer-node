package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/verifiable"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/event"
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

func (n *notification) SendCreateCredentialNotification(ctx context.Context, e pubsub.Message) error {
	var cEvent event.CreateCredential
	if err := cEvent.Unmarshal(e); err != nil {
		return errors.New("sendCredentialNotification unexpected data type")
	}

	return n.sendCreateCredentialNotification(ctx, cEvent.IssuerID, cEvent.CredentialIDs)
}

func (n *notification) SendCreateConnectionNotification(ctx context.Context, e pubsub.Message) error {
	var cEvent event.CreateConnection
	if err := cEvent.Unmarshal(e); err != nil {
		return errors.New("sendCredentialNotification unexpected data type")
	}

	return n.sendCreateConnectionNotification(ctx, cEvent.IssuerID, cEvent.ConnectionID)
}

func (n *notification) sendCreateCredentialNotification(ctx context.Context, issuerID string, credIDs []string) error {
	issuerDID, err := core.ParseDID(issuerID)
	if err != nil {
		log.Error(ctx, "sendCreateCredentialNotification: failed to parse issuerID", "err", err.Error(), "issuerID", issuerID)
		return err
	}

	currentUserID := ""
	var currentConnection *domain.Connection
	credentials := make([]*domain.Claim, 0)
	for i, credID := range credIDs {
		credUUID, err := uuid.Parse(credID)
		if err != nil {
			log.Error(ctx, "sendCreateCredentialNotification: failed to parse credID", "err", err.Error(), "issuerID", issuerID, "credID", credID)
			continue
		}

		credential, err := n.credService.GetByID(ctx, issuerDID, credUUID)
		if err != nil {
			log.Warn(ctx, "sendCreateCredentialNotification: get credential", "err", err.Error(), "issuerID", issuerID, "credID", credID)
			continue
		}

		userDID, err := core.ParseDID(credential.OtherIdentifier)
		if err != nil {
			log.Error(ctx, "sendCreateCredentialNotification: failed to parse credential userID", "err", err.Error(), "issuerID", issuerID, "credID", credID)
			continue
		}

		if currentUserID == "" {
			currentUserID = credential.OtherIdentifier
			currentConnection, err = n.connService.GetByUserID(ctx, *issuerDID, *userDID)
			if err != nil {
				log.Warn(ctx, "sendCreateCredentialNotification: get currentConnection", "err", err.Error(), "issuerID", issuerID, "credID", credID)
				continue
			}
		}

		// If the credential is for a different user or it's the last credential for the currentUserID, send the notification
		if currentUserID != credential.OtherIdentifier || i == len(credIDs)-1 {
			if i == len(credIDs)-1 {
				credentials = append(credentials, credential)
			}

			credOfferBytes, subjectDIDDoc, err := getCredentialOfferData(currentConnection, credentials...)
			if err != nil {
				log.Error(ctx, "sendCreateCredentialNotification: getCredentialOfferData", "err", err.Error(), "issuerID", issuerID, "credID", credID)
				continue
			}
			// send notification
			log.Info(ctx, "sendCreateCredentialNotification: sending notification", "issuerID", issuerID, "subjectDIDDoc", subjectDIDDoc.ID)
			err = n.send(ctx, credOfferBytes, subjectDIDDoc)
			if err != nil {
				log.Error(ctx, "sendCreateCredentialNotification: send notification", "err", err.Error(), "issuerID", issuerID, "credID", credID)
			}

			currentUserID = credential.OtherIdentifier
			conn, err := n.connService.GetByUserID(ctx, *issuerDID, *userDID)
			if err != nil {
				log.Warn(ctx, "sendCreateCredentialNotification: get currentConnection", "err", err.Error(), "issuerID", issuerID, "credID", credID)
				return err
			}
			currentConnection = conn
			credentials = make([]*domain.Claim, 0)
			credentials = append(credentials, credential)
			continue
		}
		credentials = append(credentials, credential)
	}

	return nil
}

func (n *notification) sendCreateConnectionNotification(ctx context.Context, issuerID string, connID string) error {
	issuerDID, err := core.ParseDID(issuerID)
	if err != nil {
		log.Error(ctx, "sendCreateConnectionNotification: failed to parse issuerID", "err", err.Error(), "issuerID", issuerID, "connectionID", connID)
		return err
	}

	connUUID, err := uuid.Parse(connID)
	if err != nil {
		log.Error(ctx, "sendCreateConnectionNotification: failed to parse connID", "err", err.Error(), "issuerID", issuerID, "connectionID", connID)
		return err
	}

	conn, err := n.connService.GetByIDAndIssuerID(ctx, connUUID, *issuerDID)
	if err != nil {
		log.Error(ctx, "sendCreateConnectionNotification: failed to retrieve the connection", "err", err.Error(), "issuerID", issuerID, "connectionID", connID)
		return err
	}

	credentials, err := n.credService.GetAll(ctx, conn.IssuerDID, &ports.ClaimsFilter{Subject: conn.UserDID.String(), Proofs: []verifiable.ProofType{domain.AnyProofType}})
	if err != nil {
		log.Error(ctx, "sendCreateConnectionNotification: failed to retrieve the connection credentials", "err", err.Error(), "issuerID", issuerID, "connectionID", connID)
		return err
	}

	credOfferBytes, subjectDIDDoc, err := getCredentialOfferData(conn, credentials...)
	if err != nil {
		log.Error(ctx, "sendCreateConnectionNotification: getCredentialOfferData", "err", err.Error(), "issuerID", issuerID, "connID", connID)
		return err
	}

	return n.send(ctx, credOfferBytes, subjectDIDDoc)
}

func (n *notification) send(ctx context.Context, credOfferBytes []byte, subjectDIDDoc verifiable.DIDDocument) error {
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

func getCredentialOfferData(conn *domain.Connection, credentials ...*domain.Claim) (credOfferBytes []byte, subjectDIDDoc verifiable.DIDDocument, err error) {
	var managedDIDDoc verifiable.DIDDocument
	err = json.Unmarshal(conn.IssuerDoc, &managedDIDDoc)
	if err != nil {
		return nil, verifiable.DIDDocument{}, fmt.Errorf("unmarshal managedDIDDoc, err: %v", err.Error())
	}

	managedService, err := notifications.FindServiceByType(managedDIDDoc, verifiable.Iden3CommServiceType)
	if err != nil {
		return nil, verifiable.DIDDocument{}, fmt.Errorf("unmarshal managedService, err: %v", err.Error())
	}

	credOfferBytes, err = notifications.NewOfferMsg(managedService.ServiceEndpoint, credentials...)
	if err != nil {
		return nil, verifiable.DIDDocument{}, fmt.Errorf("newOfferMsg, err: %v", err.Error())
	}

	err = json.Unmarshal(conn.UserDoc, &subjectDIDDoc)
	if err != nil {
		return nil, verifiable.DIDDocument{}, fmt.Errorf("unmarshal subjectDIDDoc, err: %v", err.Error())
	}

	return
}
