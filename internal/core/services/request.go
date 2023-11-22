package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type requests struct {
	reqRepo ports.RequestRepository
	storage *db.Storage
}

// NewConnection returns a new connection service
func NewRequests(reqRepo ports.RequestRepository, storage *db.Storage) ports.RequestService {
	return &requests{
		reqRepo: reqRepo,
		storage: storage,
	}
}

func (r *requests) CreateRequest(ctx context.Context, req domain.VCRequest) (uuid.UUID, error) {
	var s string = "Pending for KYC verification"

	user, err := r.reqRepo.GetUserID(ctx, r.storage.Pgx, req.UserDID)
	if err != nil {
		fmt.Println("Error in getting user", err)
		return uuid.Nil, err
	}
	if req.CredentialType == "KYCAgeCredentialAadhar" || req.CredentialType == "ValidCredentialAadhar" {
		if user.AdharStatus == true {
			s = "Identity is Verified"
		}
	} else if req.CredentialType == "KYCAgeCredentialPAN" || req.CredentialType == "ValidCredentialPAN" {
		if user.PANStatus == true {
			s = "Identity is Verified"
		}
	} else if req.CredentialType == "KYBGSTINCredentials" {
		if user.GstinStatus == true {
			s = "Identity is Verified"
		} else {
			s = "Pending for KYB verification"
		}
	} else {
		return uuid.Nil, errors.New("Invalid Request Type")
	}

	_req := &domain.Request{
		ID:             uuid.New(),
		User_id:        req.UserDID,
		Schema_id:      req.SchemaID,
		Issuer_id:      req.UserDID,
		Active:         true,
		Status:         s,
		Verify_Status:  "VC verification pending",
		Wallet_Status:  s,
		CredentialType: req.CredentialType,
		Type:           req.RequestType,
		RoleType:       req.RoleType,
		ProofType:      req.ProofType,
		ProofId:        req.ProofId,
		Age:            req.Age,
		Source:         req.Source,
	}
	err = r.reqRepo.Save(ctx, r.storage.Pgx, _req)
	if err != nil {
		return uuid.Nil, err
	}

	return _req.ID, err
}

func (r *requests) GetRequest(ctx context.Context, Id uuid.UUID) (domain.Responce, error) {

	res, err := r.reqRepo.GetByID(ctx, r.storage.Pgx, Id)
	return res, err
}

func (r *requests) GetAllRequests(ctx context.Context, userDID string, requestType string) ([]*domain.Responce, error) {
	res, err := r.reqRepo.Get(ctx, r.storage.Pgx, userDID, requestType)
	return res, err
}

func (r *requests) UpdateStatus(ctx context.Context, id uuid.UUID, issuer_status string, verifier_status string, wallet_status string) (int64, error) {
	res, err := r.reqRepo.UpdateStatus(ctx, r.storage.Pgx, id, issuer_status, verifier_status, wallet_status)
	return res, err
}

func (r *requests) NewNotification(ctx context.Context, req *domain.NotificationData) (bool, error) {
	res, err := r.reqRepo.NewNotification(ctx, r.storage.Pgx, req)
	return res, err
}

func (r *requests) GetNotifications(ctx context.Context, module string) ([]*domain.NotificationReponse, error) {
	res, err := r.reqRepo.GetNotifications(ctx, r.storage.Pgx, module)
	return res, err
}

// func (r *requests) GetNotificationsForIssuer(ctx context.Context) ([]*domain.NotificationReponse , error){
// 	res, err := r.reqRepo.GetNotificationsForIssuer(ctx,r.storage.Pgx)
// 	return res,err
// }

func (r *requests) GetRequestsByRequestType(ctx context.Context, requestType string) ([]*domain.Responce, error) {
	res, err := r.reqRepo.GetRequestsByRequestType(ctx, r.storage.Pgx, requestType)
	return res, err
}

func (r *requests) GetRequestsByUser(ctx context.Context, userDID string, All bool) ([]*domain.Responce, error) {
	res, err := r.reqRepo.GetRequestsByUser(ctx, r.storage.Pgx, userDID, All)
	return res, err
}

func (r *requests) DeleteNotification(ctx context.Context, id uuid.UUID) (*domain.DeleteNotificationResponse, error) {
	res, err := r.reqRepo.DeleteNotification(ctx, r.storage.Pgx, id)
	return res, err
}

func (r *requests) SaveUser(ctx context.Context, user *domain.UserRequest) (bool, error) {
	res, err := r.reqRepo.SaveUser(ctx, r.storage.Pgx, user)
	return res, err
}

func (r *requests) GetUserID(ctx context.Context, udid string) (*domain.UserResponse, error) {
	res, err := r.reqRepo.GetUserID(ctx, r.storage.Pgx, udid)
	return res, err
}

func (r *requests) SignUp(ctx context.Context, user *domain.SignUpRequest) (bool, error) {
	res, err := r.reqRepo.SignUp(ctx, r.storage.Pgx, user)
	return res, err
}

func (r *requests) SignIn(ctx context.Context, username string, password string) (*domain.LoginResponse, error) {
	res, err := r.reqRepo.SignIn(ctx, r.storage.Pgx, username, password)
	return res, err
}

func (r *requests) UpdateVerificationStatus(ctx context.Context, id string, field string, status bool) (int64, error) {
	res, err := r.reqRepo.UpdateVerificationStatus(ctx, r.storage.Pgx, id, field, status)
	return res, err
}
