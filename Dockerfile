# syntax=docker/dockerfile:1

# Shoka ships as a single container: the Go API bound to 127.0.0.1:8080 and the
# SvelteKit server on 0.0.0.0:3000. The SvelteKit server proxies /api/* to the
# backend in-process (frontend/src/hooks.server.ts), so only 3000 is published
# and the API is never directly reachable from outside the container.

ARG GO_VERSION=1.26
ARG NODE_VERSION=24
ARG BUN_VERSION=1.4

# ---------------------------------------------------------------------------
# Backend — CGO build against libvips.
# ---------------------------------------------------------------------------
FROM golang:${GO_VERSION}-trixie AS backend-build

RUN apt-get update \
 && apt-get install -y --no-install-recommends libvips-dev \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY backend/ ./

# sqlite_fts5 is mandatory. Without it go-sqlite3 still compiles, but every
# archive_fts query fails at runtime with "no such module: fts5".
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go build -tags sqlite_fts5 -trimpath -ldflags='-s -w' -o /out/shoka .

# ---------------------------------------------------------------------------
# Shared Node base. The deps and runtime stages have to agree on the Node major
# — better-sqlite3 is a V8-ABI addon — so they resolve it in one place.
# ---------------------------------------------------------------------------
FROM node:${NODE_VERSION}-trixie-slim AS node-base

# ---------------------------------------------------------------------------
# Frontend — bun honours bun.lock, which pins the toolchain that emits build/.
# ---------------------------------------------------------------------------
FROM oven/bun:${BUN_VERSION}-debian AS bun

FROM node-base AS frontend-build

COPY --from=bun /usr/local/bin/bun /usr/local/bin/bun

WORKDIR /app

COPY frontend/package.json frontend/bun.lock frontend/.npmrc ./
RUN --mount=type=cache,target=/root/.bun/install/cache bun install --frozen-lockfile

COPY frontend/ ./

# SvelteKit's postbuild analyse step imports the server modules, so the build
# dies on db/index.ts's DATABASE_URL check unless something is set. These are
# scoped to this RUN and never reach the runtime image; the app reads the real
# values from $env/dynamic/private at startup.
RUN DATABASE_URL=/tmp/build-placeholder.db \
    BETTER_AUTH_SECRET=buildtimeplaceholderbuildtimeplaceholder \
    ORIGIN=http://localhost:3000 \
    bun run build

# ---------------------------------------------------------------------------
# Runtime node_modules — better-sqlite3 is a V8-ABI addon (no N-API), so it has
# to be installed under the same Node major that will run the server. bun.lock
# can't be used here for that reason, so the runtime `dependencies` are pinned
# by package-lock.json and installed with `npm ci`; regenerate it with
# `make frontend/lock` after touching package.json, or this build fails.
# The import check fails the build if a native binary is missing, rather than
# letting the container crash on first request.
# ---------------------------------------------------------------------------
FROM node-base AS frontend-deps

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --omit=dev --no-audit --no-fund \
 && npm cache clean --force \
 && node --input-type=module -e "import 'better-sqlite3'; import 'sharp'; console.log('native modules ok')"

# ---------------------------------------------------------------------------
# Runtime
# ---------------------------------------------------------------------------
FROM node-base AS runtime

# libvips42t64 is the Debian trixie runtime package matching the libvips-dev the
# backend was linked against. gosu drops privileges after the entrypoint has
# remapped PUID/PGID; tini reaps the two supervised children.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      ca-certificates \
      gosu \
      libvips42t64 \
      tini \
      tzdata \
 && rm -rf /var/lib/apt/lists/*

# Reuse the base image's uid/gid 1000 rather than adding a second account.
RUN groupmod -n shoka node \
 && usermod -l shoka -d /app node \
 && rm -rf /home/node

WORKDIR /app

COPY --from=backend-build /out/shoka                 /app/backend/shoka
COPY --from=frontend-deps /app/node_modules          /app/frontend/node_modules
COPY --from=frontend-build /app/build                /app/frontend/build
COPY --from=frontend-build /app/drizzle              /app/frontend/drizzle
COPY --from=frontend-build /app/scripts/migrate.js   /app/frontend/scripts/migrate.js
COPY frontend/package.json                           /app/frontend/package.json
COPY docker/entrypoint.sh                            /usr/local/bin/shoka-entrypoint
COPY docker/run.sh                                   /usr/local/bin/shoka-run

RUN chmod +x /usr/local/bin/shoka-entrypoint /usr/local/bin/shoka-run

# DATABASE_URL must stay in step with the backend's own data dir; the entrypoint
# asserts it. It also anchors the avatar directory (frontend/src/lib/server/
# avatars.ts derives it from dirname), which is why it belongs inside /app/data.
#
# BODY_SIZE_LIMIT needs headroom over MAX_UPLOAD_BYTES (lib/schemas/avatar.ts):
# multipart framing makes the request body bigger than the file it carries, and
# at equal limits adapter-node returns a bare 413 before the upload action can
# report "Image is too large.". Avatars are the only upload in the app.
ENV NODE_ENV=production \
    PUID=1000 \
    PGID=1000 \
    DATABASE_URL=/app/data/shoka.db \
    ORIGIN=http://localhost:3000 \
    BODY_SIZE_LIMIT=3M \
    LOG_LEVEL=info

# Keeps the database and the generated covers/thumbnails out of the container's
# writable layer even when run without -v. /library is deliberately excluded:
# it is user-supplied input rather than app state, and an anonymous volume there
# would look like an empty collection instead of a missing mount.
VOLUME ["/app/data", "/cache"]

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=40s --retries=3 \
    CMD node --input-type=module -e "\
const urls = ['http://127.0.0.1:8080/health', 'http://127.0.0.1:3000/'];\
await Promise.all(urls.map(async (u) => { const r = await fetch(u, { redirect: 'manual' }); if (r.status >= 500) throw new Error(u + ' -> ' + r.status); }));"

ENTRYPOINT ["/usr/bin/tini", "--", "/usr/local/bin/shoka-entrypoint"]
CMD ["/usr/local/bin/shoka-run"]
