import type { PageServerLoad } from './$types';
import type { ArchiveListResponse, Character, Parody, Tag } from '$lib/types';

const DEFAULT_LIMIT = 40;
const MAX_LIMIT = 100;

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const limit = Math.min(
		MAX_LIMIT,
		Math.max(1, parseInt(url.searchParams.get('limit') ?? String(DEFAULT_LIMIT)) || DEFAULT_LIMIT)
	);

	const filters = {
		sort: url.searchParams.get('sort') ?? '',
		category: url.searchParams.get('category') ?? '',
		language: url.searchParams.get('language') ?? '',
		tag: url.searchParams.getAll('tag'),
		character: url.searchParams.getAll('character'),
		parody: url.searchParams.getAll('parody'),
		artist: url.searchParams.get('artist') ?? ''
	};

	const archiveParams = new URLSearchParams({ page: String(page), limit: String(limit) });
	if (filters.sort) archiveParams.set('sort', filters.sort);
	if (filters.category) archiveParams.set('category', filters.category);
	if (filters.language) archiveParams.set('language', filters.language);
	if (filters.artist) archiveParams.set('artist', filters.artist);
	for (const tag of filters.tag) archiveParams.append('tag', tag);
	for (const character of filters.character) archiveParams.append('character', character);
	for (const parody of filters.parody) archiveParams.append('parody', parody);

	const [archivesRes, tagsRes, charactersRes, parodiesRes] = await Promise.all([
		fetch(`/api/archives?${archiveParams}`),
		fetch('/api/tags/all'),
		fetch('/api/characters/all'),
		fetch('/api/parodies/all')
	]);

	const archivesData: ArchiveListResponse = archivesRes.ok
		? await archivesRes.json()
		: { items: [], total: 0, page, limit };
	const tags: Tag[] = tagsRes.ok ? await tagsRes.json() : [];
	const characters: Character[] = charactersRes.ok ? await charactersRes.json() : [];
	const parodies: Parody[] = parodiesRes.ok ? await parodiesRes.json() : [];

	return {
		archives: archivesData.items ?? [],
		total: Number(archivesData.total),
		page,
		limit,
		filters,
		tags,
		characters,
		parodies
	};
};
