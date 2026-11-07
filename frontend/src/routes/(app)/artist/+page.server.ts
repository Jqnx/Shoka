import { fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { errorMessage } from '$lib/server/api';
import type { ArtistListResponse } from '$lib/types';

const LIMIT = 40;

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const favorites = url.searchParams.get('favorites') === 'true';

	// Only a single digit 1-5 filters; anything else (3.5, 4abc, 99) filters
	// on nothing, so the highlighted stars stay honest about what the
	// backend actually applied.
	const rawRating = url.searchParams.get('rating') ?? '';
	const rating = /^[1-5]$/.test(rawRating) ? Number(rawRating) : 0;

	const params = new URLSearchParams({ page: String(page), limit: String(LIMIT) });
	if (favorites) params.set('favorites', 'true');
	if (rating) params.set('rating', String(rating));

	const res = await fetch(`/api/artists?${params}`);
	if (!res.ok) return { artists: [], total: 0, page, limit: LIMIT, favorites, rating };
	const data: ArtistListResponse = await res.json();
	return {
		artists: data.items ?? [],
		total: Number(data.total),
		page,
		limit: LIMIT,
		favorites,
		rating
	};
};

export const actions: Actions = {
	toggleFavorite: async ({ request, fetch }) => {
		const form = await request.formData();
		const id = Number(form.get('id'));
		if (!Number.isInteger(id) || id < 1) return fail(400, { error: 'Invalid artist.' });
		const wasFavorited = form.get('favorited') === 'true';
		const res = await fetch(`/api/artists/${id}/favorite`, {
			method: wasFavorited ? 'DELETE' : 'PUT'
		});
		if (!res.ok)
			return fail(res.status, { error: await errorMessage(res, 'Failed to update favorite.') });
		return { success: true };
	}
};
