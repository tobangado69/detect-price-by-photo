# Detect Price by Photo Monorepo

This workspace contains the Detect Price by Photo vertical slice:
- `apps/backend`: Go modular backend (PostgreSQL + Redis + Midtrans + OpenRouter integrations)
- `apps/frontend`: Vite + React SPA for sellers, premium subscribers, and admins
- `packages/`: Shared database assets, types, utilities, and AI datasets

Managed with `pnpm` for dependency orchestration across apps and shared packages.

## Getting Started

1. **Install dependencies**
   ```sh
   pnpm install
   ```
2. **Start the backend**
   ```sh
   pnpm dev:backend
   ```
3. **Start the web app**
   ```sh
   pnpm dev:frontend
   ```
4. **Bring up infrastructure**
   ```sh
   pnpm compose:up
   ```

Detailed product requirements and implementation notes live in `/docs`.

## Tooling Cheat Sheet

- `go run apps/backend/cmd/ -- migrate:up` – run database migrations
- `pnpm --filter frontend test` – execute frontend unit tests
- `pnpm --filter @detect-price/shared-types build` – bundle shared TypeScript contracts
- `pnpm compose:cleanup` – tear down local services

Follow `.cursor/rules` for tech-debt guardrails and vertical development guidelines.
