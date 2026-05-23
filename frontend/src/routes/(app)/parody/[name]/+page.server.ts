import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { ArchiveListResponse } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, params }) => {
	const res = await fetch(`/api/parodies/${encodeURIComponent(params.name)}`);
	if (!res.ok) error(res.status === 404 ? 404 : 500, 'Failed to load archives');
	const data: ArchiveListResponse = await res.json();
	return { name: params.name, archives: data.items ?? [], total: data.total };
};
