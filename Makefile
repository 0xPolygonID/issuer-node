

LOCAL_DEV_PATH = $(shell pwd)/infrastructure/local
DOCKER_COMPOSE_FILE := $(LOCAL_DEV_PATH)/docker-compose.yml
DOCKER_COMPOSE_CMD := docker-compose -p sh-id-platform -f $(DOCKER_COMPOSE_FILE)

.PHONY: up
up:
	$(DOCKER_COMPOSE_CMD) up -d redis postgres

.PHONY: down
down:
	$(DOCKER_COMPOSE_CMD) down --remove-orphans

.PHONY: up-test
up-test:
	$(DOCKER_COMPOSE_CMD) up -d test_postgres