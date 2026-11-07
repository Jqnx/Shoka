import type { Cookies } from '@sveltejs/kit';
import { and, eq } from 'drizzle-orm';
import { db } from '$lib/server/db';
import { userPreferences, userTheme } from '$lib/server/db/schema';
import { themeTokensSchema } from '$lib/schemas/theme';
import type { ThemeTokens } from '$lib/themes/tokens';
import { DEFAULT_LIGHT_THEME_ID, DEFAULT_THEME_ID, PRESET_MAP } from '$lib/themes/presets';
import type { Appearance, Theme } from '$lib/themes/types';

export const THEME_COOKIE = 'shoka_theme';
const CUSTOM_PREFIX = 'custom:';
const ONE_YEAR = 60 * 60 * 24 * 365;

export function isCustomId(id: string): boolean {
	return id.startsWith(CUSTOM_PREFIX);
}

export function customRowId(id: string): string {
	return id.slice(CUSTOM_PREFIX.length);
}

function defaultTheme(): Theme {
	return PRESET_MAP[DEFAULT_THEME_ID];
}

/** Turn a `user_theme` row into a Theme, or `null` if its tokens are corrupt. */
function rowToTheme(row: typeof userTheme.$inferSelect): Theme | null {
	const parsed = themeTokensSchema.safeParse(row.tokens);
	if (!parsed.success) return null;
	return {
		id: `${CUSTOM_PREFIX}${row.id}`,
		name: row.name,
		family: 'Custom',
		appearance: row.appearance as Appearance,
		tokens: parsed.data as ThemeTokens
	};
}

/**
 * Resolve the active theme before the first HTML byte.
 * Logged in  → `user_preferences.theme_id` (joined to `user_theme` for
 *              `custom:` ids). Logged out → the `shoka_theme` cookie.
 * Any id that no longer resolves falls back to the default - there is no
 * stale-pointer failure mode.
 */
export async function resolveTheme(userId: string | undefined, cookies: Cookies): Promise<Theme> {
	if (userId) {
		const [pref] = await db
			.select()
			.from(userPreferences)
			.where(eq(userPreferences.userId, userId))
			.limit(1);

		const themeId = pref?.themeId ?? DEFAULT_THEME_ID;

		if (isCustomId(themeId)) {
			const [row] = await db
				.select()
				.from(userTheme)
				.where(and(eq(userTheme.id, customRowId(themeId)), eq(userTheme.userId, userId)))
				.limit(1);
			return (row && rowToTheme(row)) || defaultTheme();
		}

		return PRESET_MAP[themeId] ?? defaultTheme();
	}

	const cookieId = cookies.get(THEME_COOKIE);
	return (cookieId && PRESET_MAP[cookieId]) || defaultTheme();
}

/** The user's custom themes, oldest first, skipping any with corrupt tokens. */
export async function listCustomThemes(userId: string): Promise<Theme[]> {
	const rows = await db
		.select()
		.from(userTheme)
		.where(eq(userTheme.userId, userId))
		.orderBy(userTheme.createdAt);
	return rows.map(rowToTheme).filter((t): t is Theme => t !== null);
}

/** The Shoka built-in id matching an appearance. */
export function builtinIdForAppearance(appearance: Appearance): string {
	return appearance === 'light' ? DEFAULT_LIGHT_THEME_ID : DEFAULT_THEME_ID;
}

/**
 * The built-in id to stash in the logged-out cookie for `theme`. Custom token
 * sets are too big for a cookie, so a custom selection stores the Shoka default
 * matching its appearance - keeps the login page at the right brightness.
 */
export function cookieIdFor(theme: Theme): string {
	if (!isCustomId(theme.id) && PRESET_MAP[theme.id]) return theme.id;
	return builtinIdForAppearance(theme.appearance);
}

export function setThemeCookie(cookies: Cookies, builtinId: string) {
	cookies.set(THEME_COOKIE, builtinId, {
		path: '/',
		httpOnly: true,
		sameSite: 'lax',
		maxAge: ONE_YEAR
	});
}
