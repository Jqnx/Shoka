import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { FavoritePageItem, FavoritePageListResponse, ReaderSettings } from '$lib/types';

// This route resets past the (app) layout (see +page@.svelte) for a
// chrome-free, full-bleed viewer - which also means it skips that layout's
// own auth guard, so it needs its own.
const DEFAULT_READER_SETTINGS: ReaderSettings = {
	view_mode: 'paged',
	reading_direction: 'rtl',
	page_layout: 'single',
	fit_mode: 'height',
	background: 'black',
	slideshow_interval_seconds: 5,
	slideshow_loop: true
};

const PAGE_SIZE = 100;

export const load: PageServerLoad = async ({ locals, fetch, url }) => {
	if (!locals.user) redirect(302, '/login');

	// Best-effort, same as the reader: fall back to the backend's own defaults.
	const settingsRes = await fetch('/api/reader-settings');
	const settings: ReaderSettings = settingsRes.ok
		? await settingsRes.json()
		: DEFAULT_READER_SETTINGS;

	// The whole list is loaded because the slideshow walks it in order; one
	// page of metadata is tiny and the images lazy-load.
	const items: FavoritePageItem[] = [];
	for (let p = 1; ; p++) {
		const res = await fetch(`/api/archives/favorites/pages?page=${p}&limit=${PAGE_SIZE}`);
		if (!res.ok) break;

		const data: FavoritePageListResponse = await res.json();
		const pageItems = data.items ?? [];
		items.push(...pageItems);

		if (pageItems.length === 0 || items.length >= Number(data.total)) break;
	}

	if (items.length === 0) redirect(302, '/favorites/pages');

	const rawAt = Number(url.searchParams.get('at') ?? '0');
	const at = Number.isInteger(rawAt) ? Math.min(Math.max(rawAt, 0), items.length - 1) : 0;

	return { items, at, settings };
};
