import type { PageServerLoad } from './$types';
import type { DuplicateGroup, Library } from '$lib/types';
import { errorMessage } from '$lib/server/api';

// Mirrors the backend's own clamp on GET /api/admin/duplicates - it silently
// falls back to 8 for anything out of range, so validating here keeps the
// threshold the UI displays honest about what was actually compared.
const DEFAULT_THRESHOLD = 8;
const MAX_THRESHOLD = 20;

export const load: PageServerLoad = async ({ fetch, url }) => {
	// Guard the absent/empty cases explicitly: Number(null) and Number('') are
	// both 0, which would otherwise pass the range check below and silently
	// scan at threshold 0 (exact hash matches only) instead of the default.
	const raw = url.searchParams.get('threshold')?.trim();
	const parsed = raw ? Number(raw) : Number.NaN;
	const threshold =
		Number.isInteger(parsed) && parsed >= 0 && parsed <= MAX_THRESHOLD
			? parsed
			: DEFAULT_THRESHOLD;

	const [duplicatesRes, librariesRes] = await Promise.all([
		fetch(`/api/admin/duplicates?threshold=${threshold}`),
		fetch('/api/libraries')
	]);

	const groups: DuplicateGroup[] = duplicatesRes.ok ? await duplicatesRes.json() : [];
	const libraries: Library[] = librariesRes.ok ? await librariesRes.json() : [];

	return {
		groups: groups ?? [],
		libraries,
		threshold,
		error: duplicatesRes.ok
			? null
			: await errorMessage(duplicatesRes, 'Failed to scan for duplicates.')
	};
};
