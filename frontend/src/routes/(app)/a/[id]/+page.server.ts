import { error, fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import type { Archive, ArchiveLanguage, Artist, Character, Parody, Tag } from '$lib/types';
import { errorMessage } from '$lib/server/api';

export const load: PageServerLoad = async ({ fetch, params }) => {
	const [archiveRes, categoriesRes, languagesRes, artistsRes, tagsRes, charactersRes, parodiesRes] =
		await Promise.all([
			fetch(`/api/archives/${params.id}`),
			fetch('/api/archives/categories'),
			fetch('/api/archives/languages'),
			fetch('/api/artists/all'),
			fetch('/api/tags/all'),
			fetch('/api/characters/all'),
			fetch('/api/parodies/all')
		]);
	if (archiveRes.status === 404) error(404, 'Archive not found');
	if (!archiveRes.ok) error(500, 'Failed to load archive');
	const archive: Archive = await archiveRes.json();
	const categories: string[] = categoriesRes.ok ? await categoriesRes.json() : [];
	const languages: ArchiveLanguage[] = languagesRes.ok ? await languagesRes.json() : [];
	// Circles has no equivalent "all" endpoint on the backend, so it stays
	// free-text-only in EditArchiveSheet.
	const allArtists: Artist[] = artistsRes.ok ? await artistsRes.json() : [];
	const allTags: Tag[] = tagsRes.ok ? await tagsRes.json() : [];
	const allCharacters: Character[] = charactersRes.ok ? await charactersRes.json() : [];
	const allParodies: Parody[] = parodiesRes.ok ? await parodiesRes.json() : [];
	return { archive, categories, languages, allArtists, allTags, allCharacters, allParodies };
};

export const actions: Actions = {
	update: async ({ request, fetch, params }) => {
		const form = await request.formData();

		const title = String(form.get('title') ?? '').trim();
		if (!title) {
			return fail(400, { error: 'Title cannot be empty.' });
		}

		const body: Record<string, unknown> = {
			title,
			summary: String(form.get('summary') ?? ''),
			category: String(form.get('category') ?? ''),
			// The archive GET response only ever exposes language as a converted
			// display name ("English"), while the backend stores (and this PATCH
			// endpoint writes back) a raw ISO 639-1 code with no name->code
			// conversion of its own. EditArchiveSheet resolves the archive's
			// display name to its code via /api/archives/languages before
			// prefilling the field, so what's submitted here is already the
			// correct code (or '' to clear it) - safe to always send.
			language: String(form.get('language') ?? ''),
			artists: form.getAll('artists').map(String),
			tags: form.getAll('tags').map(String),
			parodies: form.getAll('parodies').map(String),
			circles: form.getAll('circles').map(String),
			characters: form.getAll('characters').map(String),
			// The edit sheet always round-trips the current list, so an
			// untouched save is a no-op and an emptied list genuinely clears.
			urls: form.getAll('urls').map(String)
		};

		// release_date can't be explicitly cleared via this endpoint (a nil
		// value there means "leave unchanged", same as an omitted key), so
		// only send it when the user actually set a date.
		const releaseDate = String(form.get('release_date') ?? '').trim();
		if (releaseDate) {
			body.release_date = `${releaseDate}T00:00:00Z`;
		}

		const res = await fetch(`/api/archives/${params.id}`, {
			method: 'PATCH',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(body)
		});

		if (!res.ok) {
			return fail(res.status, { error: await errorMessage(res, 'Failed to update archive.') });
		}

		return { success: true };
	},

	markRead: async ({ request, fetch, params }) => {
		const form = await request.formData();
		const pageCount = Number(form.get('page_count'));
		if (!Number.isInteger(pageCount) || pageCount < 1) {
			return fail(400, { action: 'markRead' as const, error: 'Invalid page count.' });
		}

		const res = await fetch(`/api/archives/${params.id}/progress`, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ page: pageCount - 1, completed: true })
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'markRead' as const,
				error: await errorMessage(res, 'Failed to mark as read.')
			});
		}

		return { action: 'markRead' as const, success: true };
	},

	markUnread: async ({ fetch, params }) => {
		const res = await fetch(`/api/archives/${params.id}/progress`, { method: 'DELETE' });

		if (!res.ok) {
			return fail(res.status, {
				action: 'markUnread' as const,
				error: await errorMessage(res, 'Failed to mark as unread.')
			});
		}

		return { action: 'markUnread' as const, success: true };
	},

	toggleFavorite: async ({ request, fetch, params }) => {
		const form = await request.formData();
		// Whether it *was* favorited before this submit - PUT/favorite is
		// idempotent and so is DELETE/unfavorite, so this only decides which
		// direction to flip, not whether the call is safe to make.
		const wasFavorited = form.get('favorited') === 'true';

		const res = await fetch(`/api/archives/${params.id}/favorite`, {
			method: wasFavorited ? 'DELETE' : 'PUT'
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'toggleFavorite' as const,
				error: await errorMessage(res, 'Failed to update favorite.')
			});
		}

		return { action: 'toggleFavorite' as const, success: true };
	},

	setRating: async ({ request, fetch, params }) => {
		const form = await request.formData();
		const rating = Number(form.get('rating'));
		if (!Number.isInteger(rating) || rating < 1 || rating > 5) {
			return fail(400, { action: 'setRating' as const, error: 'Rating must be between 1 and 5.' });
		}

		const res = await fetch(`/api/archives/${params.id}/rating`, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ rating })
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'setRating' as const,
				error: await errorMessage(res, 'Failed to set rating.')
			});
		}

		return { action: 'setRating' as const, success: true };
	},

	clearRating: async ({ fetch, params }) => {
		const res = await fetch(`/api/archives/${params.id}/rating`, { method: 'DELETE' });

		if (!res.ok) {
			return fail(res.status, {
				action: 'clearRating' as const,
				error: await errorMessage(res, 'Failed to clear rating.')
			});
		}

		return { action: 'clearRating' as const, success: true };
	},

	saveMetadata: async ({ fetch, params }) => {
		const res = await fetch(`/api/archives/${params.id}/metadata/save`, { method: 'POST' });

		if (!res.ok) {
			return fail(res.status, {
				action: 'saveMetadata' as const,
				error: await errorMessage(res, 'Failed to save metadata to file.')
			});
		}

		return { action: 'saveMetadata' as const, success: true };
	},

	deleteArchive: async ({ request, fetch, params, url }) => {
		const form = await request.formData();
		const deleteFile = form.get('delete_file') === 'true';

		const res = await fetch(`/api/archives/${params.id}?delete_file=${deleteFile}`, {
			method: 'DELETE'
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'deleteArchive' as const,
				error: await errorMessage(res, 'Failed to delete archive.')
			});
		}

		// The archive is gone, so there's nothing left at /a/{id} to redisplay -
		// send the user back to wherever they came from (same `from` param the
		// back link and EditArchiveSheet already use), or home if there isn't one.
		redirect(303, url.searchParams.get('from') ?? '/');
	}
};
