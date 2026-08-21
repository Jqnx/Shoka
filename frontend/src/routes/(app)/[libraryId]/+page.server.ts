import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type {
	Artist,
	ArchiveLanguage,
	ArchiveListResponse,
	ArchiveSortOption,
	Character,
	Parody,
	Tag
} from '$lib/types';

const DEFAULT_LIMIT = 40;
const MAX_LIMIT = 100;

export const load: PageServerLoad = async ({ fetch, url, params, parent }) => {
	const { libraries } = await parent();
	const library = libraries.find((l) => l.id === params.libraryId);
	if (!library) error(404, 'Library not found');

	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const limit = Math.min(
		MAX_LIMIT,
		Math.max(1, parseInt(url.searchParams.get('limit') ?? String(DEFAULT_LIMIT)) || DEFAULT_LIMIT)
	);

	const filters = {
		q: url.searchParams.get('q') ?? '',
		sort: url.searchParams.get('sort') ?? '',
		category: url.searchParams.get('category') ?? '',
		language: url.searchParams.get('language') ?? '',
		tag: url.searchParams.getAll('tag'),
		character: url.searchParams.getAll('character'),
		parody: url.searchParams.getAll('parody'),
		artist: url.searchParams.getAll('artist')
	};

	const archiveParams = new URLSearchParams({
		library_id: library.id,
		page: String(page),
		limit: String(limit)
	});
	if (filters.q) archiveParams.set('q', filters.q);
	if (filters.sort) archiveParams.set('sort', filters.sort);
	if (filters.category) archiveParams.set('category', filters.category);
	if (filters.language) archiveParams.set('language', filters.language);
	for (const tag of filters.tag) archiveParams.append('tag', tag);
	for (const character of filters.character) archiveParams.append('character', character);
	for (const parody of filters.parody) archiveParams.append('parody', parody);
	for (const artist of filters.artist) archiveParams.append('artist', artist);

	const [
		archivesRes,
		artistsRes,
		tagsRes,
		charactersRes,
		parodiesRes,
		sortOptionsRes,
		categoriesRes,
		languagesRes
	] = await Promise.all([
		fetch(`/api/archives?${archiveParams}`),
		fetch('/api/artists/all'),
		fetch('/api/tags/all'),
		fetch('/api/characters/all'),
		fetch('/api/parodies/all'),
		fetch('/api/archives/sort-options'),
		fetch('/api/archives/categories'),
		fetch('/api/archives/languages')
	]);

	const archivesData: ArchiveListResponse = archivesRes.ok
		? await archivesRes.json()
		: { items: [], total: 0, page, limit };
	const artists: Artist[] = artistsRes.ok ? await artistsRes.json() : [];
	const tags: Tag[] = tagsRes.ok ? await tagsRes.json() : [];
	const characters: Character[] = charactersRes.ok ? await charactersRes.json() : [];
	const parodies: Parody[] = parodiesRes.ok ? await parodiesRes.json() : [];
	const sortOptions: ArchiveSortOption[] = sortOptionsRes.ok ? await sortOptionsRes.json() : [];
	const categories: string[] = categoriesRes.ok ? await categoriesRes.json() : [];
	const languages: ArchiveLanguage[] = languagesRes.ok ? await languagesRes.json() : [];

	return {
		library,
		archives: archivesData.items ?? [],
		total: Number(archivesData.total),
		page,
		limit,
		filters,
		artists,
		tags,
		characters,
		parodies,
		sortOptions,
		categories,
		languages
	};
};
