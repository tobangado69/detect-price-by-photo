---
title: Overview
weight: 1
---

# Detect Price by Photo Monorepo

This workspace contains the Detect Price by Photo vertical slice:
- `apps/backend`: Go modular backend (PostgreSQL + Redis + Midtrans + OpenRouter integrations)
- `apps/frontend`: Vite + React SPA for sellers, premium subscribers, and admins
- `packages/`: Shared database assets, types, utilities, and AI datasets

Managed with **Moonrepo** for task orchestration and **pnpm** for dependency management across apps and shared packages.

## Quick Start

1. **Install Moonrepo**
   ```sh
   npm install -g @moonrepo/cli
   # or
   pnpm add -g @moonrepo/cli
   ```

2. **Setup Moonrepo toolchain**
   ```sh
   pnpm prepare
   ```

3. **Install dependencies**
   ```sh
   pnpm install
   ```

4. **Sync Moonrepo projects**
   ```sh
   pnpm moon:sync
   ```

5. **Start infrastructure**
   ```sh
   pnpm compose:up
   ```

6. **Start development**
   ```sh
   pnpm dev:backend    # Start backend server
   pnpm dev:frontend   # Start frontend dev server
   ```

## Project Structure

```
.
├── apps/
│   ├── backend/          # Go backend application
│   └── frontend/         # React frontend application
├── packages/
│   ├── shared-types/     # Shared TypeScript types
│   ├── shared-utils/     # Shared utilities
│   └── shared-db/        # Shared database migrations
├── docker/
│   └── _stacks_/         # Docker Compose service stacks
├── docsite/              # Hugo documentation site
├── .moon/                # Moonrepo configuration
├── compose.yaml          # Root Docker Compose file
└── package.json          # Root package.json
```

## Key Features

- **RBAC**: Strict role-based access control (admin-only user management endpoints)
- **Email Verification**: Complete email verification flow with MailHog/Mailpit support
- **Password Reset**: Secure password reset with token-based flow
- **User Suspension**: Admin can suspend/ban users with expiration dates
- **Payment Integration**: Midtrans payment gateway integration
- **Photo Analysis**: AI-powered price estimation using OpenRouter

## Documentation

- [API Documentation](/backend/api-documentation/)
- [Backend PRD](/prd/backend-prd/)
- [Frontend PRD](/prd/frontend-prd/)
- [Technical Architecture](/technical/backend/)

