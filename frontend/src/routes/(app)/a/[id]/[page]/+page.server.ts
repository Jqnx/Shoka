import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { Archive, ReaderSettings } from '$lib/types';

// This route resets past the (app) layout (see +page@.svelte) for a
// chrome-free, full-bleed reader - which also means it skips that layout's
// own auth guard, so it needs its own.
const DEFAULT_READER_SETTINGS: ReaderSettings = {
	view_mode: 'paged',
	reading_direction: 'rtl',
	page_layout: 'single',
	fit_mode: 'width',
	background: 'black'
};

export const load: PageServerLoad = async ({ locals, fetch, params }) => {
	if (!locals.user) redirect(302, '/login');

	const [archiveRes, settingsRes] = await Promise.all([
		fetch(`/api/archives/${params.id}`),
		fetch('/api/reader-settings')
	]);

	if (archiveRes.status === 404) error(404, 'Archive not found');
	if (!archiveRes.ok) error(500, 'Failed to load archive');
	const archive: Archive = await archiveRes.json();

	// An archive with no pages (still scanning, extraction failed, etc.) has
	// no valid page to redirect to below - bail out before that logic can
	// loop between "too low" and "too high".
	if (archive.page_count < 1) error(404, 'Archive has no pages');

	// The page number is 1-based and part of the path rather than a query
	// param - out-of-range or non-numeric values redirect to the nearest
	// valid page instead of 404ing, since they're easy to end up with (typos,
	// a stale link after re-scanning shrinks the archive, etc.) and there's
	// an obvious sane page to fall back to.
	const rawPage = Number(params.page);
	if (!Number.isInteger(rawPage) || rawPage < 1) {
		redirect(302, `/a/${params.id}/1`);
	}
	if (rawPage > archive.page_count) {
		redirect(302, `/a/${params.id}/${archive.page_count}`);
	}

	// Best-effort - a user's reader settings not loading shouldn't block them
	// from reading, just fall back to the same defaults the backend itself
	// would use for a never-customized user.
	const settings: ReaderSettings = settingsRes.ok
		? await settingsRes.json()
		: DEFAULT_READER_SETTINGS;

	return { archive, settings, initialPage: rawPage - 1 };
};
