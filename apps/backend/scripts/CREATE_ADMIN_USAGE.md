# Running create_admin.go Script

## Problem
The script needs to connect to the Docker database, but it uses default config values that don't match.

## Solution Options

### Option 1: Set DATABASE_URL Environment Variable (Recommended)

**Windows (PowerShell):**
```powershell
$env:DATABASE_URL="postgresql://postgres:postgres@localhost:5432/detect_price?sslmode=disable"
go run scripts/create_admin.go -email admin@detectprice.com -password admin123 -role admin
```

**Windows (CMD):**
```cmd
set DATABASE_URL=postgresql://postgres:postgres@localhost:5432/detect_price?sslmode=disable
go run scripts/create_admin.go -email admin@detectprice.com -password admin123 -role admin
```

**Linux/Mac:**
```bash
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/detect_price?sslmode=disable"
go run scripts/create_admin.go -email admin@detectprice.com -password admin123 -role admin
```

### Option 2: Run Inside Docker Container

```bash
docker compose -p detect-price -f docker/docker-compose.yml exec backend \
  go run scripts/create_admin.go -email admin@detectprice.com -password admin123 -role admin
```

### Option 3: Create .env File in apps/backend

Create `apps/backend/.env` with:
```env
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/detect_price?sslmode=disable
```

Then run:
```bash
cd apps/backend
go run scripts/create_admin.go -email admin@detectprice.com -password admin123 -role admin
```

## Database Connection Details

- **Host:** localhost (when running from host machine)
- **Port:** 5432 (Docker port mapping)
- **Database:** detect_price
- **User:** postgres
- **Password:** postgres (from docker/.env.dev)

## Verify Connection

Before running the script, verify the database is accessible:
```bash
docker compose -p detect-price -f docker/docker-compose.yml exec db psql -U postgres -d detect_price -c "SELECT version();"
```

