BIN := $(shell pwd)/bin
VERSION ?= $(shell git rev-parse --short HEAD)
GO?=$(shell which go)
export GOBIN := $(BIN)
export PATH := $(BIN):$(PATH)

BUILD_CMD := $(GO) install -ldflags "-X main.build=${VERSION}"

LOCAL_DEV_PATH = $(shell pwd)/infrastructure/local
DOCKER_COMPOSE_FILE := $(LOCAL_DEV_PATH)/docker-compose.yml
DOCKER_COMPOSE_CMD := docker compose -p issuer -f $(DOCKER_COMPOSE_FILE)


# Local environment overrides via godotenv
DOTENV_CMD = $(BIN)/godotenv
ENV = $(DOTENV_CMD) -f .env-issuer

.PHONY: build
build:
	$(BUILD_CMD) ./cmd/...


.PHONY: build/docker
build/docker: ## Build the docker image.
	DOCKER_BUILDKIT=1 \
	docker build \
		-f ./Dockerfile \
		-t issuer/api:$(VERSION) \
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
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile" $(DOCKER_COMPOSE_CMD) up -d api

.PHONY: run-arm
run-arm:
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile-arm" $(DOCKER_COMPOSE_CMD) up -d api

.PHONY: run-ui
run-ui:
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile" $(DOCKER_COMPOSE_CMD) up -d api-ui ui

.PHONY: run-ui-arm
run-ui-arm:
	COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_FILE="Dockerfile-arm" $(DOCKER_COMPOSE_CMD) up -d api-ui ui

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

$(BIN)/godotenv: tools.go go.mod go.sum
	$(GO) install github.com/joho/godotenv/cmd/godotenv

.PHONY: db/migrate
db/migrate: $(BIN)/install-goose $(BIN)/godotenv $(BIN)/platformid-migrate ## Install goose and apply migrations.
	$(ENV) sh -c '$(BIN)/migrate'

.PHONY: config
config: $(BIN)/configurator
	sh -c '$(BIN)/configurator -template=config.toml.sample -output=config.toml'

.PHONY: lint
lint: $(BIN)/golangci-lint
	  $(BIN)/golangci-lint run

# usage: make private_key=xxx add-private-key
.PHONY: add-private-key
add-private-key:
	docker exec issuer-vault-1 \
	vault write iden3/import/pbkey key_type=ethereum private_key=$(private_key)

.PHONY: print-vault-token
print-vault-token:
	$(eval TOKEN = $(shell docker logs issuer-vault-1 2>&1 | grep " .hvs" | awk  '{print $$2}' | tail -1 ))
	@echo $(TOKEN)

.PHONY: add-vault-token
add-vault-token:
	$(eval TOKEN = $(shell docker logs issuer-vault-1 2>&1 | grep " .hvs" | awk  '{print $$2}' | tail -1 ))
	sed '/ISSUER_KEY_STORE_TOKEN/d' .env-issuer > .env-issuer.tmp
	@echo ISSUER_KEY_STORE_TOKEN=$(TOKEN) >> .env-issuer.tmp
	@MV .env-issuer.tmp .env-issuer

.PHONY: rm-issuer-imgs
rm-issuer-imgs: stop
	$(shell docker rmi -f issuer_api issuer_ui issuer_api-ui) || true

.PHONY: restart-ui
restart-ui: rm-issuer-imgs run-ui

.PHONY: restart-ui-arm
restart-ui-arm: rm-issuer-imgs run-ui-arm
