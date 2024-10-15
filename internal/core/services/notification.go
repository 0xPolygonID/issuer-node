package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/event"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/log"
	notifications2 "github.com/polygonid/sh-id-platform/internal/notifications"
	"github.com/polygonid/sh-id-platform/internal/pubsub"
)

type notification struct {
	notificationGateway ports.NotificationGateway
	connService         ports.ConnectionService
	credService         ports.ClaimService
}

// NewNotification returns a Notification Service
func NewNotification(notificationGateway ports.NotificationGateway, connService ports.ConnectionService, credService ports.ClaimService) ports.NotificationService {
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

func (n *notification) SendRevokeCredentialNotification(ctx context.Context, payload pubsub.Message) error {
	var rEvent event.CreateState
	if err := rEvent.Unmarshal(payload); err != nil {
		return errors.New("sendRevokeCredentialNotification unexpected data type")
	}

	return n.sendRevokeCredentialNotification(ctx, rEvent.State)
}

func (n *notification) SendCreateConnectionNotification(ctx context.Context, e pubsub.Message) error {
	var cEvent event.CreateConnection
	if err := cEvent.Unmarshal(e); err != nil {
		return errors.New("sendCredentialNotification unexpected data type")
	}

	return n.sendCreateConnectionNotification(ctx, cEvent.IssuerID, cEvent.ConnectionID)
}

func (n *notification) sendRevokeCredentialNotification(ctx context.Context, state string) error {
	rCreds, err := n.credService.GetRevoked(ctx, state)
	if err != nil {
		log.Error(ctx, "sendRevokeCredentialNotification: get revoked credentials", "err", err.Error(), "state", state)
		return err
	}

	log.Info(ctx, "sendRevokeCredentialNotification: credentials to revoke", "count", len(rCreds))
	for _, rCred := range rCreds {
		issuerDID, err := w3c.ParseDID(rCred.Issuer)
		if err != nil {
			log.Error(ctx, "sendRevokeCredentialNotification: failed to parse issuerID", "err", err.Error(), "issuerID", rCred.Issuer, "credID", rCred.ID)
			return err
		}

		userDID, err := w3c.ParseDID(rCred.OtherIdentifier)
		if err != nil {
			log.Error(ctx, "sendRevokeCredentialNotification: failed to parse credential userID", "err", err.Error(), "issuerID", rCred.Issuer, "credID", rCred.ID)
			return err
		}

		connection, err := n.connService.GetByUserID(ctx, *issuerDID, *userDID)
		if err != nil {
			log.Warn(ctx, "sendRevokeCredentialNotification: get connection", "err", err.Error(), "issuerID", rCred.Issuer, "credID", rCred.ID)
			return err
		}

		credOfferBytes, subjectDIDDoc, err := getRevokedCredentialData(connection, rCred)
		if err != nil {
			log.Error(ctx, "sendRevokeCredentialNotification: getCredentialOfferData", "err", err.Error(), "issuerID", rCred.Issuer, "credID", rCred.ID)
			return err
		}

		// send notification
		log.Info(ctx, "sendRevokeCredentialNotification: sending notification", "issuerID", rCred.Issuer, "subjectDIDDoc", subjectDIDDoc.ID)
		err = n.send(ctx, credOfferBytes, subjectDIDDoc)
		if err != nil {
			log.Error(ctx, "sendRevokeCredentialNotification: send notification", "err", err.Error(), "issuerID", rCred.Issuer, "credID", rCred.ID)
			return err
		}
	}

	return nil
}

func (n *notification) sendCreateCredentialNotification(ctx context.Context, issuerID string, credIDs []string) error {
	issuerDID, err := w3c.ParseDID(issuerID)
	if err != nil {
		log.Error(ctx, "sendCreateCredentialNotification: failed to parse issuerID", "err", err.Error(), "issuerID", issuerID)
		return err
	}

	credentialsUserID := ""
	var connection *domain.Connection
	credentials := make([]*domain.Claim, len(credIDs))
	for i, credID := range credIDs {
		credUUID, err := uuid.Parse(credID)
		if err != nil {
			log.Error(ctx, "sendCreateCredentialNotification: failed to parse credID", "err", err.Error(), "issuerID", issuerID, "credID", credID)
			return err
		}

		credential, err := n.credService.GetByID(ctx, issuerDID, credUUID)
		if err != nil {
			log.Warn(ctx, "sendCreateCredentialNotification: get credential", "err", err.Error(), "issuerID", issuerID, "credID", credID)
			return err
		}

		if credentialsUserID == "" {
			userDID, err := w3c.ParseDID(credential.OtherIdentifier)
			if err != nil {
				log.Error(ctx, "sendCreateCredentialNotification: failed to parse credential userID", "err", err.Error(), "issuerID", issuerID, "credID", credID)
				return err
			}

			credentialsUserID = credential.OtherIdentifier
			connection, err = n.connService.GetByUserID(ctx, *issuerDID, *userDID)
			if err != nil {
				log.Warn(ctx, "sendCreateCredentialNotification: get connection", "err", err.Error(), "issuerID", issuerID, "credID", credID)
				return err
			}
		}
		credentials[i] = credential
	}

	credOfferBytes, subjectDIDDoc, err := getCredentialOfferData(connection, credentials...)
	if err != nil {
		log.Error(ctx, "sendCreateCredentialNotification: getCredentialOfferData", "err", err.Error(), "issuerID", issuerID)
		return err
	}

	// send notification
	log.Info(ctx, "sendCreateCredentialNotification: sending notification", "issuerID", issuerID, "subjectDIDDoc", subjectDIDDoc.ID)
	err = n.send(ctx, credOfferBytes, subjectDIDDoc)
	if err != nil {
		log.Error(ctx, "sendCreateCredentialNotification: send notification", "err", err.Error(), "issuerID", issuerID)
		return err
	}

	return nil
}

func (n *notification) sendCreateConnectionNotification(ctx context.Context, issuerID string, connID string) error {
	issuerDID, err := w3c.ParseDID(issuerID)
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

	credentials, _, err := n.credService.GetAll(ctx, conn.IssuerDID, &ports.ClaimsFilter{Subject: conn.UserDID.String(), Proofs: []verifiable.ProofType{domain.AnyProofType}})
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

	managedService, err := notifications2.FindServiceByType(managedDIDDoc, verifiable.Iden3CommServiceType)
	if err != nil {
		return nil, verifiable.DIDDocument{}, fmt.Errorf("unmarshal managedService, err: %v", err.Error())
	}

	credOffer, err := notifications2.NewOfferMsg(managedService.ServiceEndpoint, credentials...)
	if err != nil {
		return nil, verifiable.DIDDocument{}, fmt.Errorf("newOfferMsg, err: %v", err.Error())
	}

	credOfferBytes, err = json.Marshal(credOffer)
	if err != nil {
		return nil, verifiable.DIDDocument{}, fmt.Errorf("marshal credOffer, err: %v", err.Error())
	}

	err = json.Unmarshal(conn.UserDoc, &subjectDIDDoc)
	if err != nil {
		return nil, verifiable.DIDDocument{}, fmt.Errorf("unmarshal subjectDIDDoc, err: %v", err.Error())
	}

	return
}

func getRevokedCredentialData(conn *domain.Connection, credential *domain.Claim) (credOfferBytes []byte, subjectDIDDoc verifiable.DIDDocument, err error) {
	var managedDIDDoc verifiable.DIDDocument
	err = json.Unmarshal(conn.IssuerDoc, &managedDIDDoc)
	if err != nil {
		return nil, verifiable.DIDDocument{}, fmt.Errorf("unmarshal managedDIDDoc, err: %v", err.Error())
	}

	credOfferBytes, err = notifications2.NewRevokedMsg(credential)
	if err != nil {
		return nil, verifiable.DIDDocument{}, fmt.Errorf("NewRevokedMsg, err: %v", err.Error())
	}

	err = json.Unmarshal(conn.UserDoc, &subjectDIDDoc)
	if err != nil {
		return nil, verifiable.DIDDocument{}, fmt.Errorf("unmarshal subjectDIDDoc, err: %v", err.Error())
	}

	return
}
