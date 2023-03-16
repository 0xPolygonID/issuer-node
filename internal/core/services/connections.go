package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

// ErrConnectionDoesNotExist connection does not exist
var ErrConnectionDoesNotExist = errors.New("connection does not exist")

type connection struct {
	connRepo ports.ConnectionsRepository
	storage  *db.Storage
}

// NewConnection returnas a new connection service
func NewConnection(connRepo ports.ConnectionsRepository, storage *db.Storage) ports.ConnectionsService {
	return &connection{
		connRepo: connRepo,
		storage:  storage,
	}
}

func (c *connection) Delete(ctx context.Context, id uuid.UUID) error {
	err := c.connRepo.Delete(ctx, c.storage.Pgx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrConnectionDoesNotExist) {
			return ErrConnectionDoesNotExist
		}
		return err
	}

	return nil
}

func (c *connection) GetByIDAndIssuerID(ctx context.Context, id uuid.UUID, issuerDID core.DID) (*domain.Connection, error) {
	conn, err := c.connRepo.GetByIDAndIssuerID(ctx, c.storage.Pgx, id, issuerDID)
	if err != nil {
		if errors.Is(err, repositories.ErrConnectionDoesNotExist) {
			return nil, ErrConnectionDoesNotExist
		}
		return nil, err
	}

	return conn, nil
}
