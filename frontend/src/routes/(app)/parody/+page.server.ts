import type { PageServerLoad } from './$types';
import type { ParodyListResponse } from '$lib/types';

const LIMIT = 100;

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const res = await fetch(`/api/parodies?page=${page}&limit=${LIMIT}`);
	if (!res.ok) return { parodies: [], total: 0, page, limit: LIMIT };
	const data: ParodyListResponse = await res.json();
	return { parodies: data.items ?? [], total: Number(data.total), page, limit: LIMIT };
};
