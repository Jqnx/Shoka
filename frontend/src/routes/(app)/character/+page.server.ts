import type { PageServerLoad } from './$types';
import type { CharacterListResponse } from '$lib/types';

const LIMIT = 100;

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const res = await fetch(`/api/characters?page=${page}&limit=${LIMIT}`);
	if (!res.ok) return { characters: [], total: 0, page, limit: LIMIT };
	const data: CharacterListResponse = await res.json();
	return { characters: data.items ?? [], total: Number(data.total), page, limit: LIMIT };
};
