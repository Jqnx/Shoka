import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';
import { DEFAULT_THEME_ID, PRESET_MAP } from '$lib/themes/presets';
import { userCount } from '$lib/server/first-run';

// The resolved active theme has to come from the *root* layout: the reader
// route uses `+page@.svelte`, which detaches from the (app) layout and skips
// its load entirely.
export const load: LayoutServerLoad = async ({ locals, url }) => {
	// First-run gate: with zero users there is nothing to sign in to and the
	// first account owns the instance, so funnel everything to the wizard.
	// The query only runs for signed-out requests — a session implies a user.
	if (!locals.user && url.pathname !== '/setup' && (await userCount()) === 0) {
		redirect(302, '/setup');
	}

	return { theme: locals.theme ?? PRESET_MAP[DEFAULT_THEME_ID] };
};
