# Shoka

Self-hosted doujinshi library manager. A Go REST API and a SvelteKit frontend
over SQLite, with full-text search handled in-database by FTS5 — there is no
external search service to run.

## Deployment

Shoka ships as a single container running both the API and the web server.
Only the web port is published; the API is reachable only from inside the
container.

### Requirements

- Docker with Compose v2
- A directory holding your collection

### Setup

```bash
cp .env.example .env
openssl rand -base64 32   # paste into BETTER_AUTH_SECRET
```

Two values need attention before the first start:

- **`BETTER_AUTH_SECRET`** — required, 32+ characters. The container refuses to
  start without it. Changing it later invalidates every existing session.
- **`ORIGIN`** — the URL you actually browse to, including scheme and port.
  Behind a reverse proxy this is the public address (`https://shoka.example.com`),
  not the container's. If it doesn't match, pages load but every form
  submission fails as a CSRF error.

Set `PUID`/`PGID` to the host user that owns your collection (`id -u`, `id -g`)
and point `SHOKA_LIBRARY` at it, then:

```bash
make run     # docker compose up -d
make logs    # tail both processes
make stop    # docker compose down
```

The UI is on `http://localhost:3000` by default. The first account you register
becomes the administrator.

### Storage layout

| Path in container | Holds                                         | Compose default               |
| ----------------- | --------------------------------------------- | ----------------------------- |
| `/app/data`       | `shoka.db`, `config.yaml`, avatars            | `shoka-data` volume           |
| `/cache`          | Generated covers, thumbnails, extracted pages | `shoka-cache` volume          |
| `/library`        | Your collection                               | `${SHOKA_LIBRARY}` bind mount |

The cache volume is disposable — deleting it costs regeneration time, nothing
more. `/app/data` is the one that matters; back it up.

**The library mount is read-write, and Shoka does modify it**: it injects
`ComicInfo.xml` into archives and rewrites 7z/rar archives as zip, replacing the
original. Keep a backup if that isn't what you want.

Add libraries from **Admin → Libraries** in the UI, using the in-container path
(`/library/...`), not the host path.

### Configuration

Settings not exposed as environment variables live in `config.yaml` inside the
data volume. Edit it with:

```bash
docker compose exec shoka vi /app/data/config.yaml
docker compose restart shoka
```

### Upgrading

```bash
git pull
docker compose build
docker compose up -d
```

Both migration sets — Goose for the API's tables, Drizzle for the auth and
theme tables — run automatically at startup. Volumes survive
`docker compose down`; only `docker compose down -v` destroys them.

> Migrating an existing development database into the container is not a
> supported path. Dev environments use `drizzle-kit push`, which leaves no
> migration history, so the startup migrator will try to recreate tables that
> already exist and fail. Start from an empty data volume.

## Development

See `AGENTS.md` for architecture notes and the full workflow. Common commands:

```bash
make dev              # backend + frontend + FlareSolverr
make dev/backend      # backend only, with Air hot-reload
make dev/frontend     # frontend dev server
make dev/services     # background services only

make migrate/up       # apply pending backend migrations
make migrate/status   # migration status
make sqlc             # regenerate Go code from SQL queries
make swag             # regenerate Swagger docs
make frontend/lock    # regenerate frontend/package-lock.json

make lint             # golangci-lint
make fmt              # format Go code
```

Run `make help` for the full list.

> Day-to-day frontend work uses `bun.lock`. The container installs its runtime
> dependencies with `npm ci` instead, because `better-sqlite3` is ABI-locked to
> a Node major, so `frontend/package-lock.json` has to be regenerated with
> `make frontend/lock` whenever `frontend/package.json` changes — otherwise the
> image build fails on the out-of-sync lockfile.
>
> Every compiled Go binary that touches the database needs the `sqlite_fts5`
> build tag, including a locally installed `goose`:
> `go install -tags sqlite_fts5 github.com/pressly/goose/v3/cmd/goose@latest`.
> Without it, FTS5 queries fail at runtime.
