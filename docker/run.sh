#!/bin/bash
# Supervises the two Shoka processes. If either dies the whole container exits,
# so the restart policy — not a half-dead container — decides what happens next.
set -euo pipefail

log() { echo "[shoka] $*"; }

# The Go backend applies its own goose migrations on startup; this covers the
# frontend's Drizzle/Better-Auth tables in the same database file.
log "applying frontend migrations"
node /app/frontend/scripts/migrate.js

# cwd must be /app: the backend resolves its data directory as ./data.
cd /app
log "starting backend on 127.0.0.1:8080"
HOST=127.0.0.1 PORT=8080 /app/backend/shoka &
backend=$!

cd /app/frontend
log "starting frontend on 0.0.0.0:3000"
HOST=0.0.0.0 PORT=3000 node build &
frontend=$!

shutdown() {
	trap - TERM INT
	kill -TERM "$backend" "$frontend" 2>/dev/null || true
	wait || true
}
trap shutdown TERM INT

status=0
wait -n || status=$?

log "a service exited with status $status; stopping the container"
shutdown
exit "$status"
