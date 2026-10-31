import { fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { errorMessage } from '$lib/server/api';
import type { ArtistListResponse } from '$lib/types';

const LIMIT = 40;

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const favorites = url.searchParams.get('favorites') === 'true';
	const params = new URLSearchParams({ page: String(page), limit: String(LIMIT) });
	if (favorites) params.set('favorites', 'true');
	const res = await fetch(`/api/artists?${params}`);
	if (!res.ok) return { artists: [], total: 0, page, limit: LIMIT, favorites };
	const data: ArtistListResponse = await res.json();
	return { artists: data.items ?? [], total: Number(data.total), page, limit: LIMIT, favorites };
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
