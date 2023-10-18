package ports

import (
	"context"

	// core "github.com/iden3/go-iden3-core"

	"github.com/google/uuid"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// IndentityRepository is the interface implemented by the identity service
type RequestRepository interface {
	Save(ctx context.Context, conn db.Querier, connection *domain.Request) error
	GetByID(ctx context.Context, conn db.Querier, id uuid.UUID)(domain.Responce,error)
	Get(ctx context.Context, conn db.Querier,userDID string,requestType string)([]*domain.Responce,error)
	UpdateStatus(ctx context.Context, conn db.Querier,id uuid.UUID)(int64, error)
	NewNotification(ctx context.Context, conn db.Querier,notification *domain.NotificationData) (bool,error)
	GetNotifications(ctx context.Context, conn db.Querier,module string) ([]*domain.NotificationReponse,error)
	GetRequestsByRequestType(ctx context.Context, conn db.Querier, requestType string) ([]*domain.Responce,error)
	GetRequestsByUser(ctx context.Context, conn db.Querier, userDID string,All bool) ([]*domain.Responce,error)
	DeleteNotification(ctx context.Context, conn db.Querier, id uuid.UUID) (*domain.DeleteNotificationResponse,error)
	SaveUser(ctx context.Context, conn db.Querier, user *domain.UserRequest) (bool,error)
	GetUserID(ctx context.Context, conn db.Querier,udid string) (*domain.UserResponse,error)
	SignUp(ctx context.Context, conn db.Querier, user *domain.SignUpRequest) (bool,error)
	SignIn(ctx context.Context, conn db.Querier, username string, password string) (*domain.LoginResponse,error)
	// GetNotificationsForIssuer(ctx context.Context, conn db.Querier) ([]*domain.NotificationReponse,error)
	// GetByID(ctx context.Context, conn db.Querier, identifier core.DID) (*domain.Identity, error)
	// Get(ctx context.Context, conn db.Querier) (identities []string, err error)
	// GetUnprocessedIssuersIDs(ctx context.Context, conn db.Querier) (issuersIDs []*core.DID, err error)
	// HasUnprocessedStatesByID(ctx context.Context, conn db.Querier, identifier *core.DID) (bool, error)
	// HasUnprocessedAndFailedStatesByID(ctx context.Context, conn db.Querier, identifier *core.DID) (bool, error)
}
