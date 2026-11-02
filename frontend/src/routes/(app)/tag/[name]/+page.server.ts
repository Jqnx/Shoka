import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { loadAudioSections } from '$lib/server/audio-sections';
import type { ArchiveListResponse } from '$lib/types';

const LIMIT = 40;

export const load: PageServerLoad = async ({ fetch, params, url, parent }) => {
	const { libraries } = await parent();
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const name = encodeURIComponent(params.name);

	const [res, audioSections] = await Promise.all([
		fetch(`/api/tags/${name}?page=${page}&limit=${LIMIT}`),
		loadAudioSections(fetch, libraries, { name: 'tag', value: params.name })
	]);
	if (!res.ok) error(res.status === 404 ? 404 : 500, 'Failed to load archives');
	const data: ArchiveListResponse = await res.json();
	return {
		name: params.name,
		archives: data.items ?? [],
		total: Number(data.total),
		page,
		limit: LIMIT,
		audioSections
	};
};
