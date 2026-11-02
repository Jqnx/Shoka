import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { AudioListResponse } from '$lib/types';

const LIMIT = 40;

export const load: PageServerLoad = async ({ fetch, url, parent }) => {
	const { libraries } = await parent();
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const libraryId = url.searchParams.get('library_id') ?? '';

	let library = null;

	if (libraryId) {
		library = libraries.find((l) => l.id === libraryId) ?? null;
		if (!library || library.type !== 'audio') error(404, 'Library not found');
	}

	// Only 'true' and a single digit 1-5 are meaningful; the backend ignores
	// anything else, so normalising junk to unset keeps the highlighted
	// controls honest about what was applied.
	const rawRating = url.searchParams.get('rating') ?? '';
	const rating = /^[1-5]$/.test(rawRating) ? rawRating : '';
	const favorites = url.searchParams.get('favorites') === 'true' ? 'true' : '';

	const filters = {
		q: url.searchParams.get('q') ?? '',
		favorites,
		rating,
		tag: url.searchParams.getAll('tag'),
		artist: url.searchParams.getAll('artist'),
		circle: url.searchParams.getAll('circle'),
		voice_actor: url.searchParams.getAll('voice_actor')
	};

	const params = new URLSearchParams({ page: String(page), limit: String(LIMIT) });
	if (library) params.set('library_id', library.id);
	if (filters.q) params.set('q', filters.q);
	if (filters.favorites) params.set('favorites', filters.favorites);
	if (filters.rating) params.set('rating', filters.rating);
	for (const tag of filters.tag) params.append('tag', tag);
	for (const artist of filters.artist) params.append('artist', artist);
	for (const circle of filters.circle) params.append('circle', circle);
	for (const voiceActor of filters.voice_actor) params.append('voice_actor', voiceActor);

	const res = await fetch(`/api/audio?${params}`);
	const audio: AudioListResponse = res.ok
		? await res.json()
		: { items: [], total: 0, page, limit: LIMIT };

	return {
		library,
		audio: audio.items ?? [],
		total: Number(audio.total),
		page,
		limit: LIMIT,
		filters
	};
};
