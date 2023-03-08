BIN := $(shell pwd)/bin
VERSION ?= $(shell git rev-parse --short HEAD)
GO?=$(shell which go)
export GOBIN := $(BIN)
export PATH := $(BIN):$(PATH)

BUILD_CMD := $(GO) install -ldflags "-X main.build=${VERSION}"

LOCAL_DEV_PATH = $(shell pwd)/infrastructure/local
DOCKER_COMPOSE_FILE := $(LOCAL_DEV_PATH)/docker-compose.yml

DOCKER_COMPOSE_CMD := docker compose -p sh-id-platform -f $(DOCKER_COMPOSE_FILE)

.PHONY: build
build:
	$(BUILD_CMD) ./cmd/...


.PHONY: build/docker
build/docker: ## Build the docker image.
	DOCKER_BUILDKIT=1 \
	docker build \
		-f ./Dockerfile \
		-t sh-id-platform/api:$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

.PHONY: clean
clean: ## Go clean
	$(GO) clean ./...

.PHONY: tests
tests:
	$(GO) test -v ./... --count=1

.PHONY: test-race
test-race:
	$(GO) test -v --race ./...

$(BIN)/oapi-codegen: tools.go go.mod go.sum ## install code generator for API files.
	go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen
	$(GO) install github.com/deepmap/oapi-codegen/cmd/oapi-codegen

.PHONY: api
api: $(BIN)/oapi-codegen
	$(BIN)/oapi-codegen -config ./api/config-oapi-codegen.yaml ./api/api.yaml > ./internal/api/api.gen.go


.PHONY: api-ui
api-ui: $(BIN)/oapi-codegen
	$(BIN)/oapi-codegen -config ./api_ui/config-oapi-codegen.yaml ./api_ui/api.yaml > ./internal/api_admin/api.gen.go

.PHONY: up
up:
	$(DOCKER_COMPOSE_CMD) up -d redis postgres vault

.PHONY: run
run:
	$(eval TOKEN = $(shell docker logs sh-id-platform-test-vault 2>&1 | grep " .hvs" | awk  '{print $$2}' | tail -1 ))
	COMPOSE_DOCKER_CLI_BUILD=1 KEY_STORE_TOKEN=$(TOKEN) DOCKER_FILE="Dockerfile" $(DOCKER_COMPOSE_CMD) up -d  api-issuer

.PHONY: run-arm
run-arm:
	$(eval TOKEN = $(shell docker logs sh-id-platform-test-vault 2>&1 | grep " .hvs" | awk  '{print $$2}' | tail -1 ))
	COMPOSE_DOCKER_CLI_BUILD=1 KEY_STORE_TOKEN=$(TOKEN) DOCKER_FILE="Dockerfile-arm" $(DOCKER_COMPOSE_CMD) up -d  api-issuer

.PHONY: run-ui
run-ui:
	$(eval TOKEN = $(shell docker logs sh-id-platform-test-vault 2>&1 | grep " .hvs" | awk  '{print $$2}' | tail -1 ))
	COMPOSE_DOCKER_CLI_BUILD=1 KEY_STORE_TOKEN=$(TOKEN) DOCKER_FILE="Dockerfile" $(DOCKER_COMPOSE_CMD) up -d api-ui ui

.PHONY: run-arm-ui
run-arm-ui:
	$(eval TOKEN = $(shell docker logs sh-id-platform-test-vault 2>&1 | grep " .hvs" | awk  '{print $$2}' | tail -1 ))
	COMPOSE_DOCKER_CLI_BUILD=1 KEY_STORE_TOKEN=$(TOKEN) DOCKER_FILE="Dockerfile-arm" $(DOCKER_COMPOSE_CMD) up -d api-ui ui

.PHONY: down
down:
	$(DOCKER_COMPOSE_CMD) down --remove-orphans

.PHONY: stop
stop:
	$(DOCKER_COMPOSE_CMD) stop

.PHONY: up-test
up-test:
	$(DOCKER_COMPOSE_CMD) up -d test_postgres vault

$(BIN)/configurator:
	$(BUILD_CMD) ./cmd/configurator

.PHONY: clean-vault
clean-vault:
	rm -R infrastructure/local/.vault/data/init.out
	rm -R infrastructure/local/.vault/file/core/
	rm -R infrastructure/local/.vault/file/logical/
	rm -R infrastructure/local/.vault/file/sys/

$(BIN)/platformid-migrate:
	$(BUILD_CMD) ./cmd/migrate

$(BIN)/install-goose: go.mod go.sum
	$(GO) install github.com/pressly/goose/v3

$(BIN)/golangci-lint: go.mod go.sum
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: db/migrate
db/migrate: $(BIN)/install-goose $(BIN)/platformid-migrate ## Install goose and apply migrations.
	sh -c '$(BIN)/migrate'

.PHONY: config
config: $(BIN)/configurator
	sh -c '$(BIN)/configurator -template=config.toml.sample -output=config.toml'

.PHONY: lint
lint: $(BIN)/golangci-lint
	  $(BIN)/golangci-lint run

# usage: make private_key=xxx add-private-key
.PHONY: add-private-key
add-private-key:
	docker exec sh-id-platform-test-vault \
	vault write iden3/import/pbkey key_type=ethereum private_key=$(private_key)
