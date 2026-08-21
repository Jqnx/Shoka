# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shoka is a self-hosted doujinshi library manager. It consists of a Go backend REST API and a SvelteKit frontend, with SQLite for storage and search (FTS5 virtual tables using the trigram tokenizer — no external search service).

## Common Commands

### Development

```bash
make dev              # Start full stack (backend + frontend + FlareResolver)
make dev/backend      # Backend only with Air hot-reload
make dev/frontend     # Frontend dev server
make dev/services     # Background services only (FlareResolver)
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
- **Search**: SQLite FTS5 with the `trigram` tokenizer (`archive_fts` virtual table, see `internal/database/migrations/`) — substring/fuzzy matching over title, summary, and denormalized artist/tag/parody/circle/character names, kept in sync automatically via SQL triggers (on `archive` and each `archive_*` join table) rather than an application-level reindex step. No external search service.
  - **Requires the `sqlite_fts5` Go build tag** — the `mattn/go-sqlite3` driver only compiles in FTS5 support when built with `-tags sqlite_fts5`. This applies to every *compiled/run* Go binary that touches this database: `go build`, `go vet`, `go test`, Air (`backend/.air.toml`'s `cmd` already includes it), and any locally-installed `goose` CLI used for `make migrate/*` — install/reinstall it with `go install -tags sqlite_fts5 github.com/pressly/goose/v3/cmd/goose@latest`, otherwise `make migrate/up` fails with "no such module: fts5" the moment it hits the FTS5 migration. Omitting the tag doesn't fail the build itself — it silently produces a binary that only errors at runtime, the first time it touches `archive_fts`, so don't skip it. `sqlc generate` is the one exception: it parses `.sql` files statically and doesn't need the tag.
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

## Svelte
You are able to use the Svelte MCP server, where you have access to comprehensive Svelte 5 and SvelteKit documentation. Here's how to use the available tools effectively:

## Available Svelte MCP Tools:

### 1. list-sections

Use this FIRST to discover all available documentation sections. Returns a structured list with titles, use_cases, and paths.
When asked about Svelte or SvelteKit topics, ALWAYS use this tool at the start of the chat to find relevant sections.

### 2. get-documentation

Retrieves full documentation content for specific sections. Accepts single or multiple sections.
After calling the list-sections tool, you MUST analyze the returned documentation sections (especially the use_cases field) and then use the get-documentation tool to fetch ALL documentation sections that are relevant for the user's task.

### 3. svelte-autofixer

Analyzes Svelte code and returns issues and suggestions.
You MUST use this tool whenever writing Svelte code before sending it to the user. Keep calling it until no issues or suggestions are returned.

### 4. playground-link

Generates a Svelte Playground link with the provided code.
After completing the code, ask the user if they want a playground link. Only call this tool after user confirmation and NEVER if code was written to files in their project.
