import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';
import type { Library } from '$lib/types';

export const load: LayoutServerLoad = async ({ locals, fetch }) => {
	if (!locals.user) redirect(302, '/login');

	const librariesRes = await fetch('/api/libraries');
	const libraries: Library[] = librariesRes.ok ? await librariesRes.json() : [];

	return { user: locals.user, libraries };
};
