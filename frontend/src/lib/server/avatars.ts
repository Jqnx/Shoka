import path from 'node:path';
import fs from 'node:fs/promises';
import sharp from 'sharp';
import { env } from '$env/dynamic/private';
import { AVATAR_FILENAME_RE } from '$lib/schemas/avatar';

const AVATAR_SIZE = 512;

// Formats sharp is willing to decode for us. SVG is deliberately excluded -
// sharp can rasterize it via librsvg, but vector input has no business being
// an "avatar upload" and librsvg has a CVE history worth just not exposing.
const ACCEPTED_FORMATS = new Set(['jpeg', 'png', 'webp', 'avif']);

// Derived from DATABASE_URL rather than a dedicated env var, so it can never
// drift from wherever `shoka.db` actually lives (same directory db/index.ts
// resolves). AVATAR_DIR is an escape hatch for deployments that want to split
// blob storage from the DB file.
function avatarDir(): string {
	return env.AVATAR_DIR ?? path.join(path.dirname(env.DATABASE_URL), 'avatars');
}

let dirReady: Promise<void> | undefined;

/** Idempotent; memoized so concurrent writers don't race duplicate mkdirs. */
function ensureAvatarDir(): Promise<void> {
	dirReady ??= fs.mkdir(avatarDir(), { recursive: true }).then(() => undefined);
	return dirReady;
}

export class AvatarError extends Error {}

/**
 * Decode, verify, and re-encode an upload into a fixed-size WebP avatar.
 *
 * Re-encoding through sharp - rather than trusting the client's Content-Type
 * or byte-copying the upload - is the actual security control: whatever
 * polyglot or malformed payload comes in, what goes out is a WebP that sharp
 * itself produced. `bytes` is expected to already be a client-side-cropped
 * square (see AvatarCropDialog); `.resize(..., { fit: 'cover' })` is a
 * safety net for non-square input, not the primary cropping step.
 */
export async function processAvatar(bytes: Buffer): Promise<Buffer> {
	const image = sharp(bytes, { limitInputPixels: 64_000_000, failOn: 'warning' });

	// Inferred structurally rather than typed as `sharp.Metadata` - the
	// `export =` module TS uses for its default export doesn't merge that
	// namespace onto a `verbatimModuleSyntax`-style default import.
	let metadata: Awaited<ReturnType<typeof image.metadata>>;
	try {
		metadata = await image.metadata();
	} catch {
		throw new AvatarError('Could not read image.');
	}

	if (!metadata.format || !ACCEPTED_FORMATS.has(metadata.format)) {
		throw new AvatarError('Unsupported image format. Use JPEG, PNG, WebP, or AVIF.');
	}

	try {
		return await image
			.rotate() // bake in EXIF orientation before resizing
			.resize(AVATAR_SIZE, AVATAR_SIZE, { fit: 'cover', position: 'centre' })
			.webp({ quality: 82 })
			.toBuffer();
	} catch {
		throw new AvatarError('Could not process image.');
	}
}

/** Writes `buffer` under a fresh UUID filename and returns it. Atomic rename. */
export async function writeAvatar(buffer: Buffer): Promise<string> {
	await ensureAvatarDir();

	const filename = `${crypto.randomUUID()}.webp`;
	const dest = path.join(avatarDir(), filename);
	const tmp = `${dest}.tmp`;

	await fs.writeFile(tmp, buffer);
	await fs.rename(tmp, dest);

	return filename;
}

/** Absolute path for a filename already validated against AVATAR_FILENAME_RE. */
export function avatarPath(filename: string): string {
	return path.join(avatarDir(), filename);
}

/** Best-effort unlink; a missing file is not an error worth surfacing. */
export async function deleteAvatarFile(filename: string): Promise<void> {
	try {
		await fs.unlink(avatarPath(filename));
	} catch (err) {
		if ((err as NodeJS.ErrnoException).code !== 'ENOENT') throw err;
	}
}

/**
 * True only for a path this module wrote (`/avatars/<uuid>.webp`). `user.image`
 * is a free-form column - an OAuth provider could point it at an external URL -
 * so this guards every unlink against ever touching something we don't own.
 */
export function isManagedAvatar(image: string | null | undefined): image is string {
	if (!image) return false;
	const filename = image.replace(/^\/avatars\//, '');
	return filename !== image && AVATAR_FILENAME_RE.test(filename);
}

/** Inverse of the `/avatars/<file>` URL this module hands to better-auth. */
export function filenameFromImage(image: string): string {
	return image.replace(/^\/avatars\//, '');
}

export { avatarDir };
