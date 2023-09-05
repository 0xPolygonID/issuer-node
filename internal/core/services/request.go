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



func ( r *requests) CreateRequest(ctx context.Context,userId uuid.UUID , schemaId string) (uuid.UUID,error){

	req := &domain.Request{
		ID: uuid.New(),
		User_id: userId,
		Schema_id: schemaId,
		Issuer_id: schemaId,
		Active: true,
		
	}
	err := r.reqRepo.Save(ctx,r.storage.Pgx,req)
	if err != nil{
		return uuid.Nil,err;
	}

return req.ID, err;
}

func ( r *requests) GetRequest(ctx context.Context,userId uuid.UUID) error{

	r.reqRepo.GetByID(ctx,r.storage.Pgx,userId)
	return nil;
}
