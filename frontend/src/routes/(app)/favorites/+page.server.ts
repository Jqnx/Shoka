import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { ArchiveListResponse, AudioListResponse } from '$lib/types';

const DEFAULT_LIMIT = 40;
const MAX_LIMIT = 100;

function pageParam(raw: string | null): number {
	const parsed = parseInt(raw ?? '1', 10);

	return Number.isFinite(parsed) && parsed > 0 ? parsed : 1;
}

// A list can shrink (unfavoriting its last item) between visits; send the
// request to the last real page instead of rendering a false empty state.
function redirectToLastPage(
	url: URL,
	param: string,
	requested: number,
	total: number,
	limit: number
): void {
	const last = Math.max(1, Math.ceil(total / limit));
	if (requested <= last) return;

	const params = new URLSearchParams(url.searchParams);
	params.set(param, String(last));

	redirect(307, `${url.pathname}?${params}`);
}

export const load: PageServerLoad = async ({ fetch, url }) => {
	const page = pageParam(url.searchParams.get('page'));
	const audioPage = pageParam(url.searchParams.get('audio_page'));
	const limit = Math.min(
		MAX_LIMIT,
		Math.max(1, parseInt(url.searchParams.get('limit') ?? String(DEFAULT_LIMIT)) || DEFAULT_LIMIT)
	);

	const [archivesRes, audioRes] = await Promise.all([
		fetch(`/api/archives/favorites?page=${page}&limit=${limit}`),
		fetch(`/api/audio?favorites=true&page=${audioPage}&limit=${limit}`)
	]);

	const archivesData: ArchiveListResponse = archivesRes.ok
		? await archivesRes.json()
		: { items: [], total: 0, page, limit };
	const audioData: AudioListResponse = audioRes.ok
		? await audioRes.json()
		: { items: [], total: 0, page: audioPage, limit };

	const total = Number(archivesData.total);
	const audioTotal = Number(audioData.total);
	const audioLimit = Number(audioData.limit) || limit;

	// A list can shrink (unfavoriting its last item) between visits; send the
	// request to the last real page instead of rendering a false empty state.
	redirectToLastPage(url, 'page', page, total, limit);
	redirectToLastPage(url, 'audio_page', audioPage, audioTotal, audioLimit);

	return {
		archives: archivesData.items ?? [],
		total,
		page,
		limit,
		audio: audioData.items ?? [],
		audioTotal,
		audioPage,
		audioLimit
	};
};
