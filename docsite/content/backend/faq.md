---
title: FAQ
weight: 7
---

# Frequently Asked Questions

## General

### How do I start the development environment?

1. Start infrastructure: `pnpm compose:up`
2. Run migrations: `moon run backend:migrate:up`
3. Start backend: `pnpm dev:backend`
4. Start frontend: `pnpm dev:frontend`

### What ports are used?

- Backend API: `http://localhost:8080`
- Frontend: `http://localhost:5173`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- MailHog UI: `http://localhost:8025`
- MailHog SMTP: `localhost:1025`
- pgweb: `http://localhost:8081`

## Authentication

### How do I get an admin token?

1. Sign in as admin: `POST /api/v1/auth/signin/email` with `admin@detectprice.com` / `admin123`
2. Use the `access_token` from the response

### Why can't I access `/api/v1/users` endpoints?

All `/api/v1/users/*` endpoints require admin role. Regular users receive `403 Forbidden`. Use `/api/v1/auth/signup` for public registration.

## Database

### How do I reset the database?

```bash
pnpm compose:reset
moon run backend:migrate:up
moon run backend:seed
```

### How do I view the database?

Use pgweb: `http://localhost:8081` (after starting services with `pnpm compose:up`)

## Email

### Why am I not receiving emails?

1. Check MailHog is running: `docker compose -f compose.yaml ps`
2. Check MailHog UI: `http://localhost:8025`
3. Verify backend logs show MailHog connection

### How do I test email in production?

Update environment variables to use a real SMTP service (SendGrid, AWS SES, etc.)

## Moonrepo

### How do I see all available tasks?

```bash
moon run --help
```

### How do I run tasks for a specific project?

```bash
moon run backend:dev    # Backend tasks
moon run frontend:dev   # Frontend tasks
```

### How do I check project configuration?

```bash
moon check
```

