---
title: Configuration
weight: 6
---

# Backend Configuration

## Environment Variables

⚠️ **SECURITY WARNING:** All API keys and secrets must be set via environment variables. Never hardcode credentials in source code or commit `.env` files to version control.

The backend reads configuration from environment variables. For local development, copy `docker/.env.example` to `docker/.env.dev` and fill in your actual values.

### Quick Setup

```bash
# Copy example file
cp docker/.env.example docker/.env.dev

# Edit with your actual credentials
# Use your preferred editor to fill in API keys and secrets
```

### Required Environment Variables

#### Database

```bash
DATABASE_URL=postgresql://postgres:postgres@db:5432/detect_price?sslmode=disable
```

#### Redis Cache

```bash
REDIS_URL=redis://localhost:6379
```

#### JWT Authentication

```bash
# REQUIRED: Generate a secure random secret (minimum 32 characters)
# Example: openssl rand -base64 32
JWT_SECRET_KEY=your-secret-key-here
```

#### Midtrans Payment Gateway

```bash
# REQUIRED: Get credentials from https://dashboard.midtrans.com/
MIDTRANS_SERVER_KEY=your-server-key-here
MIDTRANS_CLIENT_KEY=your-client-key-here
MIDTRANS_ENV=sandbox                # sandbox or production
MIDTRANS_MERCHANT_ID=your-merchant-id-here
```

### Optional Environment Variables

#### SMTP (Email)

```bash
SMTP_HOST=mailhog          # Docker: mailhog, Production: smtp.provider.com
SMTP_PORT=1025             # MailHog: 1025, Production: 587 or 465
SMTP_USERNAME=             # Optional, required for authenticated SMTP
SMTP_PASSWORD=             # Optional, required for authenticated SMTP
SMTP_SENDER_NAME=Detect Price by Photo
SMTP_SENDER_EMAIL=noreply@detectprice.com
```

#### Application

```bash
APP_BASE_URL=http://localhost:5173  # Frontend URL
APP_PORT=8000                        # Backend port
APP_ENV=development                  # development, staging, production
```

#### OpenRouter (AI)

```bash
OPENROUTER_API_KEY=your-api-key-here     # Optional, for AI features
```

### Complete Configuration Reference

See `docker/.env.example` for all available environment variables with descriptions and default values.

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

