import type { PageServerLoad } from './$types';
import type { ArchiveListResponse } from '$lib/types';

const CAROUSEL_LIMIT = 10;

export const load: PageServerLoad = async ({ fetch, parent }) => {
	const { libraries } = await parent();
	// GetArchives is scoped to a single library; default to the first one
	// until there's a "current library" concept for the home page to draw from.
	// Recently-read is its own endpoint and deliberately spans every library,
	// so it doesn't need libraryId at all.
	const libraryId = libraries[0]?.id;

	const [addedRes, releasedRes, readRes] = await Promise.all([
		libraryId
			? fetch(`/api/archives?library_id=${libraryId}&sort=created_at_desc&limit=${CAROUSEL_LIMIT}`)
			: null,
		libraryId
			? fetch(`/api/archives?library_id=${libraryId}&sort=release_date_desc&limit=${CAROUSEL_LIMIT}`)
			: null,
		fetch(`/api/archives/recently-read?limit=${CAROUSEL_LIMIT}`)
	]);

	const addedData: ArchiveListResponse =
		addedRes?.ok ? await addedRes.json() : { items: [], total: 0, page: 1, limit: CAROUSEL_LIMIT };
	const releasedData: ArchiveListResponse =
		releasedRes?.ok
			? await releasedRes.json()
			: { items: [], total: 0, page: 1, limit: CAROUSEL_LIMIT };
	const readData: ArchiveListResponse = readRes.ok
		? await readRes.json()
		: { items: [], total: 0, page: 1, limit: CAROUSEL_LIMIT };

	return {
		recentlyAdded: addedData.items ?? [],
		recentlyReleased: (releasedData.items ?? []).filter((a) => a.release_date !== null),
		recentlyRead: readData.items ?? []
	};
};
