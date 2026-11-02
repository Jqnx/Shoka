import { error, fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { errorMessage } from '$lib/server/api';
import { guard } from '$lib/server/guard';
import { can } from '$lib/auth/permissions';
import type { Artist, ArtistDetail, ArchiveListResponse, AudioListResponse } from '$lib/types';

const LIMIT = 40;

export const load: PageServerLoad = async ({ fetch, params, url, locals }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const canMerge =
		can(locals.user, { taxonomy: ['update'] }) && can(locals.user, { taxonomy: ['delete'] });

	// The archives-by-artist endpoint doesn't include the artist's own name,
	// so it's fetched separately.
	const [artistRes, archivesRes, allArtistsRes] = await Promise.all([
		fetch(`/api/artists/${params.id}`),
		fetch(`/api/artists/${params.id}/archives?page=${page}&limit=${LIMIT}`),
		// Only the merge dialog reads this list, and only merge-capable users
		// can open it; skip the payload for everyone else.
		canMerge ? fetch('/api/artists/all') : null
	]);
	if (artistRes.status === 404) error(404, 'Artist not found');
	if (!artistRes.ok) error(500, 'Failed to load artist');
	if (!archivesRes.ok) error(500, 'Failed to load archives');

	const artist: ArtistDetail = await artistRes.json();

	// Audio is a separate section on the page; a failure there must not fail
	// the archive listing.
	const audioRes = await fetch(`/api/audio?artist=${encodeURIComponent(artist.name)}&limit=12`);
	const audioData: AudioListResponse | null = audioRes.ok ? await audioRes.json() : null;

	const data: ArchiveListResponse = await archivesRes.json();
	// The merge picker is optional UI; a missing or failed list must not fail
	// the page.
	const allArtists: Artist[] = allArtistsRes?.ok ? await allArtistsRes.json() : [];

	// Only trust a `from` that points back into the artist listing; anything
	// else would make the back link leave the section. Matching the whole
	// pathname (plus a query) matters: a bare prefix check would accept
	// `/artist/../tags`, which resolves elsewhere.
	const from = url.searchParams.get('from');
	const backHref = from && (from === '/artist' || from.startsWith('/artist?')) ? from : '/artist';

	return {
		artist,
		archives: data.items ?? [],
		total: Number(data.total),
		page,
		limit: LIMIT,
		backHref,
		allArtists,
		audio: audioData?.items ?? [],
		audioTotal: Number(audioData?.total ?? 0)
	};
};

export const actions: Actions = {
	toggleFavorite: async ({ request, fetch, params }) => {
		const form = await request.formData();
		// PUT/DELETE are both idempotent, so this only picks the direction.
		const wasFavorited = form.get('favorited') === 'true';

		const res = await fetch(`/api/artists/${params.id}/favorite`, {
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

		const res = await fetch(`/api/artists/${params.id}/rating`, {
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
		const res = await fetch(`/api/artists/${params.id}/rating`, { method: 'DELETE' });

		if (!res.ok) {
			return fail(res.status, {
				action: 'clearRating' as const,
				error: await errorMessage(res, 'Failed to clear rating.')
			});
		}

		return { action: 'clearRating' as const, success: true };
	},

	update: async ({ request, fetch, params, locals }) => {
		const denied = guard(
			locals.user,
			{ taxonomy: ['update'] },
			'update',
			'You cannot edit artists.'
		);
		if (denied) return denied;

		const form = await request.formData();
		const name = String(form.get('name') ?? '').trim();
		if (!name) {
			return fail(400, { action: 'update' as const, error: 'Name cannot be empty.' });
		}

		// The sheet round-trips the current lists, so sending both always is a
		// full replace on every save and an emptied list genuinely clears.
		const res = await fetch(`/api/artists/${params.id}`, {
			method: 'PATCH',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				name,
				aliases: form.getAll('aliases').map(String),
				urls: form.getAll('urls').map(String)
			})
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'update' as const,
				error: await errorMessage(res, 'Failed to update artist.')
			});
		}

		return { action: 'update' as const, success: true };
	},

	setCover: async ({ request, fetch, params, locals }) => {
		const denied = guard(
			locals.user,
			{ taxonomy: ['update'] },
			'setCover',
			'You cannot edit artists.'
		);
		if (denied) return denied;

		const form = await request.formData();
		// "" restores the default (the artist's oldest archive).
		const coverArchiveID = String(form.get('archive_id') ?? '');

		const res = await fetch(`/api/artists/${params.id}`, {
			method: 'PATCH',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ cover_archive_id: coverArchiveID })
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'setCover' as const,
				error: await errorMessage(res, 'Failed to change the artist image.')
			});
		}

		return { action: 'setCover' as const, success: true };
	},

	deleteArtist: async ({ fetch, params, locals }) => {
		const denied = guard(
			locals.user,
			{ taxonomy: ['delete'] },
			'deleteArtist',
			'You cannot delete artists.'
		);
		if (denied) return denied;

		const res = await fetch(`/api/artists/${params.id}`, { method: 'DELETE' });

		if (!res.ok) {
			return fail(res.status, {
				action: 'deleteArtist' as const,
				error: await errorMessage(res, 'Failed to delete artist.')
			});
		}

		redirect(303, '/artist');
	},

	mergeArtist: async ({ request, fetch, params, locals }) => {
		// Both verbs are checked separately: better-auth's authorize semantics
		// for a multi-action check are not relied on elsewhere.
		const denied =
			guard(locals.user, { taxonomy: ['update'] }, 'mergeArtist', 'You cannot merge artists.') ??
			guard(locals.user, { taxonomy: ['delete'] }, 'mergeArtist', 'You cannot merge artists.');
		if (denied) return denied;

		const form = await request.formData();
		const targetID = Number(form.get('target_id'));
		if (!Number.isInteger(targetID) || targetID < 1 || targetID === Number(params.id)) {
			return fail(400, {
				action: 'mergeArtist' as const,
				error: 'Choose a different artist to merge into.'
			});
		}

		const res = await fetch(`/api/artists/${params.id}/merge`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ target_id: targetID })
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'mergeArtist' as const,
				error: await errorMessage(res, 'Failed to merge artists.')
			});
		}

		redirect(303, '/artist/' + targetID);
	}
};
