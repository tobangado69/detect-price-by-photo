# Database Setup Summary

## ⚙️ Local Docker Workflow

The repository now provides a `Makefile` to streamline local Docker operations.  
Use the targets below from the project root:

```bash
# Start PostgreSQL and Redis (with health checks)
make db-up

# Run database migrations inside an ephemeral Go container
make migrate

# Tear everything down
make docker-down
```

Under the hood:
- `docker/.env.dev` holds the shared database credentials
- `docker/docker-compose.yml` reads those credentials and exposes Postgres on `localhost:5432`
- `make migrate` runs `go run -tags=debug ./cmd/ migrate:up` inside a `golang:1.25-alpine` container on the same Docker network as the database

### Default credentials

```
POSTGRES_DB=detect_price
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
```

### Verification commands

```bash
# Start the full stack, including pgweb
make docker-up

# Open a psql shell inside the database container
make db-shell

# Tail database logs
make db-logs

# Launch the pgweb UI
make pgweb-up   # browse http://localhost:8081

# Manually inspect tables (psql one-liner)
docker compose -p detect-price -f docker/docker-compose.yml exec db \
  psql -U postgres -d detect_price -c '\dt'
```

### Web UI (pgweb)

The stack now ships with [pgweb](https://github.com/sosedoff/pgweb). After running `make pgweb-up`, open  
`http://localhost:8081` in your browser—the connection is preconfigured for you.

### External tools

If you prefer to connect with DBeaver, TablePlus, or similar, use  
`postgresql://postgres:postgres@localhost:5432/detect_price`.

