package domain

import (
	"encoding/json"

	"github.com/iden3/go-schema-processor/v2/verifiable"
)

const (
	// DeviceNotificationStatusSuccess is for pushes that are sent to APNS / FCM
	DeviceNotificationStatusSuccess DeviceNotificationStatus = "success"
)

// DeviceNotificationStatus is a notification status
type DeviceNotificationStatus string

// Notification contains the information to be sent
type Notification struct {
	Metadata verifiable.PushMetadata `json:"metadata"`
	Message  json.RawMessage         `json:"message"`
}

// UserNotificationResult is a result of push gateway
type UserNotificationResult struct {
	Devices []DeviceNotificationResult `json:"devices"`
}

// DeviceNotificationResult is a result of push gateway
type DeviceNotificationResult struct {
	Device verifiable.EncryptedDeviceMetadata `json:"device"`
	Status DeviceNotificationStatus           `json:"status"`
	Reason string                             `json:"reason"`
}
