package services

import (
	"context"

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



func ( r *requests) CreateRequest(ctx context.Context,req domain.VCRequest) (uuid.UUID,error){
	var s string = "Pending for KYC verification"
	if req.RequestType == "KYBGSTINCredentials"{
		s ="Pending for KYB verification"
	}

	_req := &domain.Request{
		ID: uuid.New(),
		User_id: req.UserDID,
		Schema_id: req.SchemaID,
		Issuer_id: req.UserDID,
		Active: true,
		Status: s,
		Verify_Status:"VC verification pending",
		Wallet_Status:s,
		CredentialType: req.CredentialType,
		Type: req.RequestType,
		RoleType: req.RoleType,
		ProofType: req.ProofType,
		ProofId: req.ProofId,
		Age: req.Age,
		Source: req.Source,
	}
	err := r.reqRepo.Save(ctx,r.storage.Pgx,_req)
	if err != nil{
		return uuid.Nil,err;
	}

return _req.ID, err;
}

func ( r *requests) GetRequest(ctx context.Context,Id uuid.UUID) (domain.Responce,error){

	res,err := r.reqRepo.GetByID(ctx,r.storage.Pgx,Id)
	return res,err;
}

func ( r *requests) GetAllRequests(ctx context.Context,userDID string,requestType string) ([]*domain.Responce,error){
	res,err := r.reqRepo.Get(ctx,r.storage.Pgx,userDID,requestType)
	return res,err;
}


func (r *requests) UpdateStatus(ctx context.Context,id uuid.UUID) (int64 , error){
	res, err := r.reqRepo.UpdateStatus(ctx,r.storage.Pgx,id)
	return res,err
}


func (r *requests) NewNotification(ctx context.Context,req *domain.NotificationData) (bool , error){
	res, err := r.reqRepo.NewNotification(ctx,r.storage.Pgx,req)
	return res,err
}

func (r *requests) GetNotifications(ctx context.Context,module string) ([]*domain.NotificationReponse , error){
	res, err := r.reqRepo.GetNotifications(ctx,r.storage.Pgx,module)
	return res,err
}

// func (r *requests) GetNotificationsForIssuer(ctx context.Context) ([]*domain.NotificationReponse , error){
// 	res, err := r.reqRepo.GetNotificationsForIssuer(ctx,r.storage.Pgx)
// 	return res,err
// }

func (r *requests) GetRequestsByRequestType(ctx context.Context, requestType string) ([]*domain.Responce , error){
	res, err := r.reqRepo.GetRequestsByRequestType(ctx,r.storage.Pgx,requestType)
	return res,err
}

func (r *requests) GetRequestsByUser(ctx context.Context, userDID string) ([]*domain.Responce , error){
	res, err := r.reqRepo.GetRequestsByUser(ctx,r.storage.Pgx,userDID)
	return res,err
}

