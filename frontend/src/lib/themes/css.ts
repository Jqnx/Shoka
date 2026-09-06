import { TOKEN_KEYS, isValidTokenValue, type ThemeTokens } from './tokens';

export const THEME_STYLE_ID = 'shoka-theme';

/**
 * Emit the `<style>` block that carries the active theme, for injection into
 * the SSR'd `<head>` (no flash).
 *
 * Two deliberate choices:
 *  - Selector `html:root` (specificity 0,1,1) beats the plain `:root` (0,1,0)
 *    that `layout.css` defines. In dev, Vite injects `layout.css` into the head
 *    *after* the SSR head content, so an equal-specificity block would lose the
 *    cascade. The extra `html` makes the active theme win regardless of order.
 *  - Every value is re-validated as it is written. The zod schema already does
 *    this on save, but rows outlive schema changes - defence in depth for a
 *    string that lands unescaped inside a `<style>` element.
 */
export function serializeThemeCss(tokens: ThemeTokens): string {
	const decls: string[] = [];
	for (const key of TOKEN_KEYS) {
		const value = tokens[key];
		if (!isValidTokenValue(key, value)) continue;
		decls.push(`\t--${key}: ${value.trim()};`);
	}
	return `<style id="${THEME_STYLE_ID}">\nhtml:root {\n${decls.join('\n')}\n}\n</style>`;
}
