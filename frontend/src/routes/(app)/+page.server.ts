import type { PageServerLoad } from './$types';
import type { ArchiveListResponse } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, request }) => {
	const res = await fetch('http://localhost:8080/api/archives?limit=100', {
		headers: { cookie: request.headers.get('cookie') ?? '' }
	});
	if (!res.ok) return { archives: [] };
	const data: ArchiveListResponse = await res.json();
	return { archives: data.items ?? [] };
};
