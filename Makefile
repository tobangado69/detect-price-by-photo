COMPOSE_FILE := docker/docker-compose.yml
ENV_FILE := docker/.env.dev
COMPOSE_PROJECT_NAME ?= detect-price
COMPOSE := docker compose -p $(COMPOSE_PROJECT_NAME) -f $(COMPOSE_FILE) --env-file $(ENV_FILE)
NETWORK := $(COMPOSE_PROJECT_NAME)_default
DOCKER_ENV := MSYS_NO_PATHCONV=1 MSYS2_ARG_CONV_EXCL=*
ROOT_DIR := $(shell pwd -W 2>/dev/null || pwd)

include $(ENV_FILE)
export $(shell sed -n 's/^\([^#= ][^=]*\)=.*/\1/p' $(ENV_FILE))

.PHONY: help docker-up docker-down db-up db-down backend-up frontend-up mailhog-up mailhog-logs db-logs cache-logs migrate migrate-down seed fix-password update-rohim-password clean db-shell backend-build frontend-build full-build

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
	@echo "  make seed           # Seed database with initial data"
	@echo "  make fix-password   # Fix password hash validation error"
	@echo "  make update-rohim-password # Update Rohim's password after fix"
	@echo "  make mailhog-up     # Start MailHog (email testing service)"
	@echo "  make mailhog-logs   # View MailHog logs"
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

mailhog-up:
	$(COMPOSE) up -d mailhog
	@echo "✓ MailHog started!"
	@echo "  SMTP Server: localhost:1025"
	@echo "  Web UI: http://localhost:8025"

mailhog-logs:
	$(COMPOSE) logs -f mailhog

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

seed: db-up
	@echo "Seeding database with initial data..."
	@$(DOCKER_ENV) docker run --rm --network $(NETWORK) \
		-v "$(ROOT_DIR)/apps/backend:/workspace" \
		-w /workspace \
		-e DATABASE_URL=postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@db:5432/$(POSTGRES_DB)?sslmode=disable \
		golang:1.25-alpine \
		sh -c "apk add --no-cache git && go run -tags=debug ./cmd/ migrate:seed --force"
	@echo "✓ Database seeded successfully!"

fix-password: db-up
	@echo "======================================"
	@echo "Password Hash Validation Fix"
	@echo "======================================"
	@echo ""
	@echo "Step 1: Running migration to change password_hash to TEXT..."
	@$(DOCKER_ENV) docker run --rm --network $(NETWORK) \
		-v "$(ROOT_DIR)/apps/backend:/workspace" \
		-w /workspace \
		-e DATABASE_URL=postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@db:5432/$(POSTGRES_DB)?sslmode=disable \
		golang:1.25-alpine \
		sh -c "apk add --no-cache git && go run -tags=debug ./cmd/ migrate:up"
	@echo "✓ Migration completed"
	@echo ""
	@echo "Step 2: Re-seeding users with correct password hashes..."
	@$(DOCKER_ENV) docker run --rm --network $(NETWORK) \
		-v "$(ROOT_DIR)/apps/backend:/workspace" \
		-w /workspace \
		-e DATABASE_URL=postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@db:5432/$(POSTGRES_DB)?sslmode=disable \
		golang:1.25-alpine \
		sh -c "apk add --no-cache git && go run -tags=debug ./cmd/ migrate:seed --force"
	@echo "✓ Users re-seeded"
	@echo ""
	@echo "Step 3: Restarting backend service..."
	@$(COMPOSE) restart backend
	@echo "✓ Backend restarted"
	@echo ""
	@echo "======================================"
	@echo "Fix Applied Successfully!"
	@echo "======================================"
	@echo ""
	@echo "Test with these credentials:"
	@echo "  Admin User:"
	@echo "    Email: admin@detectprice.com"
	@echo "    Password: admin123"
	@echo ""
	@echo "  Regular User:"
	@echo "    Email: johndoe@example.com"
	@echo "    Password: secure.password"
	@echo ""
	@echo "API Endpoint: POST http://localhost:9871/api/v1/auth/signin/email"

update-rohim-password: db-up
	@echo "Updating Rohim's password to correct format..."
	@$(COMPOSE) exec -T db psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) < apps/backend/scripts/update_rohim_password.sql
	@echo "✓ Password updated! Login with: rohimjoy70@gmail.com / admin123"

clean:
	$(COMPOSE) down -v

