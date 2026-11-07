import { TOKEN_KEY_SET, isValidTokenValue, type ThemeTokens, type TokenKey } from './tokens';
import { toOklchString } from './color';

export type ImportResult = {
	tokens: ThemeTokens;
	matched: TokenKey[];
	rejected: string[];
};

const DECL_RE = /--([a-z0-9-]+)\s*:\s*([^;{}]+?)\s*(?:;|$)/gi;

/**
 * Ingest a shadcn / tweakcn `:root { … }` block. Scans for `--<known-token>:
 * <value>;`, runs every value through the same validator the schema uses,
 * ignores unknown properties, and leaves untouched tokens at their `base`
 * value (the theme being duplicated).
 */
export function parseThemeCss(css: string, base: ThemeTokens): ImportResult {
	const tokens: ThemeTokens = { ...base };
	const matched: TokenKey[] = [];
	const rejected: string[] = [];

	for (const m of css.matchAll(DECL_RE)) {
		const key = m[1].toLowerCase();
		if (!TOKEN_KEY_SET.has(key)) continue;
		const rawValue = m[2].trim();
		const value = key === 'radius' ? rawValue : (toOklchString(rawValue) ?? rawValue);
		if (!isValidTokenValue(key, value)) {
			rejected.push(key);
			continue;
		}
		tokens[key as TokenKey] = value;
		matched.push(key as TokenKey);
	}

	return { tokens, matched, rejected };
}
