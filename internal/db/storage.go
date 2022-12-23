package db

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/gommon/log"
)

type Storages struct {
	Pgx *pgxpool.Pool
}

func NewStorage(connectionString string) (*Storages, error) {
	pgxConn, err := pgxpool.Connect(context.Background(), connectionString)
	if err != nil {
		return nil, err
	}
	return &Storages{
		Pgx: pgxConn,
	}, nil
}

// Close all connections to database
func (s *Storages) Close() error {
	log.Info("pgx are closing connection")
	s.Pgx.Close()
	return nil
}
