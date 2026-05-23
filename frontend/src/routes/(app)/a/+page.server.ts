import type { PageServerLoad } from './$types';
import type { ArchiveListResponse } from '$lib/types';

const LIMIT = 40;

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const res = await fetch(`/api/archives?page=${page}&limit=${LIMIT}`);
	if (!res.ok) return { archives: [], total: 0, page, limit: LIMIT };
	const data: ArchiveListResponse = await res.json();
	return { archives: data.items ?? [], total: Number(data.total), page, limit: LIMIT };
};
