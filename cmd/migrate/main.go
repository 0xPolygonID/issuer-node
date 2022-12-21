package main

import (
	"log"

	"github.com/polygonid/polygonid-api/sh-id-platform/internal/config"
	"github.com/polygonid/polygonid-api/sh-id-platform/internal/db/schema"
)

func main() {
	cfg, err := config.Load("internal/config")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(cfg.Database.Url)
	if err := schema.Migrate(cfg.Database.Url); err != nil {
		log.Fatal(err)
	}

	log.Println("migration done!")
}
