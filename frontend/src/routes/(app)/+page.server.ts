import type { PageServerLoad } from './$types';
import type { ArchiveListResponse } from '$lib/types';

export const load: PageServerLoad = async ({ fetch }) => {
	const res = await fetch('/api/archives?limit=100');
	if (!res.ok) return { archives: [] };
	const data: ArchiveListResponse = await res.json();
	return { archives: data.items ?? [] };
};
