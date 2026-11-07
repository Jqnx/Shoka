import { fileURLToPath } from 'node:url';
import Database from 'better-sqlite3';
import { drizzle } from 'drizzle-orm/better-sqlite3';
import { migrate } from 'drizzle-orm/better-sqlite3/migrator';

// Standalone so the container can migrate before the server boots, without
// shipping drizzle-kit (and its dev dependency tree) in the runtime image.

const url = process.env.DATABASE_URL;
if (!url) {
	console.error('DATABASE_URL is not set');
	process.exit(1);
}

const sqlite = new Database(url);
sqlite.pragma('journal_mode = WAL');

try {
	migrate(drizzle(sqlite), {
		migrationsFolder: fileURLToPath(new URL('../drizzle', import.meta.url))
	});
} finally {
	sqlite.close();
}
