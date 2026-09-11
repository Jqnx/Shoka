import { error } from '@sveltejs/kit';
import fs from 'node:fs/promises';
import type { RequestHandler } from './$types';
import { avatarPath } from '$lib/server/avatars';
import { AVATAR_FILENAME_RE } from '$lib/schemas/avatar';

// Self-hosted library - avatars aren't world-readable. A same-origin <img>
// sends cookies automatically, so this is transparent to the browser.
export const GET: RequestHandler = async ({ params, locals }) => {
	if (!locals.user) error(401, 'Unauthorized');

	// The regex - not the DB - is the path-traversal defense here: it must run
	// before the filename ever touches the filesystem.
	if (!AVATAR_FILENAME_RE.test(params.file)) error(400, 'Invalid avatar filename');

	let bytes: Buffer;
	try {
		bytes = await fs.readFile(avatarPath(params.file));
	} catch (err) {
		if ((err as NodeJS.ErrnoException).code === 'ENOENT') error(404, 'Not found');
		throw err;
	}

	return new Response(new Uint8Array(bytes), {
		headers: {
			'Content-Type': 'image/webp',
			'X-Content-Type-Options': 'nosniff',
			// Safe because the filename is a fresh UUID every upload - replacing
			// an avatar naturally busts the cache instead of needing one.
			'Cache-Control': 'private, max-age=31536000, immutable'
		}
	});
};
