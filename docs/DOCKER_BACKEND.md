# Running Backend in Docker

## Quick Start

### Start Backend Service
```bash
make backend-up
```

### Start Full Stack (Backend + Frontend + DB + Cache)
```bash
make docker-up
```

### View Backend Logs
```bash
docker logs -f detect-price-backend-1
```

### Stop Backend
```bash
make docker-down
# or stop only backend
docker compose -p detect-price -f docker/docker-compose.yml stop backend
```

## Backend Service Details

- **Container Name:** `detect-price-backend-1`
- **Port Mapping:** `8080:8000` (host:container)
- **Health Check:** `http://localhost:8080/healthz`
- **API Base URL:** `http://localhost:8080/api/v1`

## Environment Variables

The backend reads from `docker/.env.dev`:
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `OPENROUTER_API_KEY` - AI service API key (optional)
- `MIDTRANS_SERVER_KEY` - Payment gateway key (optional)

## Troubleshooting

### Backend not starting?
1. Check logs: `docker logs detect-price-backend-1`
2. Verify database is running: `docker ps | grep db`
3. Check environment variables: `cat docker/.env.dev`

### Backend exits immediately?
- Check if database is accessible from container
- Verify `DATABASE_URL` is correct
- Check logs for connection errors

### Rebuild backend image:
```bash
make backend-build
make backend-up
```

## Testing Backend

### Health Check
```bash
curl http://localhost:8080/healthz
```

### API Endpoints
```bash
# Get OpenAPI spec
curl http://localhost:8080/api/openapi.json

# Sign in (if you have an account)
curl -X POST http://localhost:8080/api/v1/auth/signin/email \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@detectprice.com", "password": "admin123"}'
```

## Development Workflow

1. **Start infrastructure:**
   ```bash
   make db-up
   ```

2. **Run migrations:**
   ```bash
   make migrate
   ```

3. **Start backend:**
   ```bash
   make backend-build
   make backend-up
   ```

4. **View logs:**
   ```bash
   docker logs -f detect-price-backend-1
   ```

5. **Stop everything:**
   ```bash
   make docker-down
   ```

