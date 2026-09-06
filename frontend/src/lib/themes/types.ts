import type { ThemeTokens } from './tokens';

export type Appearance = 'light' | 'dark';

/**
 * A standalone palette. Built-ins and user-authored themes share this shape;
 * there is no inheritance and no light/dark pairing - `appearance` exists only
 * so the `.dark` class (for the `dark:` utilities inside `components/ui/`) and
 * `color-scheme` line up with the palette's brightness.
 */
export type Theme = {
	id: string;
	name: string;
	family: string;
	appearance: Appearance;
	tokens: ThemeTokens;
};
