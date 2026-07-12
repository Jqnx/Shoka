# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shoka is a self-hosted doujinshi library manager. It consists of a Go backend REST API and a SvelteKit frontend, with SQLite for storage and Meilisearch for full-text search.

## Common Commands

### Development

```bash
make dev              # Start full stack (backend + frontend + Meilisearch)
make dev/backend      # Backend only with Air hot-reload
make dev/frontend     # Frontend dev server
make dev/services     # Background services only (Meilisearch, FlareResolver)
```

### Database

```bash
make migrate/up                  # Apply pending migrations
make migrate/down                # Rollback last migration
make migrate/status              # Show migration status
make migrate/create NAME=<name>  # Create new SQL migration
make sqlc                        # Regenerate Go code from SQL queries (after editing .sql files)
```

### Docs & Deps

```bash
make swag   # Regenerate Swagger API docs (after editing handler annotations)
make tidy   # Go module cleanup
```

## Architecture

### Backend (`/backend`)

- **Router**: `chi/v5` — routes defined in `internal/api/routes.go`
- **Handlers**: `internal/api/handlers/` — one file per domain (archive, metadata, admin)
- **Database**: SQLite3 with WAL mode via `sqlc` for type-safe queries
  - SQL queries live in `internal/database/queries/*.sql`
  - Generated Go code in `internal/database/sqlc/` — never edit these files directly
  - Migrations in `internal/database/migrations/` managed by Goose
  - Database files in `data/` - shared with frontend
- **Job system**: `internal/jobs/` — async workers (10 concurrent) for scan, cover generation, thumbnails, and metadata extraction
- **Image processing**: `govips/v2` for fast image manipulation; cache in `../cache/`
- **Library scanner**: `internal/library/` — watches the filesystem, handles ZIP/7Z/RAR/PDF archives
  - Avoid the `gen2brain/go-fitz` go package - it causes issues with vips
- **Metadata pipeline**: `internal/metadata/` — pluggable sources (ComicInfo XML, filename parsing, e-Hentai, nHentai)
- **Config**: `viper` loading from `data/config.yaml` with env overrides (`HOST`, `PORT`, `LOG_LEVEL`)

### Frontend (`/frontend`)

- **Framework**: SvelteKit 2 + Svelte 5 with TypeScript
- **Database**: SQLite3 with WAL mode - shared with backend
- **Auth**: Better-Auth with Drizzle ORM
  - Prefer JWT auth over sessions
- **Styling**: Tailwind CSS 4 and shadcn-svelte
- **API calls**: frontend calls backend at `localhost:8080` (configurable)

### External Services

- **Meilisearch**: full-text search for archives (run via Docker)
- **FlareResolver**: Cloudflare bypass for scraping metadata sources (run via Docker)

## Key Workflows

### Adding a new API endpoint

1. Add the SQL query to the relevant `internal/database/queries/*.sql` file
2. Run `make sqlc` to regenerate Go bindings
3. Add the handler in `internal/api/handlers/`
4. Register the route in `internal/api/routes.go`
5. Add Swagger annotations and run `make swag`

### Adding a database migration

1. `make migrate/create NAME=<descriptive_name>` — creates a new Goose SQL file
2. Edit the generated file in `internal/database/migrations/`
3. `make migrate/up` to apply

### sqlc pattern

Queries use named parameters (`:param`) and are annotated with `-- name: FuncName :one/:many/:exec`. After editing `.sql` files always run `make sqlc`.
Use named parameters (`sqlc.arg('param')`) for the sql limit and offset values.

### Frontend Authentication

Prefer using server-side actions over a client-side client.

### Smoke-test the dev server

The dev server should be running at `localhost:5173` by default.
Check if it is running first before starting the dev server.
