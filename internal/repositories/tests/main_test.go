package tests

import (
	"log"
	"os"
	"testing"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
)

var storage *db.Storage

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	conn := lookupPostgresURL()
	if conn == "" {
		conn = "postgres://postgres:postgres@localhost:5435"
	}

	cfg := config.Configuration{
		Database: config.Database{
			URL: conn,
		},
	}
	s, teardown, err := tests.NewTestStorage(&cfg)
	defer teardown()
	if err != nil {
		log.Println("failed to acquire test database")
		return 1
	}
	storage = s
	return m.Run()
}

func lookupPostgresURL() string {
	con, ok := os.LookupEnv("POSTGRES_TEST_DATABASE")
	if !ok {
		return ""
	}
	return con
}
