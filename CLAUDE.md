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
make lint   # golangci-lint plus the frontend type check (npm run check)
```

`make lint` and `make swag` both shell out to a CLI that parses the stdlib. After a Go
toolchain upgrade, reinstall whichever one breaks rather than chasing its `--version`:
an older build aborts in stdlib typechecking (`method must have no type parameters` on
`math/rand/v2`) before it ever reaches project code, which looks like a repo error but
isn't. Rebuild with the current toolchain, e.g.
`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`.

## Architecture

### Backend (`/backend`)

- **Router**: `chi/v5` — routes defined in `internal/api/routes.go`
- **Handlers**: `internal/api/handlers/` — one file per domain (archive, metadata, admin, …), plus a focused file wherever a shape is genuinely shared: `taxonomy.go` (characters, parodies, tags and artists list/page/delete the same way — the per-entity file then holds little more than its swagger blocks), `duplicates.go`, `feeds.go`, `library_create.go`, `library_sources.go`, `archive_user_state.go`. `characters_parodies.go` holds both entities because they are identical apart from the name.
- **Database**: SQLite3 with WAL mode via `sqlc` for type-safe queries
  - SQL queries live in `internal/database/queries/*.sql`
  - Generated Go code in `internal/database/sqlc/` — never edit these files directly
  - Migrations in `internal/database/migrations/` managed by Goose
  - Database files in `data/` - shared with frontend
- **Search**: SQLite FTS5 with the `trigram` tokenizer (`archive_fts` virtual table, see `internal/database/migrations/`) — substring/fuzzy matching over title, summary, and denormalized artist/tag/parody/circle/character names, kept in sync automatically via SQL triggers (on `archive` and each `archive_*` join table) rather than an application-level reindex step. No external search service.
  - **Requires the `sqlite_fts5` Go build tag** — the `mattn/go-sqlite3` driver only compiles in FTS5 support when built with `-tags sqlite_fts5`. This applies to every *compiled/run* Go binary that touches this database: `go build`, `go vet`, `go test`, Air (`backend/.air.toml`'s `cmd` already includes it), and any locally-installed `goose` CLI used for `make migrate/*` — install/reinstall it with `go install -tags sqlite_fts5 github.com/pressly/goose/v3/cmd/goose@latest`, otherwise `make migrate/up` fails with "no such module: fts5" the moment it hits the FTS5 migration. Omitting the tag doesn't fail the build itself — it silently produces a binary that only errors at runtime, the first time it touches `archive_fts`, so don't skip it. `sqlc generate` is the one exception: it parses `.sql` files statically and doesn't need the tag.
- **Job system**: `internal/jobs/` — async workers (10 concurrent) for scan, cover generation, thumbnails, metadata extraction, and perceptual hashing
- **Duplicate detection**: `phash` job hashes each archive at four sample points (cover + 25/50/75% through the pages) via `govips` → `goimagehash`, stored on `archive.phash_p0/p25/p50/p75`. `GET /api/admin/duplicates` groups archives where a majority of comparable points fall within a Hamming-distance threshold; `POST /api/admin/phashes` backfills/re-hashes the library.
- **Image processing**: `govips/v2` for fast image manipulation; cache in `../cache/`
- **Library scanner**: `internal/library/` — watches the filesystem, handles ZIP/7Z/RAR/PDF archives
  - Avoid the `gen2brain/go-fitz` go package - it causes issues with vips
  - Two independent triggers per library, both configured on the `library` row and editable at `/admin/libraries/[id]`: the fsnotify watcher (`watch_enabled`) and a periodic scan (`scan_interval_minutes`, 0 = off, otherwise >= 15). The timer exists because inotify events don't arrive over network shares or some container mounts. `Manager.OnLibraryChanged` reconciles the watcher on save, so both settings take effect without a restart.
  - `Manager.runScheduler` (`scheduler.go`) polls `ListLibrariesDueForScan` every minute and enqueues via `EnqueueOnce`. Due-ness is measured from `library.last_scanned_at`, which `ScanLibrary` sets after every scan *attempt* — including failures, so a library that can't be scanned doesn't get re-enqueued on every tick.
  - Watchers hang off `Manager.baseCtx` (captured in `Start`), never a request context — deriving one from `r.Context()` kills the watcher when the HTTP response is written.
- **Scheduled maintenance**: `internal/maintenance/` runs a second minute-tick scheduler over the `maintenance_task` table (`cover` and `phash` rows, seeded by migration 00038). `Manager.RunTask` fans one job out per archive and stamps `last_run_at` at enqueue (`MarkMaintenanceTaskRun`), after every attempt — manual runs on `/admin/maintenance` go through the same path, so both count as "last run". A due task is skipped while any job of its type is pending/running (`HasPendingJobByType`, type-wide — unlike the scan scheduler's per-library `EnqueueOnce`), because one run is a whole-library pass rather than a per-target job.
- **Metadata pipeline**: `internal/metadata/` — pluggable sources (ComicInfo XML, filename parsing, e-Hentai, nHentai)
- **Authorization (RBAC)**: two roles, `user` (default) and `admin`. `authMiddleware` resolves the session's role/banned flag from the shared `user` table and puts the role on the request context; `requirePermission(auth.ArchiveDelete)` in `internal/api/middleware.go` gates each mutating route in `routes.go`. Reads are ungated for any authenticated user, and per-user data (progress, favourites, ratings, reader settings) is scoped by `auth.UserIDFromContext` rather than by role.
  - The argument is a typed `auth.Permission` declared in `backend/internal/auth/rbac.go`, so a misspelled resource/action pair is a compile error rather than a route nobody — admins included — can reach.
  - The role matrix lives in **two** places and must be kept in sync: `backend/internal/auth/rbac.go` and `frontend/src/lib/auth/permissions.ts` (the canonical better-auth `createAccessControl` definition). The frontend copy only hides UI and guards loads/actions — the Go copy is what actually enforces, because `hooks.server.ts` proxies `/api/*` straight through without consulting `locals.user`, and several components call the API directly from the browser.
  - `internal/auth/rbac_test.go` parses the `statement` literal out of `permissions.ts` and diffs it against the Go statement set. The `user`/`session` resources are excluded: the frontend takes them from better-auth's `...defaultStatements`, which the parser can't resolve, so the Go side owns those values.
  - `requirePermission` is **not** how the frontend page actions guard themselves — they call `guard(...)` from `$lib/server/guard.ts`, which returns a uniform `fail(403, { action, error })`. Loads use `can(...)` plus `error(403, …)` because they have no error payload.
  - The `role`/`banned` columns are **frontend-owned** (better-auth `admin` plugin, Drizzle migration `frontend/drizzle/0003_*`). Go only reads them — do not add a Goose migration for them.
- **Config**: `viper` loading from `data/config.yaml` with env overrides (`HOST`, `PORT`, `LOG_LEVEL`)

### Frontend (`/frontend`)

- **Framework**: SvelteKit 2 + Svelte 5 with TypeScript
- **Database**: SQLite3 with WAL mode - shared with backend
- **Auth**: Better-Auth with Drizzle ORM
  - Always use the following url when searching for better-auth documentation: `https://better-auth.com/llms.txt`.
  - You have access to the better-auth MCP server. Prefer using it over the `llms.txt` url.
  - Prefer JWT auth over sessions
  - The `admin` plugin provides roles/bans. `src/lib/auth/permissions.ts` holds the canonical statement set plus `can(user, …)` / `canAccessAdmin(user)` — it lives outside `$lib/server/` on purpose so components can hide controls with it. Mirrored in `backend/internal/auth/rbac.go` (see the backend RBAC note above). `roles.user` grants nothing but must stay in the map: better-auth validates `setRole`/`createUser` against the map's keys, so dropping it would make demoting an admin to `user` fail.
  - The first account created becomes `admin` via a root-level `databaseHooks.user.create.before` in `auth.ts`; every later signup gets `user`. Signup stays publicly open.
  - A load guard does **not** protect a form action — guard both. Admin routes re-check their own permission in each `+page.server.ts` load *and* in every mutating action; the action guards go through `guard(...)` from `$lib/server/guard.ts` so every 403 has the same `{ action, error }` shape.
  - The action guards are a deliberate second layer, not the enforcement: every guarded action ends at a `/api/*` route (gated in `routes.go`) or a better-auth endpoint (gated by its own `hasPermission`), both of which re-check the same permission. What the guard buys is a local rejection with a page-specific message — it is not what stops a forged POST, and it does not need to stay in step with the Go matrix at the route level. Keep it that way: if a new action guards a permission the API does not enforce, that is a bug in the API route, not a reason to rely on the page.
- **Styling**: Tailwind CSS 4 and shadcn-svelte
  - Use the following url when searching for shadcn-svelte documentation: `https://www.shadcn-svelte.com/llms.txt`.
- **Theming**: per-user, frontend-only (the Go backend never reads it). Built-in
  palettes live in `src/lib/themes/presets/*`; the canonical token list is
  `src/lib/themes/tokens.ts` (31 colours + `--radius`, driven off one array).
  `hooks.server.ts` (`handleTheme`) resolves the active theme via a direct
  Drizzle read (`src/lib/server/themes.ts`) and injects an `html:root { … }`
  `<style>` block before the first byte (no flash); the client switches without
  a reload via `src/lib/themes/theme.svelte.ts`. Tables `user_theme` /
  `user_preferences` are in `db/schema.ts` (NOT `auth.schema.ts`). Token values
  are injected into a `<style>` block, so `src/lib/schemas/theme.ts` allowlists
  the exact value shape - treat it as a security control. `theme_id` is a plain
  string (`rose-pine-moon` or `custom:<uuid>`); an unresolvable id falls back to
  `shoka-dark`. `layout.css` keeps `:root`/`.dark` as the degraded-path fallback.
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

## Available Svelte MCP Tools

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

### Linting & formatting

After writing or editing any Svelte code, run `prettier --check` and `eslint` (from `/frontend`) on the changed files and fix any issues before considering the work done.

## Go

You have access to a range of skills related to Go.

### golang-how-to

Orchestrator for using the correct and relevant skills. ALWAYS use this when writing Go code.

## Coding Style

### Comments

Use clear and concise comments. Try to keep comments as short as possible, but don't sacrifice clarity for brevity. Only add the complex logic or important details that are not immediately obvious from the code. Swagger comments do not fall under this category.
