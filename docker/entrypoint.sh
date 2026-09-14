#!/bin/bash
# Runs as root only long enough to align the service account with the host's
# uid/gid and prepare the mounted volumes, then drops privileges.
set -euo pipefail

# Fixed in-container locations. Both are declared as VOLUMEs and mounted by
# compose — relocate storage by moving the mount, not by changing these.
data_dir=/app/data
cache_dir=/cache

log() { echo "[shoka] $*"; }
die() {
	echo "[shoka] FATAL: $*" >&2
	exit 1
}

if [ -z "${BETTER_AUTH_SECRET:-}" ]; then
	die "BETTER_AUTH_SECRET is not set. Generate one with: openssl rand -base64 32"
fi

if [ "${#BETTER_AUTH_SECRET}" -lt 32 ]; then
	die "BETTER_AUTH_SECRET must be at least 32 characters (got ${#BETTER_AUTH_SECRET})."
fi

# The two services reach the same SQLite file by unrelated means: the backend
# joins config.DataDir with database.DatabaseFile, the frontend opens
# DATABASE_URL. Nothing links them, so assert it here rather than let a mismatch
# quietly run two separate databases.
if [ "${DATABASE_URL:-}" != "$data_dir/shoka.db" ]; then
	die "DATABASE_URL must be $data_dir/shoka.db (got '${DATABASE_URL:-unset}'), or the frontend and backend open different databases."
fi

# Remapping the account, rather than chowning the library, is what lets a
# bind-mounted collection stay owned by the host user that manages it.
if [ "$(id -u shoka)" != "$PUID" ] || [ "$(id -g shoka)" != "$PGID" ]; then
	groupmod -o -g "$PGID" shoka
	usermod -o -u "$PUID" -g "$PGID" shoka
fi

for dir in "$data_dir" "$cache_dir"; do
	mkdir -p "$dir"
	# Recurse only when the directory itself is misowned — on an established
	# cache a blanket chown -R costs minutes of startup for nothing.
	if [ "$(stat -c %u "$dir")" != "$PUID" ] || [ "$(stat -c %g "$dir")" != "$PGID" ]; then
		log "taking ownership of $dir"
		chown -R "$PUID:$PGID" "$dir"
	fi
done

# The backend reads cache.dir from data/config.yaml and has no env override for
# it; seeding the file points it at the cache volume instead of the default
# ../cache, which would land outside every mount.
config="$data_dir/config.yaml"
if [ ! -s "$config" ]; then
	log "writing default config to $config"
	cat >"$config" <<-EOF
		cache:
		    dir: $cache_dir
		    lru_size: 128
		images:
		    retention_period: 14
	EOF
	chown "$PUID:$PGID" "$config"
fi

log "starting as uid=$PUID gid=$PGID"
exec gosu shoka "$@"
