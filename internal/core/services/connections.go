package services

import (
	"context"
	"errors"

	"github.com/google/uuid"

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
