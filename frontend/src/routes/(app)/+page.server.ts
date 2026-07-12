import type { PageServerLoad } from './$types';
import type { ArchiveListResponse } from '$lib/types';

const CAROUSEL_LIMIT = 10;

export const load: PageServerLoad = async ({ fetch }) => {
	const [addedRes, releasedRes, readRes] = await Promise.all([
		fetch(`/api/archives?sort=created_at_desc&limit=${CAROUSEL_LIMIT}`),
		fetch(`/api/archives?sort=release_date_desc&limit=${CAROUSEL_LIMIT}`),
		// No backend endpoint/sort exists yet for "has progress, ordered by last read" -
		// fetch a broader page and filter/sort for reading progress client-side.
		fetch('/api/archives?limit=100')
	]);

	const addedData: ArchiveListResponse = addedRes.ok
		? await addedRes.json()
		: { items: [], total: 0, page: 1, limit: CAROUSEL_LIMIT };
	const releasedData: ArchiveListResponse = releasedRes.ok
		? await releasedRes.json()
		: { items: [], total: 0, page: 1, limit: CAROUSEL_LIMIT };
	const readData: ArchiveListResponse = readRes.ok
		? await readRes.json()
		: { items: [], total: 0, page: 1, limit: 100 };

	const recentlyRead = (readData.items ?? [])
		.filter((a) => a.progress !== null)
		.sort((a, b) => {
			const da = a.progress ? new Date(a.progress.last_read).getTime() : 0;
			const db = b.progress ? new Date(b.progress.last_read).getTime() : 0;
			return db - da;
		})
		.slice(0, CAROUSEL_LIMIT);

	return {
		recentlyAdded: addedData.items ?? [],
		recentlyReleased: (releasedData.items ?? []).filter((a) => a.release_date !== null),
		recentlyRead
	};
};
