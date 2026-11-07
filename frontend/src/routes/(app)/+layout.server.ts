import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';
import type { Library } from '$lib/types';
import { PRESETS } from '$lib/themes/presets';
import { listCustomThemes } from '$lib/server/themes';

export const load: LayoutServerLoad = async ({ locals, fetch }) => {
	if (!locals.user) redirect(302, '/login');

	const librariesRes = await fetch('/api/libraries');
	const libraries: Library[] = librariesRes.ok ? await librariesRes.json() : [];

	// The selectable list - only the user menu and the settings pages need it.
	const customThemes = await listCustomThemes(locals.user.id);

	return { user: locals.user, libraries, presetThemes: PRESETS, customThemes };
};
