// Removes avatar files that no `user.image` row references, plus any
// `*.webp.tmp` write-in-progress artifacts left by a crash between write and
// rename (see writeAvatar() in src/lib/server/avatars.ts).
//
// Deliberately standalone: it opens the sqlite file directly with
// better-sqlite3 rather than importing $lib/server/db, so it can run without
// the SvelteKit Vite plugin providing $env/$app virtual modules. Run via
// plain `node` (see the `avatars:gc` package.json script), not `bun run` -
// better-sqlite3's native binding does not load under Bun's own JS engine
// ("'better-sqlite3' is not yet supported in Bun", oven-sh/bun#4290), which
// is exactly why the rest of this app's DB-touching code always runs under
// real Node too (`npm run dev`, `node build/index.js` via adapter-node) even
// though bun is the package manager. AVATAR_FILENAME_RE is still imported
// from $lib/schemas/avatar (via a relative path, since $lib itself is a
// SvelteKit-only alias) - that module has no SvelteKit-only imports, so
// there's no drift risk in keeping that one constant shared.
import path from 'node:path';
import fs from 'node:fs/promises';
import Database from 'better-sqlite3';
import { AVATAR_FILENAME_RE } from '../src/lib/schemas/avatar.ts';

const ONE_HOUR_MS = 60 * 60 * 1000;
const TMP_SUFFIX_RE = /\.webp\.tmp$/;

function resolveDbUrl(): string {
	const url = process.env.DATABASE_URL;
	if (!url) throw new Error('DATABASE_URL is not set.');
	return url;
}

function resolveAvatarDir(dbUrl: string): string {
	return process.env.AVATAR_DIR ?? path.join(path.dirname(dbUrl), 'avatars');
}

function referencedFilenames(dbUrl: string): Set<string> {
	const db = new Database(dbUrl, { readonly: true });
	try {
		const rows = db.prepare('SELECT image FROM user WHERE image IS NOT NULL').all() as {
			image: string;
		}[];
		return new Set(
			rows
				.map((r) => r.image.replace(/^\/avatars\//, ''))
				.filter((filename) => AVATAR_FILENAME_RE.test(filename))
		);
	} finally {
		db.close();
	}
}

/** Old enough that it can't be racing an in-flight upload. */
async function isStale(fullPath: string): Promise<boolean> {
	const stat = await fs.stat(fullPath);
	return Date.now() - stat.mtimeMs >= ONE_HOUR_MS;
}

async function main() {
	const dbUrl = resolveDbUrl();
	const dir = resolveAvatarDir(dbUrl);
	const referenced = referencedFilenames(dbUrl);

	let entries: string[];
	try {
		entries = await fs.readdir(dir);
	} catch (err) {
		if ((err as NodeJS.ErrnoException).code === 'ENOENT') {
			console.log(`No avatar directory at ${dir} - nothing to do.`);
			return;
		}
		throw err;
	}

	let removed = 0;
	let skipped = 0;

	for (const entry of entries) {
		const isCandidateAvatar = AVATAR_FILENAME_RE.test(entry);
		const isTmpArtifact = TMP_SUFFIX_RE.test(entry);
		if (!isCandidateAvatar && !isTmpArtifact) continue;
		if (isCandidateAvatar && referenced.has(entry)) continue;

		const full = path.join(dir, entry);
		if (!(await isStale(full))) {
			skipped += 1;
			continue;
		}

		await fs.unlink(full);
		removed += 1;
		console.log(`Removed orphaned file: ${entry}`);
	}

	console.log(
		`Done. Removed ${removed}, skipped ${skipped} (too recent to be safe), ${referenced.size} referenced.`
	);
}

main().catch((err) => {
	console.error(err);
	process.exit(1);
});
