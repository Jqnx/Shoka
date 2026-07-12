import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { ArchiveListResponse } from '$lib/types';

const LIMIT = 40;

export const load: PageServerLoad = async ({ fetch, params, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const res = await fetch(
		`/api/tags/${encodeURIComponent(params.name)}?page=${page}&limit=${LIMIT}`
	);
	if (!res.ok) error(res.status === 404 ? 404 : 500, 'Failed to load archives');
	const data: ArchiveListResponse = await res.json();
	return {
		name: params.name,
		archives: data.items ?? [],
		total: Number(data.total),
		page,
		limit: LIMIT
	};
};
