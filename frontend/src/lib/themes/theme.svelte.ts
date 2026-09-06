import { browser } from '$app/environment';
import { COLOR_TOKENS, isValidTokenValue, type ThemeTokens } from './tokens';
import type { Appearance, Theme } from './types';

function paint(tokens: Partial<ThemeTokens>, appearance: Appearance, id?: string) {
	if (!browser) return;
	const root = document.documentElement;
	for (const key of COLOR_TOKENS) {
		const value = tokens[key];
		if (isValidTokenValue(key, value)) root.style.setProperty(`--${key}`, value.trim());
	}
	if (isValidTokenValue('radius', tokens.radius)) {
		root.style.setProperty('--radius', tokens.radius.trim());
	}
	root.classList.toggle('dark', appearance === 'dark');
	root.style.colorScheme = appearance;
	if (id) root.dataset.theme = id;
}

/**
 * `$state`-backed holder for the active theme plus the DOM plumbing to switch
 * without a reload. Inline styles on `<html>` outrank the SSR'd `<style>`
 * block, and because the token list is fixed and every theme is complete there
 * are no stale properties to clear.
 *
 * It deliberately does NOT apply on mount - the SSR block is already correct,
 * so touching the DOM at hydration would only risk a flash. `apply()` /
 * `preview()` run on user action only.
 */
class ThemeController {
	#active = $state<Theme | undefined>(undefined);

	get active(): Theme | undefined {
		return this.#active;
	}

	/** Record the SSR-resolved theme without touching the DOM. */
	hydrate(theme: Theme) {
		this.#active = theme;
	}

	/** Switch to `theme` immediately (client-side). */
	apply(theme: Theme) {
		this.#active = theme;
		paint(theme.tokens, theme.appearance, theme.id);
	}

	/** Live-preview an in-progress token set (editor). Does not change `active`. */
	preview(tokens: Partial<ThemeTokens>, appearance: Appearance) {
		paint(tokens, appearance);
	}

	/** Re-apply the saved active theme, e.g. after the editor is cancelled. */
	restore() {
		if (this.#active) this.apply(this.#active);
	}

	/**
	 * Apply `next`, then persist. If persistence fails, revert so the UI never
	 * disagrees with the database. Resolves to whether persistence succeeded.
	 */
	async select(next: Theme, persist: () => Promise<boolean>): Promise<boolean> {
		const previous = this.#active;
		this.apply(next);
		const ok = await persist().catch(() => false);
		if (!ok && previous) this.apply(previous);
		return ok;
	}
}

export const theme = new ThemeController();
