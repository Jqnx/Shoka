import type { PageServerLoad } from './$types';
import type { TagListResponse } from '$lib/types';

const LIMIT = 100;

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const res = await fetch(`/api/tags?page=${page}&limit=${LIMIT}`);
	if (!res.ok) return { tags: [], total: 0, page, limit: LIMIT };
	const data: TagListResponse = await res.json();
	return { tags: data.items ?? [], total: Number(data.total), page, limit: LIMIT };
};
