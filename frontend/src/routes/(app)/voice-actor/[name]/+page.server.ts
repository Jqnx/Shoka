import type { PageServerLoad } from './$types';
import type { AudioListResponse } from '$lib/types';

const LIMIT = 40;

export const load: PageServerLoad = async ({ fetch, params, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const res = await fetch(
		`/api/audio?voice_actor=${encodeURIComponent(params.name)}&page=${page}&limit=${LIMIT}`
	);
	if (!res.ok) return { name: params.name, audio: [], total: 0, page, limit: LIMIT };
	const data: AudioListResponse = await res.json();
	return {
		name: params.name,
		audio: data.items ?? [],
		total: Number(data.total),
		page,
		limit: LIMIT
	};
};
