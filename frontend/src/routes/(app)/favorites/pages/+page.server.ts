import { fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import type { FavoritePageListResponse } from '$lib/types';
import { errorMessage } from '$lib/server/api';

const DEFAULT_LIMIT = 40;
const MAX_LIMIT = 100;

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const limit = Math.min(
		MAX_LIMIT,
		Math.max(1, parseInt(url.searchParams.get('limit') ?? String(DEFAULT_LIMIT)) || DEFAULT_LIMIT)
	);

	const pagesRes = await fetch(`/api/archives/favorites/pages?page=${page}&limit=${limit}`);
	const pagesData: FavoritePageListResponse = pagesRes.ok
		? await pagesRes.json()
		: { items: [], total: 0, page, limit };

	return {
		items: pagesData.items ?? [],
		total: Number(pagesData.total),
		page,
		limit
	};
};

export const actions: Actions = {
	removeFavorite: async ({ request, fetch }) => {
		const form = await request.formData();
		const archiveID = String(form.get('archive_id') ?? '');
		const pageIndex = Number(form.get('page_index'));

		if (!archiveID || !Number.isInteger(pageIndex) || pageIndex < 0) {
			return fail(400, { action: 'removeFavorite' as const, error: 'Invalid favorite.' });
		}

		const res = await fetch(`/api/archives/${archiveID}/pages/${pageIndex}/favorite`, {
			method: 'DELETE'
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'removeFavorite' as const,
				error: await errorMessage(res, 'Failed to remove favorite')
			});
		}

		return { action: 'removeFavorite' as const, success: true };
	}
};
