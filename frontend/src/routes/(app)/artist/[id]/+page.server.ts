import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { Artist, ArchiveListResponse } from '$lib/types';

const LIMIT = 40;

export const load: PageServerLoad = async ({ fetch, params, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));

	// The archives-by-artist endpoint doesn't include the artist's own name,
	// so it's fetched separately (unlike tag/character/parody, which are
	// looked up by name in the first place and so already have it).
	const [artistRes, archivesRes] = await Promise.all([
		fetch(`/api/artists/${params.id}`),
		fetch(`/api/artists/${params.id}/archives?page=${page}&limit=${LIMIT}`)
	]);
	if (artistRes.status === 404) error(404, 'Artist not found');
	if (!artistRes.ok) error(500, 'Failed to load artist');
	if (!archivesRes.ok) error(500, 'Failed to load archives');

	const artist: Artist = await artistRes.json();
	const data: ArchiveListResponse = await archivesRes.json();

	return {
		id: params.id,
		name: artist.name,
		archives: data.items ?? [],
		total: Number(data.total),
		page,
		limit: LIMIT
	};
};
