package db

import (
	"github.com/jmoiron/sqlx"
)

type Sqlx struct {
	DB *sqlx.DB
}

func NewSqlx(datasource string) *Sqlx {
	//db, err := sqlx.Connect("postgres", datasource)
	//if err != nil {
	//	log.Fatal(err)
	//}

	return &Sqlx{}
}
