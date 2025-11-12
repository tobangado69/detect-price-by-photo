# Migration to Zero One Group Monorepo Structure

This document summarizes the migration of the Detect Price by Photo monorepo to align with Zero One Group monorepo conventions.

## Changes Made

### 1. Moonrepo Configuration

**Created:**
- `.moon/workspace.yml` - Workspace configuration with project discovery
- `.moon/toolchain.yml` - Toolchain settings (Node.js 22.19, pnpm 10.18.0)
- `.moon/tasks.yml` - Global task definitions

**Removed:**
- `moon.yml` (root) - Replaced by `.moon/workspace.yml`

**Updated:**
- `apps/backend/moon.yml` - Removed `language: 'go'` (not supported)
- `apps/frontend/moon.yml` - Kept as-is
- `packages/*/moon.yml` - Kept as-is

### 2. Docker Compose Structure

**Created:**
- `compose.yaml` (root) - Main compose file using includes pattern
- Updated `docker/_stacks_/postgres.yaml` - Adapted for our database setup
- Updated `docker/_stacks_/redis.yaml` - Simplified authentication
- Updated `docker/_stacks_/mailpit.yaml` - Changed to MailHog (our current setup)

**Updated:**
- `Makefile` - Changed `COMPOSE_FILE` to `compose.yaml`
- All documentation references updated to use `compose.yaml`

### 3. Package.json Scripts

**Updated scripts to match template:**
- `prepare` - Runs `moon setup`
- `lint` - Uses `moon :lint`
- `check` - Uses `moon :check`
- `format` - Uses Biome directly
- `compose:*` - Updated to use `compose.yaml`
- `update-deps` - Includes `moon :update-deps`
- Added cleanup scripts

**Kept for convenience:**
- `dev:backend`, `dev:frontend` - Wrapper scripts
- `test:backend`, `test:frontend` - Wrapper scripts
- `build:*` - Wrapper scripts

### 4. Documentation Migration

**Created Hugo docsite structure:**
- `hugo.yaml` - Hugo configuration
- `docsite/content/` - Migrated all docs with frontmatter
- `docsite/data/sidebar.yaml` - Navigation structure

**Migrated documents:**
- API Documentation → `docsite/content/backend/api-documentation.md`
- Forgot Password Guide → `docsite/content/backend/forgot-password.md`
- Database Setup → `docsite/content/backend/database-setup.md`
- Admin Seed → `docsite/content/backend/admin-seed.md`
- Midtrans Setup → `docsite/content/integration/midtrans-setup.md`
- PRDs → `docsite/content/prd/`
- Technical Docs → `docsite/content/technical/`

**Created new docs:**
- `docsite/content/_overview.md` - Project overview
- `docsite/content/backend/project-structure.md` - Backend structure
- `docsite/content/backend/configuration.md` - Configuration guide
- `docsite/content/backend/faq.md` - FAQ
- `docsite/content/backend/troubleshooting.md` - Troubleshooting guide

**Updated:**
- All API documentation to clarify RBAC (admin-only user endpoints)
- All port references (9871 → 8080)
- All Docker Compose references (docker/docker-compose.yml → compose.yaml)
- All container name references

### 5. README Updates

- Updated to reference docsite
- Aligned with Zero One Group template style
- Updated project structure diagram
- Added Moonrepo benefits section

## Key Differences from Template

1. **Backend/Frontend Services**: Template doesn't include backend/frontend in compose.yaml (infrastructure only). We kept our Makefile for building/running apps.

2. **MailHog vs Mailpit**: We use MailHog, template uses Mailpit. Updated mailpit.yaml stack to use MailHog.

3. **PostgreSQL Image**: We use `pgvector/pgvector:pg16` for vector support, template uses `postgres:18-alpine`.

4. **Service Names**: We use `db` instead of `pgsql` to match existing code.

## Migration Checklist

- [x] Create `.moon/workspace.yml`, `.moon/toolchain.yml`, `.moon/tasks.yml`
- [x] Create root `compose.yaml` with includes
- [x] Update `docker/_stacks_/*` files
- [x] Update `package.json` scripts
- [x] Remove root `moon.yml`
- [x] Create Hugo docsite structure
- [x] Migrate all documentation
- [x] Update API docs for RBAC
- [x] Update README
- [x] Update Makefile
- [x] Fix all port and path references

## Next Steps

1. **Test the setup:**
   ```bash
   pnpm prepare
   pnpm install
   pnpm moon:sync
   pnpm moon:check
   ```

2. **Start infrastructure:**
   ```bash
   pnpm compose:up
   ```

3. **Verify services:**
   - Database: `docker compose -f compose.yaml ps db`
   - Redis: `docker compose -f compose.yaml ps redis`
   - MailHog: `docker compose -f compose.yaml ps mailhog`

4. **View documentation:**
   ```bash
   # Install Hugo if not already installed
   # Then run:
   hugo server
   # Visit http://localhost:1313
   ```

## Breaking Changes

1. **Docker Compose**: Path changed from `docker/docker-compose.yml` to `compose.yaml`
2. **Project Name**: Compose project name changed from `detect-price` to `detect-price-by-photo`
3. **User Endpoints**: All `/api/v1/users/*` endpoints now require admin role (was already implemented, just documented)

## Backward Compatibility

- Makefile still works (updated to use new paths)
- All pnpm scripts still work
- Moonrepo tasks are backward compatible
- Old `docker/docker-compose.yml` can coexist (but not recommended)

## Notes

- The old `docs/` directory still exists for reference but new content should go in `docsite/`
- Hugo themes need to be set up separately if you want custom styling
- The docsite can be built statically: `hugo` (outputs to `docsite/public/`)

