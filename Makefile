include .env-issuer
BIN := $(shell pwd)/bin
VERSION ?= $(shell git rev-parse --short HEAD)
GO?=$(shell which go)
export GOBIN := $(BIN)
export PATH := $(BIN):$(PATH)

BUILD_CMD := $(GO) install -ldflags "-X main.build=${VERSION}"

LOCAL_DEV_PATH = $(shell pwd)/infrastructure/local
DOCKER_COMPOSE_FILE := $(LOCAL_DEV_PATH)/docker-compose.yml
DOCKER_COMPOSE_FILE_INFRA := $(LOCAL_DEV_PATH)/docker-compose-infra.yml
DOCKER_COMPOSE_CMD := docker compose -p issuer -f $(DOCKER_COMPOSE_FILE)
DOCKER_COMPOSE_INFRA_CMD := docker compose -p issuer -f $(DOCKER_COMPOSE_FILE_INFRA)
ENVIRONMENT := ${ISSUER_ENVIRONMENT}

ISSUER_KMS_PROVIDER_LOCAL_STORAGE_FILE_PATH := ${ISSUER_KMS_PROVIDER_LOCAL_STORAGE_FILE_PATH}
ISSUER_KMS_ETH_PROVIDER := ${ISSUER_KMS_ETH_PROVIDER}

ISSUER_RESOLVER_FILE := ${ISSUER_RESOLVER_FILE}
REQUIRED_FILE := ${ISSUER_RESOLVER_PATH}

# Local environment overrides via godotenv
DOTENV_CMD = $(BIN)/godotenv
ENV = $(DOTENV_CMD) -f .env-issuer

.PHONY: build-local
build-local:
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
	$(GO) install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

.PHONY: api
api: $(BIN)/oapi-codegen
	$(BIN)/oapi-codegen -config ./api/config-oapi-codegen.yaml ./api/api.yaml > ./internal/api/api.gen.go

# Starts the infrastructure services
.PHONY: up
up:
ifeq ($(ISSUER_KMS_ETH_PROVIDER)$(ISSUER_KMS_BJJ_PROVIDER), localstoragelocalstorage)
		$(DOCKER_COMPOSE_INFRA_CMD) up -d redis postgres
else
	$(DOCKER_COMPOSE_INFRA_CMD) up -d redis postgres vault
endif

# If you want to use localstorage as a KMS provider, you need to run this command
.PHONY: up/localstorage
 up/localstorage:
	$(DOCKER_COMPOSE_INFRA_CMD) up -d redis postgres

# Build the docker image for the issuer-api
.PHONY: build
build:
	docker build -t issuer-api:local -f ./Dockerfile .

# Build the docker image for the issuer-ui
.PHONY: build-ui
build-ui:
	docker build -t issuer-ui:local -f ./ui/Dockerfile ./ui


.PHOMY: validate_issuer_resolver_file
validate_issuer_resolver_file:
	@if [ ! -f "$(REQUIRED_FILE)" ]; then \
  		if [ -z "$(ISSUER_RESOLVER_FILE)" ]; then \
			echo "ISSUER_RESOLVER_FILE env var is empty, and the file $(REQUIRED_FILE) doesn't exists."; \
			exit 1; \
		else \
			echo "ISSUER_RESOLVER_FILE is set, using it..."; \
		fi \
	else \
		echo "$(REQUIRED_FILE) environment is present, using it "; \
	fi

# Run the api, pending_publisher and notifications services
.PHONY: run
run: validate_issuer_resolver_file up
	COMPOSE_DOCKER_CLI_BUILD=1 $(DOCKER_COMPOSE_CMD) up -d api pending_publisher notifications

# Run the ui.
# First build the ui image and the api image
.PHONY: run-ui
run-ui: build-ui add-host-url-swagger
	COMPOSE_DOCKER_CLI_BUILD=1 $(DOCKER_COMPOSE_CMD) up -d ui

# Run all services
.PHONE: run-all
run-all: build build-ui add-host-url-swagger
	COMPOSE_DOCKER_CLI_BUILD=1 $(DOCKER_COMPOSE_CMD) up -d ui api pending_publisher notifications


.PHONY: down
down:
	$(DOCKER_COMPOSE_INFRA_CMD) down --remove-orphans
	$(DOCKER_COMPOSE_CMD) down --remove-orphans

.PHONY: stop
stop:
	$(DOCKER_COMPOSE_CMD) stop

.PHONY: stop-all
stop-all:
	$(DOCKER_COMPOSE_INFRA_CMD) stop
	$(DOCKER_COMPOSE_CMD) stop


.PHONY: up-test
up-test:
	$(DOCKER_COMPOSE_INFRA_CMD) up -d test_postgres vault test_local_files_apache

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

.PHONY: lint
lint: $(BIN)/golangci-lint
	  $(BIN)/golangci-lint run

.PHONY: lint-fix
lint-fix: $(BIN)/golangci-lint
		  $(BIN)/golangci-lint run --fix

