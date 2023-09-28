package ports

import (
	"context"
	"encoding/json"

	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
)

// NotificationService represents the notification service interface
type NotificationService interface {
	SendCreateCredentialNotification(ctx context.Context, payload pubsub.Message) error
	SendCreateConnectionNotification(ctx context.Context, payload pubsub.Message) error
}

// NotificationGateway represents the notification interface
type NotificationGateway interface {
	Notify(ctx context.Context, msg json.RawMessage, userDIDDocument verifiable.DIDDocument) (*domain.UserNotificationResult, error)
}
