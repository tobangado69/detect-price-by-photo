---
title: Configuration
weight: 6
---

# Backend Configuration

## Environment Variables

The backend reads configuration from environment variables. For local development, these are set in `docker/.env.dev`.

### Database

```bash
DATABASE_URL=postgresql://postgres:postgres@db:5432/detect_price?sslmode=disable
```

### Redis Cache

```bash
REDIS_URL=redis://localhost:6379
```

### SMTP (Email)

```bash
SMTP_HOST=mailhog          # Docker: mailhog, Production: smtp.provider.com
SMTP_PORT=1025             # MailHog: 1025, Production: 587 or 465
SMTP_USERNAME=             # Optional
SMTP_PASSWORD=             # Optional
SMTP_SENDER_NAME=Detect Price by Photo
SMTP_SENDER_EMAIL=noreply@detectprice.com
```

### JWT

```bash
JWT_SECRET_KEY=your-secret-key-here  # Must be at least 32 characters
```

### Application

```bash
APP_BASE_URL=http://localhost:5173  # Frontend URL
APP_PORT=8000                        # Backend port
APP_ENV=development                  # development, staging, production
```

### OpenRouter (AI)

```bash
OPENROUTER_API_KEY=your-api-key     # Optional, for AI features
```

### Midtrans (Payments)

```bash
MIDTRANS_SERVER_KEY=your-server-key
MIDTRANS_CLIENT_KEY=your-client-key
MIDTRANS_ENV=sandbox                # sandbox or production
MIDTRANS_MERCHANT_ID=your-merchant-id
```

## Configuration Files

### `apps/backend/config/config.go`

Central configuration management with validation and defaults.

### `docker/.env.dev`

Local development environment variables (not committed to git).

## Moonrepo Configuration

Backend tasks are defined in `apps/backend/moon.yml`:

- `dev` - Run development server
- `test` - Run tests
- `build` - Build binary
- `migrate:up` - Run migrations
- `migrate:down` - Rollback migrations
- `seed` - Seed database

## Docker Configuration

See [Database Setup](/backend/database-setup/) for Docker Compose configuration.

