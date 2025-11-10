COMPOSE_FILE := docker/docker-compose.yml
ENV_FILE := docker/.env.dev
COMPOSE_PROJECT_NAME ?= detect-price
COMPOSE := docker compose -p $(COMPOSE_PROJECT_NAME) -f $(COMPOSE_FILE) --env-file $(ENV_FILE)
NETWORK := $(COMPOSE_PROJECT_NAME)_default
DOCKER_ENV := MSYS_NO_PATHCONV=1 MSYS2_ARG_CONV_EXCL=*
ROOT_DIR := $(shell pwd -W 2>/dev/null || pwd)

include $(ENV_FILE)
export $(shell sed -n 's/^\([^#= ][^=]*\)=.*/\1/p' $(ENV_FILE))

.PHONY: help docker-up docker-down db-up db-down backend-up frontend-up db-logs cache-logs migrate migrate-down clean db-shell backend-build frontend-build full-build

help:
	@echo "Available targets:"
	@echo "  make db-up          # Start PostgreSQL + Redis services"
	@echo "  make db-down        # Stop PostgreSQL + Redis services"
	@echo "  make backend-build  # Build backend container image"
	@echo "  make frontend-build # Build frontend container image"
	@echo "  make full-build     # Build backend + frontend images"
	@echo "  make docker-up      # Start backend, frontend, db, cache (expects images built)"
	@echo "  make docker-down    # Stop and remove the entire stack"
	@echo "  make pgweb-up       # Launch pgweb UI on $(PGWEB_PORT)"
	@echo "  make migrate        # Run database migrations"
	@echo "  make migrate-down   # Roll back the last migration"
	@echo "  make db-shell       # Open psql shell inside the db container"
	@echo "  make db-logs        # Tail PostgreSQL logs"

backend-build:
	$(COMPOSE) build backend

frontend-build:
	$(COMPOSE) build frontend

full-build: backend-build frontend-build

docker-up:
	$(COMPOSE) up -d --no-build

docker-down:
	$(COMPOSE) down

db-up:
	$(COMPOSE) up -d db cache
	@$(COMPOSE) exec db sh -c 'until pg_isready -U $$POSTGRES_USER >/dev/null 2>&1; do sleep 1; done'

db-down:
	$(COMPOSE) stop db cache

pgweb-up: db-up
	$(COMPOSE) up -d pgweb

backend-up:
	$(COMPOSE) up -d backend

frontend-up:
	$(COMPOSE) up -d frontend

db-logs:
	$(COMPOSE) logs -f db

cache-logs:
	$(COMPOSE) logs -f cache

db-shell:
	$(COMPOSE) exec db psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

migrate: db-up
	@$(DOCKER_ENV) docker run --rm --network $(NETWORK) \
		-v "$(ROOT_DIR)/apps/backend:/workspace" \
		-w /workspace \
		-e DATABASE_URL=postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@db:5432/$(POSTGRES_DB)?sslmode=disable \
		golang:1.25-alpine \
		sh -c "apk add --no-cache git && go run -tags=debug ./cmd/ migrate:up"

migrate-down: db-up
	@$(DOCKER_ENV) docker run --rm --network $(NETWORK) \
		-v "$(ROOT_DIR)/apps/backend:/workspace" \
		-w /workspace \
		-e DATABASE_URL=postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@db:5432/$(POSTGRES_DB)?sslmode=disable \
		golang:1.25-alpine \
		sh -c "apk add --no-cache git && go run -tags=debug ./cmd/ migrate:down"

clean:
	$(COMPOSE) down -v

