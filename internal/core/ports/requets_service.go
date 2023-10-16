package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	// // core "github.com/iden3/go-iden3-core"
	// "github.com/google/uuid"
	// "github.com/polygonid/sh-id-platform/internal/core/domain"
	// "github.com/polygonid/sh-id-platform/internal/db"
)

// IndentityRepository is the interface implemented by the identity service
type RequestService interface {
	CreateRequest(ctx context.Context, req domain.VCRequest) (uuid.UUID, error)
	GetRequest(ctx context.Context, Id uuid.UUID) (domain.Responce, error)
	GetAllRequests(ctx context.Context, userDID string, requestType string) ([]*domain.Responce, error)
	UpdateStatus(ctx context.Context, id uuid.UUID) (int64, error)
	NewNotification(ctx context.Context, notification *domain.NotificationData) (bool, error)
	GetNotifications(ctx context.Context, module string) ([]*domain.NotificationReponse, error)
	GetRequestsByRequestType(ctx context.Context, requestType string) ([]*domain.Responce, error)
	GetRequestsByUser(ctx context.Context, userDID string,All bool) ([]*domain.Responce, error)
	DeleteNotification(ctx context.Context, id uuid.UUID) (*domain.DeleteNotificationResponse, error)
	SaveUser(ctx context.Context, user *domain.UserRequest) (bool,error)
	GetUserID(ctx context.Context, username string, password string) (*domain.UserResponse, error)
	SignUp(ctx context.Context, user *domain.SignUpRequest) (bool,error)
	SignIn(ctx context.Context, username string, password string) (*domain.LoginResponse,error)

	// Save(ctx context.Context, conn db.Querier, connection *domain.Request) error
	// GetByID(ctx context.Context, conn db.Querier, id uuid.UUID)
	// GetByID(ctx context.Context, conn db.Querier, identifier core.DID) (*domain.Identity, error)
	// Get(ctx context.Context, conn db.Querier) (identities []string, err error)
	// GetUnprocessedIssuersIDs(ctx context.Context, conn db.Querier) (issuersIDs []*core.DID, err error)
	// HasUnprocessedStatesByID(ctx context.Context, conn db.Querier, identifier *core.DID) (bool, error)
	// HasUnprocessedAndFailedStatesByID(ctx context.Context, conn db.Querier, identifier *core.DID) (bool, error)
}
