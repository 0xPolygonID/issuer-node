BIN := $(shell pwd)/bin
VERSION ?= $(shell git rev-parse --short HEAD)
GO?=$(shell which go)
export GOBIN := $(BIN)
export PATH := $(BIN):$(PATH)

BUILD_CMD := $(GO) install -ldflags "-X main.build=${VERSION}"

LOCAL_DEV_PATH = $(shell pwd)/infrastructure/local
DOCKER_COMPOSE_FILE := $(LOCAL_DEV_PATH)/docker-compose.yml
DOCKER_COMPOSE_CMD := docker-compose -p sh-id-platform -f $(DOCKER_COMPOSE_FILE)

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

.PHONY: up
up:
	$(DOCKER_COMPOSE_CMD) up -d redis postgres vault

.PHONY: down
down:
	$(DOCKER_COMPOSE_CMD) down --remove-orphans

.PHONY: up-test
up-test:
	$(DOCKER_COMPOSE_CMD) up -d test_postgres

$(BIN)/platformid-migrate:
	$(BUILD_CMD) ./cmd/migrate

$(BIN)/install-goose: go.mod go.sum
	$(GO) install github.com/pressly/goose/v3

$(BIN)/golangci-lint: go.mod go.sum
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: db/migrate
db/migrate: $(BIN)/install-goose $(BIN)/platformid-migrate ## Install goose and apply migrations.
	sh -c '$(BIN)/migrate'

.PHONY: lint
lint: $(BIN)/golangci-lint
	  $(BIN)/golangci-lint run