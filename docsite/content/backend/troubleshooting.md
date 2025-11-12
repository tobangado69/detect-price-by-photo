---
title: Troubleshooting
weight: 8
---

# Troubleshooting Guide

## Backend Won't Start

### Database Connection Error

**Symptoms:** `failed to connect to database` or `connection refused`

**Solutions:**
1. Verify database is running: `docker compose -f compose.yaml ps`
2. Check database logs: `docker compose -f compose.yaml logs db`
3. Verify `DATABASE_URL` in environment variables
4. Ensure database is healthy: `docker compose -f compose.yaml exec db pg_isready`

### Port Already in Use

**Symptoms:** `bind: address already in use`

**Solutions:**
1. Find process using port: `lsof -i :8080` (macOS/Linux) or `netstat -ano | findstr :8080` (Windows)
2. Stop the conflicting service
3. Or change the port in configuration

## Authentication Issues

### 403 Forbidden on User Endpoints

**Cause:** User endpoints require admin role.

**Solution:** Use an admin account token, or use `/api/v1/auth/signup` for public registration.

### Invalid Token Error

**Symptoms:** `invalid token` or `token expired`

**Solutions:**
1. Check token hasn't expired (24 hours for access tokens)
2. Verify `JWT_SECRET_KEY` matches between token generation and validation
3. Ensure token is sent in `Authorization: Bearer <token>` header format

## Database Issues

### Migration Fails

**Symptoms:** `migration failed` or `relation already exists`

**Solutions:**
1. Check migration status: Query `goose_db_version` table
2. Rollback last migration: `moon run backend:migrate:down`
3. Check for conflicting migrations
4. Verify database connection

### Seed Data Not Appearing

**Solutions:**
1. Verify seed ran successfully: Check logs
2. Check database directly: `make db-shell`
3. Re-run seed: `moon run backend:seed --force`

## Email Issues

### Emails Not Appearing in MailHog

**Solutions:**
1. Verify MailHog is running: `docker compose -f compose.yaml ps mailhog`
2. Check backend logs for email sending errors
3. Verify SMTP configuration in environment variables
4. Check MailHog UI: `http://localhost:8025`

### SMTP Connection Failed

**Solutions:**
1. Verify `SMTP_HOST` and `SMTP_PORT` are correct
2. For Docker: Use service name (`mailhog`) not `localhost`
3. Check network connectivity: `docker compose -f compose.yaml exec backend ping mailhog`

## Moonrepo Issues

### Tasks Not Found

**Symptoms:** `task not found` or `project not found`

**Solutions:**
1. Sync projects: `moon sync`
2. Check project configuration: `moon check`
3. Verify `moon.yml` exists in project directory

### Cache Issues

**Symptoms:** Tasks using stale cache

**Solutions:**
1. Clear cache: `pnpm cleanup:cache`
2. Clear specific project cache: `moon cache clean backend`
3. Run task without cache: `moon run backend:test --no-cache`

## Docker Issues

### Services Won't Start

**Solutions:**
1. Check logs: `docker compose -f compose.yaml logs`
2. Verify environment file exists: `docker/.env.dev`
3. Check port conflicts
4. Restart Docker daemon

### Network Issues

**Symptoms:** Services can't communicate

**Solutions:**
1. Verify network exists: `docker network ls | grep detect_price`
2. Recreate network: `docker compose -f compose.yaml down` then `up`
3. Check service names match in configuration

## Performance Issues

### Slow API Responses

**Solutions:**
1. Check database query performance
2. Verify Redis cache is working
3. Check for N+1 queries
4. Review backend logs for slow operations

### High Memory Usage

**Solutions:**
1. Check container limits in compose files
2. Review application memory usage
3. Check for memory leaks in long-running processes