## Usage:
## AWS: make private_key=XXX aws_access_key=YYY aws_secret_key=ZZZ aws_region=your-region import-private-key-to-kms
## localstorage and vault: make private_key=XXX import-private-key-to-kms
.PHONY: import-private-key-to-kms
import-private-key-to-kms:
ifeq ($(ISSUER_KMS_ETH_PROVIDER), aws)
	@echo "AWS"
	docker build --build-arg ISSUER_KMS_ETH_PROVIDER_AWS_ACCESS_KEY=$(aws_access_key) \
    		  --build-arg ISSUER_KMS_ETH_PROVIDER_AWS_SECRET_KEY=$(aws_secret_key) \
    		  --build-arg ISSUER_KMS_ETH_PROVIDER_AWS_REGION=$(aws_region) -t privadoid-kms-importer -f ./Dockerfile-kms-importer .
	$(eval result = $(shell docker run -it -v ./.env-issuer:/.env-issuer  \
		--network issuer-network \
		privadoid-kms-importer ./kms_priv_key_importer --privateKey=$(private_key)))
	@echo "result: $(result)"
	$(eval keyID = $(shell echo $(result) | grep "key created keyId=" | sed 's/.*keyId=//'))
	@if [ -n "$(keyID)" ]; then \
		docker run -it --rm -v ./.env-issuer:/.env-issuer --network issuer-network \
			privadoid-kms-importer sh ./aws_kms_material_key_importer.sh $(private_key) $(keyID) privadoid; \
	else \
		echo "something went wrong because keyID is empty"; \
	fi
else ifeq ($(ISSUER_KMS_ETH_PROVIDER), localstorage)
	@echo "LOCALSTORAGE"
	docker build -t privadoid-kms-importer -f ./Dockerfile-kms-importer .
	docker run --rm -it -v ./.env-issuer:/.env-issuer -v $(ISSUER_KMS_PROVIDER_LOCAL_STORAGE_FILE_PATH)/kms_localstorage_keys.json:/localstoragekeys/kms_localstorage_keys.json \
		--network issuer-network \
		privadoid-kms-importer ./kms_priv_key_importer --privateKey=$(private_key)
else ifeq ($(ISSUER_KMS_ETH_PROVIDER), vault)
	@echo "VAULT"
	docker build -t privadoid-kms-importer -f ./Dockerfile-kms-importer .
	docker run --rm -it -v ./.env-issuer:/.env-issuer --network issuer-network \
		privadoid-kms-importer ./kms_priv_key_importer --privateKey=$(private_key)
else
	@echo "ISSUER_KMS_ETH_PROVIDER is not set"
endif

.PHONY: print-vault-token
print-vault-token:
	$(eval TOKEN = $(shell docker logs issuer-vault-1 2>&1 | grep " .hvs" | awk  '{print $$2}' | tail -1 ))
	echo $(TOKEN)

.PHONY: add-vault-token
add-vault-token:
	$(eval TOKEN = $(shell docker logs issuer-vault-1 2>&1 | grep " .hvs" | awk  '{print $$2}' | tail -1 ))
	sed '/ISSUER_KEY_STORE_TOKEN/d' .env-issuer > .env-issuer.tmp
	@echo ISSUER_KEY_STORE_TOKEN=$(TOKEN) >> .env-issuer.tmp
	mv .env-issuer.tmp .env-issuer

.PHONY: add-host-url-swagger
add-host-url-swagger:
	@if [ $(ENVIRONMENT) != "" ] && [ $(ENVIRONMENT) != "local" ]; then \
		sed -i -e  "s#server-url = [^ ]*#server-url = \""${ISSUER_API_UI_SERVER_URL}"\"#g" api_ui/spec.html; \
	fi

# usage: make vault_token=xxx vault-export-keys
.PHONY: vault-export-keys
vault-export-keys:
	docker build -t issuer-vault-export-keys .
	docker run --rm -it --network=issuer-network -v $(shell pwd):/keys issuer-vault-export-keys ./vault-migrator -operation=export -output-file=keys.json -vault-token=$(vault_token) -vault-addr=http://vault:8200

# usage: make vault_token=xxx vault-import-keys
.PHONY: vault-import-keys
vault-import-keys:
	docker build -t issuer-vault-import-keys .
	docker run --rm -it --network=issuer-network -v $(shell pwd)/keys.json:/keys.json issuer-vault-import-keys ./vault-migrator -operation=import -input-file=keys.json -vault-token=$(vault_token) -vault-addr=http://vault:8200

# usage: make new_password=xxx change-vault-password
.PHONY: change-vault-password
change-vault-password:
	docker exec issuer-vault-1 \
	vault write auth/userpass/users/issuernode password=$(new_password)

.PHONY: print-commands
print-commands:
	@grep '^\s*\.[a-zA-Z_][a-zA-Z0-9_]*' Makefile


.PHONY: clean-volumes
clean-volumes:
	$(DOCKER_COMPOSE_INFRA_CMD) down -v