import type { PageServerLoad } from './$types';
import type { ArtistListResponse } from '$lib/types';

const LIMIT = 100;

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const res = await fetch(`/api/artists?page=${page}&limit=${LIMIT}`);
	if (!res.ok) return { artists: [], total: 0, page, limit: LIMIT };
	const data: ArtistListResponse = await res.json();
	return { artists: data.items ?? [], total: Number(data.total), page, limit: LIMIT };
};
