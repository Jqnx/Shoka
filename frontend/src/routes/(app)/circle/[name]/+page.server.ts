import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { loadAudioSections } from '$lib/server/audio-sections';
import type { ArchiveListResponse, CircleArtist } from '$lib/types';

const LIMIT = 40;

export const load: PageServerLoad = async ({ fetch, params, url, parent }) => {
	const { libraries } = await parent();
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const name = encodeURIComponent(params.name);

	const [res, artistsRes, audioSections] = await Promise.all([
		fetch(`/api/circles/${name}?page=${page}&limit=${LIMIT}`),
		fetch(`/api/circles/${name}/artists`),
		loadAudioSections(fetch, libraries, { name: 'circle', value: params.name })
	]);
	if (!res.ok) error(res.status === 404 ? 404 : 500, 'Failed to load archives');
	const data: ArchiveListResponse = await res.json();
	const artists: CircleArtist[] = artistsRes.ok ? await artistsRes.json() : [];
	return {
		name: params.name,
		archives: data.items ?? [],
		total: Number(data.total),
		page,
		limit: LIMIT,
		artists,
		audioSections
	};
};
