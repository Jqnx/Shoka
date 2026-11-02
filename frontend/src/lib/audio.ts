// Helpers shared by the audio components. Kept out of types.ts because they
// are presentation, not wire shapes.

// formatDuration renders a media length as m:ss, or h:mm:ss from an hour up.
// A zero/negative/NaN length (ffprobe unavailable, or a file it could not
// read) reads as 0:00 rather than a broken label.
export function formatDuration(seconds: number): string {
	const total = Number.isFinite(seconds) && seconds > 0 ? Math.floor(seconds) : 0;
	const hours = Math.floor(total / 3600);
	const minutes = Math.floor((total % 3600) / 60);
	const secs = total % 60;

	if (hours > 0) {
		return `${hours}:${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
	}

	return `${minutes}:${String(secs).padStart(2, '0')}`;
}

// Track and image URLs carry updated_at because a rescan can reuse index 0
// for a different file; without a version the browser would keep the old
// bytes under a URL it has already cached.
export function audioTrackSrc(id: string, index: number, updatedAt: string): string {
	return `/api/audio/${id}/tracks/${index}?v=${encodeURIComponent(updatedAt)}`;
}

export function audioImageSrc(id: string, index: number, updatedAt: string): string {
	return `/api/audio/${id}/images/${index}?v=${encodeURIComponent(updatedAt)}`;
}
