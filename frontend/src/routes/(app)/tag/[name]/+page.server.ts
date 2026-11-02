import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { ArchiveListResponse, AudioListResponse } from '$lib/types';

const LIMIT = 40;

export const load: PageServerLoad = async ({ fetch, params, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const name = encodeURIComponent(params.name);

	const [res, audioRes] = await Promise.all([
		fetch(`/api/tags/${name}?page=${page}&limit=${LIMIT}`),
		fetch(`/api/audio?tag=${name}&limit=12`)
	]);
	if (!res.ok) error(res.status === 404 ? 404 : 500, 'Failed to load archives');
	const data: ArchiveListResponse = await res.json();
	const audioData: AudioListResponse | null = audioRes.ok ? await audioRes.json() : null;
	return {
		name: params.name,
		archives: data.items ?? [],
		total: Number(data.total),
		page,
		limit: LIMIT,
		audio: audioData?.items ?? [],
		audioTotal: Number(audioData?.total ?? 0)
	};
};
