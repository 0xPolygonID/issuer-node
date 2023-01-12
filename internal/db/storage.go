package db

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/gommon/log"
)

// Storage defines the postgres storage
type Storage struct {
	Pgx *pgxpool.Pool
}

// NewStorage creates and returns a new Pgx storage connection
func NewStorage(connectionString string) (*Storage, error) {
	pgxConn, err := pgxpool.Connect(context.Background(), connectionString)
	if err != nil {
		return nil, err
	}
	return &Storage{
		Pgx: pgxConn,
	}, nil
}

// Close all connections to database
func (s *Storage) Close() error {
	log.Info("pgx is closing connection")
	s.Pgx.Close()
	return nil
}
