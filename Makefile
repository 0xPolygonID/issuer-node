BIN := $(shell pwd)/bin
VERSION ?= $(shell git rev-parse --short HEAD)
GO?=$(shell which go)
export GOBIN := $(BIN)
export PATH := $(BIN):$(PATH)

BUILD_CMD := $(GO) install -ldflags "-X main.build=${VERSION}"

LOCAL_DEV_PATH = $(shell pwd)/infrastructure/local
DOCKER_COMPOSE_FILE := $(LOCAL_DEV_PATH)/docker-compose.yml
DOCKER_COMPOSE_CMD := docker-compose -p polygonid -f $(DOCKER_COMPOSE_FILE)

SCHEMA_DB_NAME := schema-$(shell date +"%s")
SCHEMA_DB_URL := "postgres://postgres:postgres@localhost:5432/$(SCHEMA_DB_NAME)?sslmode=disable"
SCHEMA_FILE_PATH := ./internal/db/schema/schema.sql
MIGRATION_PATH := ./internal/db/schema/migrations/

.PHONY: build
build:
	$(BUILD_CMD) ./cmd/...

.PHONY: clean
clean: ## Go clean
	$(GO) clean ./...

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: test-race
test-race:
	$(GO) test -v --race ./...

$(BIN)/oapi-codegen: tools.go go.mod go.sum ## install code generator for API files.
	go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen
	$(GO) install github.com/deepmap/oapi-codegen/cmd/oapi-codegen

.PHONY: api
api: $(BIN)/oapi-codegen
	$(BIN)/oapi-codegen -config ./api/config-oapi-codegen.yaml ./api/api.yaml > ./internal/api/api.gen.go
