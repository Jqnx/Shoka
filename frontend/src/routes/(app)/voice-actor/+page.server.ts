import type { PageServerLoad } from './$types';
import type { VoiceActorListResponse } from '$lib/types';

const LIMIT = 40;

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const res = await fetch(`/api/voice-actors?page=${page}&limit=${LIMIT}`);
	if (!res.ok) return { voiceActors: [], total: 0, page, limit: LIMIT };
	const data: VoiceActorListResponse = await res.json();
	return {
		voiceActors: data.items ?? [],
		total: Number(data.total),
		page,
		limit: LIMIT
	};
};
