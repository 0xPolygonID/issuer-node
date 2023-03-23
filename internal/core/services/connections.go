package services

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"

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

func (c *connection) Delete(ctx context.Context, id uuid.UUID, deleteCredentials bool, issuerDID core.DID) error {
	return c.storage.Pgx.BeginFunc(ctx,
		func(tx pgx.Tx) error {
			if deleteCredentials {
				err := c.deleteCredentials(ctx, id, issuerDID, tx)
				if err != nil {
					return err
				}
			}
			return c.delete(ctx, id, issuerDID, tx)
		})
}

func (c *connection) DeleteCredentials(ctx context.Context, id uuid.UUID, issuerID core.DID) error {
	return c.deleteCredentials(ctx, id, issuerID, c.storage.Pgx)
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

func (c *connection) GetAllByIssuerID(ctx context.Context, issuerDID core.DID, query *string) ([]*domain.Connection, error) {
	return c.connRepo.GetAllByIssuerID(ctx, c.storage.Pgx, issuerDID, query)
}

func (c *connection) delete(ctx context.Context, id uuid.UUID, issuerDID core.DID, pgx db.Querier) error {
	err := c.connRepo.Delete(ctx, c.storage.Pgx, id, issuerDID)
	if err != nil {
		if errors.Is(err, repositories.ErrConnectionDoesNotExist) {
			return ErrConnectionDoesNotExist
		}
		return err
	}

	return nil
}

func (c *connection) deleteCredentials(ctx context.Context, id uuid.UUID, issuerID core.DID, pgx db.Querier) error {
	return c.connRepo.DeleteCredentials(ctx, pgx, id, issuerID)
}
