import type { PageServerLoad } from './$types';
import type { Library } from '$lib/types';

export const load: PageServerLoad = async ({ fetch }) => {
	const res = await fetch('/api/libraries');
	const libraries: Library[] = res.ok ? await res.json() : [];

	return { libraries };
};
