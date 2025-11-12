# Detect Price by Photo Monorepo

A unified foundation for building the Detect Price by Photo SaaS platform with Go backend, React frontend, and AI-powered price estimation.

This workspace contains:
- `apps/backend`: Go modular backend (PostgreSQL + Redis + Midtrans + OpenRouter integrations)
- `apps/frontend`: Vite + React SPA for sellers, premium subscribers, and admins
- `packages/`: Shared database assets, types, utilities, and AI datasets

Powered by [Moonrepo](https://moonrepo.dev/docs/install) for task orchestration and **pnpm** for dependency management.

**Source:** Forked from [zero-one-group/monorepo](https://github.com/zero-one-group/monorepo).

**Documentation:** Full documentation is available in the [docsite](/docsite/) directory. To view locally, run `hugo server` from the project root.

## Prerequisites

- **Node.js** 20.18.0+ (managed by Moonrepo)
- **pnpm** 10.18.0+ (managed by Moonrepo)
- **Go** 1.25+ (for backend)
- **Docker** & **Docker Compose** (for infrastructure)

## Getting Started

1. **Install Moonrepo**
   ```sh
   # Using npm
   npm install -g @moonrepo/cli
   
   # Or using pnpm
   pnpm add -g @moonrepo/cli
   
   # Or using Homebrew (macOS/Linux)
   brew install moonrepo
   ```

2. **Setup Moonrepo toolchain**
   ```sh
   pnpm moon:setup
   ```

3. **Install dependencies**
   ```sh
   pnpm install
   ```

4. **Configure environment variables**
   ```sh
   # Copy example file and fill in your API keys
   cp docker/.env.example docker/.env.dev
   # Edit docker/.env.dev with your actual credentials
   # REQUIRED: Set MIDTRANS_SERVER_KEY, MIDTRANS_CLIENT_KEY, JWT_SECRET_KEY
   ```

5. **Sync Moonrepo projects**
   ```sh
   pnpm moon:sync
   ```

6. **Start the backend**
   ```sh
   pnpm dev:backend
   # Or using Moonrepo directly:
   moon run backend:dev
   ```

7. **Start the web app**
   ```sh
   pnpm dev:frontend
   # Or using Moonrepo directly:
   moon run frontend:dev
   ```

8. **Bring up infrastructure**
   ```sh
   pnpm compose:up
   ```

Detailed product requirements, API documentation, and technical guides live in the [docsite](/docsite/) directory. View the documentation locally using Hugo or access it via the generated static site.

## Moonrepo Commands

Moonrepo provides a unified task runner across all projects. Common commands:

### Development
```sh
# Run development servers
moon run backend:dev      # Start backend server
moon run frontend:dev    # Start frontend dev server

# Run all dev tasks
moon run :dev
```

### Building
```sh
# Build specific project
moon run backend:build   # Build backend binary
moon run frontend:build  # Build frontend for production

# Build all projects
moon run :build
```

### Testing
```sh
# Run tests
moon run backend:test    # Run backend Go tests
moon run frontend:test   # Run frontend tests

# Run all tests
moon run :test
```

### Code Quality
```sh
# Lint all projects
moon run :lint

# Format all code
moon run :format

# Type check TypeScript projects
moon run :typecheck
```

### Database Operations
```sh
# Run migrations
moon run backend:migrate:up

# Rollback migrations
moon run backend:migrate:down

# Seed database
moon run backend:seed
```

### Project Management
```sh
# Check project configuration
moon check

# Sync project dependencies
moon sync

# View project graph
moon project-graph

# View task graph
moon task-graph
```

## Tooling Cheat Sheet

### Using pnpm scripts (wraps Moonrepo)
- `pnpm dev:backend` – start backend development server
- `pnpm dev:frontend` – start frontend development server
- `pnpm test:backend` – run backend tests
- `pnpm test:frontend` – run frontend tests
- `pnpm build:all` – build all projects
- `pnpm compose:up` – start Docker services
- `pnpm compose:down` – stop Docker services

### Using Moonrepo directly
- `moon run backend:migrate:up` – run database migrations
- `moon run frontend:test` – execute frontend unit tests
- `moon run shared-types:build` – build shared TypeScript contracts
- `moon run :lint` – lint all projects
- `moon run :format` – format all code

### Legacy commands (still work)
- `go run apps/backend/cmd/ -- migrate:up` – run database migrations
- `pnpm --filter frontend test` – execute frontend unit tests

## Project Structure

```
.
├── apps/
│   ├── backend/          # Go backend application
│   │   └── moon.yml      # Moonrepo config
│   └── frontend/         # React frontend application
│       └── moon.yml      # Moonrepo config
├── packages/
│   ├── shared-types/     # Shared TypeScript types
│   │   └── moon.yml
│   ├── shared-utils/     # Shared utilities
│   │   └── moon.yml
│   └── shared-db/        # Shared database migrations
├── .moon/
│   ├── workspace.yml     # Moonrepo workspace config
│   ├── toolchain.yml     # Moonrepo toolchain config
│   └── tasks.yml         # Moonrepo global tasks
├── compose.yaml          # Root Docker Compose file
├── pnpm-workspace.yaml   # PNPM workspace config
└── package.json          # Root package.json with scripts
```

## Benefits of Moonrepo

- **Unified Task Runner**: Single command interface for all projects
- **Dependency Graph**: Automatically handles project dependencies
- **Caching**: Intelligent caching of task outputs
- **Parallel Execution**: Runs independent tasks in parallel
- **Cross-Platform**: Works consistently across Windows, macOS, and Linux
- **Type Safety**: TypeScript project references automatically synced

Follow `.cursor/rules` for tech-debt guardrails and vertical development guidelines.
