# Shoka

[![CI](https://github.com/Jqnx/Shoka/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Jqnx/Shoka/actions/workflows/ci.yml)
[![License: GPL-3.0](https://img.shields.io/badge/license-GPL--3.0-blue)](LICENSE)

Shoka is a self-hosted home for your doujinshi and voice-work collection.

## Features

- **Comic libraries.** Add folders of `.cbz`, `.cbr`, `.zip`, `.rar`, or `.7z`
  archives.
- **Audio libraries.** Each top-level folder becomes one work with its tracks
  and gallery images.
- **Browsing and search.** Move through the collection by artist, circle,
  parody, character, tag, or voice actor, or search titles, summaries, and names.
- **Reader.** Single- or double-page view, left-to-right, right-to-left, or
  continuous scrolling, adjustable image fit and background, slideshow, and
  reading progress.
- **Favorite pages.** Star individual pages and revisit them in a full-screen
  viewer.
- **Edit metadata.** Change titles, summaries, dates, categories, and relations
  in the UI, and save the result back into the archive as ComicInfo.xml.
- **External sources.** Configure metadata sources per
  library, and choose the order they run in. See [SOURCES.md](SOURCES.md) for
  which sources are supported.
- **Duplicates.** Compare covers and sampled pages to flag likely duplicates.
- **Themes.** Pick from 16 built-in themes or create your own in the theme
  editor.
- **Multi-user.** Separate accounts with an admin role, per-user library access,
  avatars, and personal reader settings.
- **API.** A REST API for the library, with Swagger docs at `/docs` when the
  backend port is reachable (e.g. `localhost:8080` in development; the web UI
  does not proxy it).

## Quick start

### Docker Compose (recommended)

Requirements: Docker with Compose v2 and a directory holding your collection.
Save this as `compose.yaml`:

```yaml
name: shoka

services:
  shoka:
    image: ghcr.io/jqnx/shoka:latest
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      ORIGIN: http://localhost:3000
      BETTER_AUTH_SECRET: "<32+ character secret>"
      PUID: "1000"
      PGID: "1000"
      TZ: UTC
    volumes:
      - shoka-data:/app/data
      - shoka-cache:/cache
      - /path/to/library:/library

volumes:
  shoka-data:
  shoka-cache:
```

> **Warning:** `/library` is mounted read-write, and Shoka does modify it:
> saving metadata injects `ComicInfo.xml` into archives, and for 7z and RAR
> archives it rewrites the container as a `.cbz` with the same base name,
> replacing the original. Keep your own backups if that isn't what you want.
> `/app/data` (the `shoka-data` volume) holds the database and config — back it
> up too.

Generate the secret once with `openssl rand -base64 32` (32+ characters) and
paste it into `BETTER_AUTH_SECRET`; changing it later signs everyone out. Set
`ORIGIN` to the URL you actually browse — behind a reverse proxy, the public
address — or form submissions fail as CSRF errors. Set `PUID`/`PGID` to the
owner of your collection (`id -u`, `id -g`) and `/path/to/library` to that
collection. Then run, from the directory holding the file:

```bash
docker compose up -d      # pulls the image and starts the container
docker compose logs -f    # follow the logs
docker compose down       # stop the stack (volumes are kept)
```

### Docker

Requirements: Docker and a directory holding your collection.

```bash
docker run -d --name shoka \
  -p 3000:3000 \
  -e BETTER_AUTH_SECRET='<the secret generated above>' \
  -e ORIGIN=http://localhost:3000 \
  -e PUID="$(id -u)" -e PGID="$(id -g)" \
  -v shoka_shoka-data:/app/data \
  -v shoka_shoka-cache:/cache \
  -v /path/to/library:/library \
  ghcr.io/jqnx/shoka:latest
```

### First start

Open the web UI at your chosen URL, e.g `http://localhost:3000`. On first start everything
redirects to the setup wizard:

1. **Account** — the first account owns the instance and becomes its
   administrator. Later sign-ins use the username chosen here.
2. **Library** — name it, choose a category, and enter the
   collection path as seen inside the container (`/library/...`), not on the
   host.
3. **First scan** — Shoka waits for the initial scan and reports the result.

After the wizard, covers and metadata for new archives are generated in the
background — follow them under **Admin → Jobs**. Files added to the library
later are picked up by the watcher, and more libraries can be added at
**Admin → Libraries**.

## Configuration

`docker-compose.yml` reads these from your environment or a `.env` file. We
recommend keeping them in `.env`, or using docker secrets:

| Variable             | Default                 | Purpose                                                  |
| -------------------- | ----------------------- | -------------------------------------------------------- |
| `BETTER_AUTH_SECRET` | — (required)            | Session signing secret, 32+ characters.                  |
| `ORIGIN`             | `http://localhost:3000` | Public URL, used for CSRF and cookies.                   |
| `SHOKA_PORT`         | `3000`                  | Host port for the web UI.                                |
| `PUID` / `PGID`      | `1000`                  | UID/GID the container runs as and chowns the volumes to. |
| `TZ`                 | `UTC`                   | Container timezone.                                      |
| `LOG_LEVEL`          | `info`                  | `panic`/`fatal`/`error`/`warn`/`info`/`debug`/`trace`.   |

The library path is not an environment variable: it is the `/library` bind
mount in your Compose file. Point the volume's host side at your collection
(the quick-start snippet uses `/path/to/library`), then enter the container
path `/library` in the setup wizard.

Settings not exposed as environment variables live in `config.yaml` inside the
data volume. The keys it accepts are:

```yaml
cache:
  dir: /cache # extracted pages and generated images; dev default is ../cache
  lru_size: 128 # reserved
images:
  retention_period: 14 # reserved
metadata:
  flaresolverr:
    url: "" # reserved
```

To edit the config file bind-mount it into the data volume. Create `config.yaml` next to your Compose
file with the keys you want (the example above works as-is), then add one line
to the service's volume list:

```yaml
volumes:
  - shoka-data:/app/data
  - shoka-cache:/cache
  - /path/to/library:/library
  - ./config.yaml:/app/data/config.yaml
```

Recreate the stack once with `docker compose up -d` so the new mount takes
effect; after that, edits to the host file only need
`docker compose restart shoka`. For a plain `docker run`, pass
`-v "$PWD/config.yaml:/app/data/config.yaml"` when creating the container, and
restart it after editing.

## Upgrading

If you set `SHOKA_LIBRARY` in `.env`, earlier versions read the library's host
path from it. That variable is gone and the repo's Compose file ignores it, so
before pulling, replace the host side of the `/library` volume in
`docker-compose.yml` (currently `./content`) with that path. Otherwise the
recreated container mounts `./content`, the startup scan sees an empty library,
and deletes its catalog entries — reading progress, favorites, and metadata
included. Files on disk are not touched. A Compose file of your own that still
interpolates `${SHOKA_LIBRARY}` keeps working as before.

With Docker Compose:

```bash
docker compose pull
docker compose up -d
```

If you run the container directly instead of through Compose, pull the new tag
and recreate it with the same `docker run` flags shown in the quick start:

```bash
docker pull ghcr.io/jqnx/shoka:latest
docker rm -f shoka
```

## Development

Prerequisites: Go 1.27.x with CGO and `libvips-dev`, Node 24 (`.nvmrc`), Docker
(for FlareSolverr and for regenerating the frontend lockfile), plus the CLIs the
Makefile shells out to: `air`, `golangci-lint` v2.14, `sqlc`, `swag`, and
`goose`.

```bash
make dev              # backend + frontend + FlareSolverr
make dev/backend      # backend only, with Air hot-reload
make dev/frontend     # frontend dev server
make dev/services     # background services only
make help             # every target with a description
```

Frontend environment lives in `frontend/.env` (template:
`frontend/.env.example`); the backend creates `backend/data/config.yaml` with
defaults on first start and reads `HOST`, `PORT`, and `LOG_LEVEL` from the
environment. Point `DATABASE_URL` at the backend's database
(`../backend/data/shoka.db`, resolved from `frontend/`) so both halves share one
file; start the backend once first so `backend/data` exists. The backend applies
its Goose migrations on startup, but the frontend's Drizzle tables (users,
sessions, themes, instance settings) are not created by `make dev` — apply them
once, and again after changing `db/schema.ts`:

```bash
cd frontend
npm run db:push
```

`db:push` loads `frontend/.env` itself, so no separate export is needed.

The API listens on `localhost:8080` by default, the dev server on
`localhost:5173`, and the generated Swagger UI is at
`http://localhost:8080/docs` while developing.

Database and generated code:

```bash
make migrate/up                  # apply pending Go migrations
make migrate/down                # roll back the last one
make migrate/status              # show migration status
make migrate/create NAME=<name>  # new Goose SQL migration
make sqlc                        # regenerate Go code after editing queries/*.sql
make swag                        # regenerate Swagger docs after handler edits
make frontend/lock               # regenerate frontend/package-lock.json
```

Checks:

```bash
cd backend && go test -tags sqlite_fts5 ./...
make lint                        # golangci-lint + svelte-check
cd frontend && npm run lint      # prettier --check + eslint
```

> Day-to-day frontend work uses `bun.lock`. The container's build stage installs
> with `bun install --frozen-lockfile`, while its runtime dependencies come from
> `npm ci` against `frontend/package-lock.json`, because `better-sqlite3` is
> ABI-locked to a Node major. The two lockfiles must move together —
> regenerate the npm one with `make frontend/lock` whenever
> `frontend/package.json` changes, or the image build fails.
>
> Every compiled Go binary that touches the database needs the `sqlite_fts5`
> build tag, including a locally installed `goose`:
> `go install -tags sqlite_fts5 github.com/pressly/goose/v3/cmd/goose@latest`.
> Without it, FTS5 queries fail at runtime. `sqlc generate` is the one exception
> — it parses SQL statically.

## Security

Shoka stores and serves a private library; see [SECURITY.md](SECURITY.md) for
the reporting process and scope. Do not open a public issue for a suspected
vulnerability.

## License

GPL-3.0 — see [LICENSE](LICENSE).
