package gateways

import (
	"context"
	"encoding/json"

	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/pkg/errors"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/http"
	"github.com/polygonid/sh-id-platform/internal/notifications"
)

// ErrNoDeviceInfoInPushService is an error when push service in did document doesn't contain device metadata
var ErrNoDeviceInfoInPushService = errors.New("no devices in push service")

const (
	// DeviceNotificationStatusSuccess is for pushes that are sent to APNS / FCM
	DeviceNotificationStatusSuccess domain.DeviceNotificationStatus = "success"
	// DeviceNotificationStatusRejected is for pushes that are rejected by APNS / FCM
	DeviceNotificationStatusRejected domain.DeviceNotificationStatus = "rejected"
	// DeviceNotificationStatusFailed is for pushes that were not sent
	DeviceNotificationStatusFailed domain.DeviceNotificationStatus = "failed"
)

// PushClient PPG for notify devices.
type PushClient struct {
	conn *http.Client
}

// NewPushNotificationClient create PPG client.
func NewPushNotificationClient(conn *http.Client) ports.NotificationGateway {
	return &PushClient{
		conn: conn,
	}
}

// Notify send notification in json format to push service with device metadata.
func (c *PushClient) Notify(ctx context.Context, msg json.RawMessage, userDIDDocument verifiable.DIDDocument) (*domain.UserNotificationResult, error) {
	// find service for push in did document
	pushService, err := notifications.FindNotificationService(userDIDDocument)
	if err != nil {
		return nil, err
	}

	if len(pushService.Metadata.Devices) == 0 {
		return nil, ErrNoDeviceInfoInPushService
	}

	reqData := domain.Notification{
		Metadata: pushService.Metadata,
		Message:  msg,
	}
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp, err := c.conn.Post(ctx, pushService.ServiceEndpoint, reqBody)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var result []domain.DeviceNotificationResult
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &domain.UserNotificationResult{Devices: result}, nil
}
