# Detect Price Backend

Golang service powering authentication, photo analysis orchestration, and subscription billing.

## Directory Layout

- `cmd/` – CLI entrypoints (`serve`, `migrate`, etc.)
- `internal/ai` – AI integration surface (OpenRouter, pgvector) *(placeholder)*
- `internal/payment` – Midtrans hooks and billing logic *(placeholder)*
- `internal/photo` – Photo upload and enrichment APIs *(placeholder)*
- `internal/user` – Authentication, accounts, session handling
- `internal/db` – Migration runner and seeders
- `internal/utils` – Shared helpers and test utilities
- `test/` – High level integration tests

## Local Commands

```sh
# install dependencies
go mod tidy

# run the API
go run ./cmd/ -- serve

# execute migrations
go run ./cmd/ -- migrate:up

# run tests
go test ./...
```

## Environment

Copy `.env.example` from the repository root, populate required secrets, and export before running the service.
