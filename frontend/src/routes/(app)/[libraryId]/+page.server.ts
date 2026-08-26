import { error, fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { errorMessage } from '$lib/server/api';
import type {
	Artist,
	ArchiveLanguage,
	ArchiveListResponse,
	ArchiveSortOption,
	Character,
	MetadataSourceInfo,
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

	// Only 'true'/'false' are meaningful here - the backend ignores anything
	// else and filters on neither, so normalising junk to '' keeps the UI's
	// displayed selection honest about what was actually applied.
	const rawHasPHash = url.searchParams.get('has_phash');
	const hasPHash = rawHasPHash === 'true' || rawHasPHash === 'false' ? rawHasPHash : '';

	const filters = {
		q: url.searchParams.get('q') ?? '',
		has_phash: hasPHash,
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
	if (filters.has_phash) archiveParams.set('has_phash', filters.has_phash);
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
		languagesRes,
		sourcesRes
	] = await Promise.all([
		fetch(`/api/archives?${archiveParams}`),
		fetch('/api/artists/all'),
		fetch('/api/tags/all'),
		fetch('/api/characters/all'),
		fetch('/api/parodies/all'),
		fetch('/api/archives/sort-options'),
		fetch('/api/archives/categories'),
		fetch('/api/archives/languages'),
		// Which sources the bulk-identify picker may offer. Enablement is
		// per-library, and the backend silently skips a source that's disabled
		// for the archive's library, so the picker needs this to avoid
		// offering a no-op.
		fetch(`/api/metadata/sources?library_id=${library.id}`)
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
	const metadataSources: MetadataSourceInfo[] = sourcesRes.ok ? await sourcesRes.json() : [];

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
		languages,
		metadataSources
	};
};

// Every bulk endpoint answers with this shape and is deliberately
// partial-success: a single bad archive lands in `failed` instead of
// rolling back the ones that worked.
type BulkResult = {
	requested: number;
	succeeded: number;
	failed: { id: string; error: string }[];
};

function summarize(action: string, result: BulkResult) {
	return {
		action,
		success: true as const,
		requested: result.requested,
		succeeded: result.succeeded,
		failed: result.failed?.length ?? 0
	};
}

export const actions: Actions = {
	bulkProgress: async ({ request, fetch }) => {
		const form = await request.formData();
		const ids = form.getAll('ids').map(String);
		const read = form.get('read') === 'true';
		if (ids.length === 0) {
			return fail(400, { action: 'bulkProgress' as const, error: 'No archives selected.' });
		}

		const res = await fetch('/api/archives/bulk/progress', {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ ids, read })
		});
		if (!res.ok) {
			return fail(res.status, {
				action: 'bulkProgress' as const,
				error: await errorMessage(res, 'Failed to update archives.')
			});
		}
		return summarize(read ? 'markRead' : 'markUnread', await res.json());
	},

	bulkIdentify: async ({ request, fetch }) => {
		const form = await request.formData();
		const ids = form.getAll('ids').map(String);
		if (ids.length === 0) {
			return fail(400, { action: 'bulkIdentify' as const, error: 'No archives selected.' });
		}

		// Omitted means "run the normal priority-ordered pipeline"; the backend
		// rejects an unknown name outright, so only send a non-empty one.
		const source = String(form.get('source') ?? '').trim();

		const res = await fetch('/api/archives/bulk/metadata', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(source ? { ids, source } : { ids })
		});
		if (!res.ok) {
			return fail(res.status, {
				action: 'bulkIdentify' as const,
				error: await errorMessage(res, 'Failed to queue identification.')
			});
		}
		return summarize('identify', await res.json());
	},

	bulkEdit: async ({ request, fetch }) => {
		const form = await request.formData();
		const ids = form.getAll('ids').map(String);
		if (ids.length === 0) {
			return fail(400, { action: 'bulkEdit' as const, error: 'No archives selected.' });
		}

		// Relations are additive on the backend (existing values are kept and
		// unioned with these), and an omitted key means "leave alone" - so
		// empty lists are dropped rather than sent, which would be a no-op
		// anyway. Scalars overwrite, and '' is how you deliberately blank one,
		// hence the explicit "was it submitted at all" check.
		const body: Record<string, unknown> = { ids };
		for (const field of ['artists', 'tags', 'parodies', 'circles', 'characters']) {
			const values = form.getAll(field).map(String).filter(Boolean);
			if (values.length > 0) body[field] = values;
		}
		for (const field of ['language', 'category']) {
			if (form.has(field)) body[field] = String(form.get(field) ?? '');
		}

		const res = await fetch('/api/archives/bulk', {
			method: 'PATCH',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(body)
		});
		if (!res.ok) {
			return fail(res.status, {
				action: 'bulkEdit' as const,
				error: await errorMessage(res, 'Failed to update archives.')
			});
		}
		return summarize('edit', await res.json());
	}
};
