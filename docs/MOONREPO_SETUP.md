# Moonrepo Setup Guide

This document explains how Moonrepo is configured and used in the Detect Price by Photo monorepo.

## What is Moonrepo?

Moonrepo is a build system and task runner for monorepos. It provides:

- **Unified task execution** across all projects
- **Intelligent caching** to speed up builds
- **Dependency graph** management
- **Parallel task execution**
- **Cross-platform support**

## Installation

### Global Installation

```sh
# Using npm
npm install -g @moonrepo/cli

# Using pnpm
pnpm add -g @moonrepo/cli

# Using Homebrew (macOS/Linux)
brew install moonrepo

# Using Scoop (Windows)
scoop install moonrepo
```

### Verify Installation

```sh
moon --version
```

## Configuration Files

### `.moon/toolchain.yml`

Defines the toolchain configuration for the entire workspace:

- Node.js version (20.18.0)
- pnpm version (10.18.0)
- TypeScript version (5.8.3)
- Plugin configurations

### `moon.yml` (Root)

Workspace-level configuration:

- Project discovery patterns (`apps/*`, `packages/*`)
- Global task definitions
- Dependency manager settings
- Codegen configuration

### Project `moon.yml` Files

Each project (`apps/backend`, `apps/frontend`, `packages/*`) has its own `moon.yml`:

- Project type (application/library)
- Language (go/typescript)
- Task definitions specific to that project
- Dependencies on other projects

## Common Tasks

### Development

```sh
# Run backend dev server
moon run backend:dev

# Run frontend dev server
moon run frontend:dev

# Run all dev tasks
moon run :dev
```

### Building

```sh
# Build specific project
moon run backend:build
moon run frontend:build

# Build all projects
moon run :build
```

### Testing

```sh
# Run tests for specific project
moon run backend:test
moon run frontend:test

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

# Reset database
moon run backend:migrate:reset

# Seed database
moon run backend:seed
```

## Project Dependencies

Moonrepo automatically handles project dependencies. For example:

- `frontend` depends on `shared-types` and `shared-utils`
- When building `frontend`, Moonrepo will build dependencies first

View the dependency graph:

```sh
moon project-graph
```

## Task Dependencies

Tasks can depend on other tasks. For example:

- `frontend:build` depends on `shared-types:build` and `shared-utils:build`
- Moonrepo will run dependencies first

View the task graph:

```sh
moon task-graph
```

## Caching

Moonrepo caches task outputs based on:

- Input files (hashed)
- Environment variables
- Task command

If inputs haven't changed, Moonrepo will use cached outputs.

### Cache Management

```sh
# View cache statistics
moon cache

# Clear cache
moon cache clean

# Clear cache for specific project
moon cache clean backend
```

## Parallel Execution

Moonrepo runs independent tasks in parallel by default. For example:

```sh
moon run :test
```

This will run tests for `backend` and `frontend` simultaneously if they don't depend on each other.

## Integration with pnpm

Moonrepo works alongside pnpm:

- **pnpm** manages dependencies (`node_modules`, lock files)
- **Moonrepo** manages task execution and caching

Both tools complement each other:

```sh
# Install dependencies (pnpm)
pnpm install

# Run tasks (Moonrepo)
moon run backend:dev
```

## Troubleshooting

### Setup Issues

If Moonrepo isn't recognizing projects:

```sh
# Sync projects
moon sync

# Check configuration
moon check
```

### Task Not Found

If a task isn't found:

1. Check the project's `moon.yml` file
2. Verify the task name matches exactly
3. Run `moon check` to validate configuration

### Cache Issues

If tasks aren't using cache:

1. Check that outputs are defined in `moon.yml`
2. Verify input files haven't changed
3. Clear cache and rebuild: `moon cache clean && moon run :build`

## Migration from pnpm scripts

We've migrated from direct pnpm scripts to Moonrepo tasks:

### Before
```sh
pnpm dev:backend          # cd apps/backend && go run ./cmd/ -- serve
pnpm test:backend        # cd apps/backend && go test ./...
```

### After
```sh
moon run backend:dev     # Same command, but cached and parallelized
moon run backend:test    # Same command, but cached and parallelized
```

The `package.json` scripts now wrap Moonrepo commands for backward compatibility.

## Best Practices

1. **Define inputs and outputs** for all tasks to enable caching
2. **Use tags** for shared packages (`shared-types`, `shared-utils`)
3. **Set dependencies** explicitly in `moon.yml`
4. **Run `moon check`** before committing changes
5. **Use `moon sync`** after pulling changes

## Resources

- [Moonrepo Documentation](https://moonrepo.dev/docs)
- [Moonrepo GitHub](https://github.com/moonrepo/moon)
- [Moonrepo Discord](https://discord.gg/moonrepo)

