import type { PageServerLoad } from './$types';
import type { ArchiveListResponse } from '$lib/types';

const DEFAULT_LIMIT = 40;
const MAX_LIMIT = 100;

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1'));
	const limit = Math.min(
		MAX_LIMIT,
		Math.max(1, parseInt(url.searchParams.get('limit') ?? String(DEFAULT_LIMIT)) || DEFAULT_LIMIT)
	);

	const archivesRes = await fetch(`/api/archives/favorites?page=${page}&limit=${limit}`);
	const archivesData: ArchiveListResponse = archivesRes.ok
		? await archivesRes.json()
		: { items: [], total: 0, page, limit };

	return {
		archives: archivesData.items ?? [],
		total: Number(archivesData.total),
		page,
		limit
	};
};
