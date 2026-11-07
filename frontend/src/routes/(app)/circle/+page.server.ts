import type { PageServerLoad } from './$types';
import type { CircleListResponse } from '$lib/types';

const LIMIT = 100;

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const res = await fetch(`/api/circles?page=${page}&limit=${LIMIT}`);
	if (!res.ok) return { circles: [], total: 0, page, limit: LIMIT };
	const data: CircleListResponse = await res.json();
	return { circles: data.items ?? [], total: Number(data.total), page, limit: LIMIT };
};
