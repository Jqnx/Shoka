// Enforced server-side on the *uploaded* (already client-cropped 512x512)
// blob - a generous ceiling, since a real one lands around 40-80KB.
export const MAX_UPLOAD_BYTES = 2 * 1024 * 1024;

// Client-only: guards against feeding an absurdly large source photo into an
// <img> for cropping before any resizing has happened. Not enforced server
// side - by the time a file reaches the action it has already been through
// the crop dialog's canvas resize, which makes this limit moot.
export const MAX_SOURCE_BYTES = 20 * 1024 * 1024;

// Sniffed via sharp's own format detection server-side, never trusted from a
// client-supplied Content-Type. This list is only used for the client-side
// `<input accept>` hint and the fast pre-check.
export const ACCEPTED_MIME_TYPES = ['image/jpeg', 'image/png', 'image/webp', 'image/avif'];

export const AVATAR_FILENAME_RE =
	/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.webp$/;
