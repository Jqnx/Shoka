import type { LayoutServerLoad } from './$types';
import { DEFAULT_THEME_ID, PRESET_MAP } from '$lib/themes/presets';

// The resolved active theme has to come from the *root* layout: the reader
// route uses `+page@.svelte`, which detaches from the (app) layout and skips
// its load entirely.
export const load: LayoutServerLoad = ({ locals }) => {
	return { theme: locals.theme ?? PRESET_MAP[DEFAULT_THEME_ID] };
};
