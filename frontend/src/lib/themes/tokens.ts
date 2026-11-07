// Canonical, ordered token list for the theme system. Taken verbatim from the
// `:root` block in `src/routes/layout.css` - 31 colour tokens plus `radius`.
// Everything else (the editor grouping, the DB/zod schema, the SSR serializer,
// the client switcher) is driven off this one list, so adding a token is a
// single-line change here.

export const COLOR_TOKENS = [
	'background',
	'foreground',
	'card',
	'card-foreground',
	'popover',
	'popover-foreground',
	'primary',
	'primary-foreground',
	'secondary',
	'secondary-foreground',
	'muted',
	'muted-foreground',
	'accent',
	'accent-foreground',
	'destructive',
	'border',
	'input',
	'ring',
	'chart-1',
	'chart-2',
	'chart-3',
	'chart-4',
	'chart-5',
	'sidebar',
	'sidebar-foreground',
	'sidebar-primary',
	'sidebar-primary-foreground',
	'sidebar-accent',
	'sidebar-accent-foreground',
	'sidebar-border',
	'sidebar-ring'
] as const;

export type ColorToken = (typeof COLOR_TOKENS)[number];

export const TOKEN_KEYS = [...COLOR_TOKENS, 'radius'] as const;

export type TokenKey = (typeof TOKEN_KEYS)[number];

export type ThemeTokens = Record<TokenKey, string>;

/** UI grouping for the token editor. Every colour token appears exactly once. */
export const TOKEN_GROUPS: { label: string; tokens: ColorToken[] }[] = [
	{
		label: 'Base surfaces',
		tokens: [
			'background',
			'foreground',
			'card',
			'card-foreground',
			'popover',
			'popover-foreground',
			'muted',
			'muted-foreground'
		]
	},
	{
		label: 'Brand',
		tokens: [
			'primary',
			'primary-foreground',
			'secondary',
			'secondary-foreground',
			'accent',
			'accent-foreground'
		]
	},
	{ label: 'Feedback', tokens: ['destructive'] },
	{ label: 'Form', tokens: ['border', 'input', 'ring'] },
	{
		label: 'Sidebar',
		tokens: [
			'sidebar',
			'sidebar-foreground',
			'sidebar-primary',
			'sidebar-primary-foreground',
			'sidebar-accent',
			'sidebar-accent-foreground',
			'sidebar-border',
			'sidebar-ring'
		]
	},
	{ label: 'Charts', tokens: ['chart-1', 'chart-2', 'chart-3', 'chart-4', 'chart-5'] }
];

export const TOKEN_KEY_SET: ReadonlySet<string> = new Set(TOKEN_KEYS);

// Token values land inside a server-rendered `<style>` block, so this is a
// security control, not a nicety: a value containing `}` or `</style>` would
// break out of the block. Allowlist an exact shape rather than escaping -
// `oklch()/hsl()/rgb()` with numeric-or-percent args (optionally an alpha
// after `/`), or a 3/6/8-digit hex.
const COLOR_VALUE_RE =
	/^(?:#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})|(?:oklch|hsl|rgb)\(\s*[0-9.\s%/+-]+\))$/;

const RADIUS_RE = /^\d+(?:\.\d+)?(?:rem|px)$/;

/** True when `value` is safe to inject as the value of custom property `key`. */
export function isValidTokenValue(key: string, value: unknown): value is string {
	if (typeof value !== 'string') return false;
	const v = value.trim();
	if (v.length === 0 || v.length > 64) return false;
	return key === 'radius' ? RADIUS_RE.test(v) : COLOR_VALUE_RE.test(v);
}
