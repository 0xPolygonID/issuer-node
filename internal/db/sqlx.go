package db

import (
	"log"

	"github.com/jmoiron/sqlx"
)

// Sqlx is a db conn
type Sqlx struct {
	DB *sqlx.DB
}

// NewSqlx opens database connection and returns it
func NewSqlx(datasource string) *Sqlx {
	db, err := sqlx.Connect("postgres", datasource)
	if err != nil {
		log.Fatal(err)
	}

	return &Sqlx{
		DB: db,
	}
}
