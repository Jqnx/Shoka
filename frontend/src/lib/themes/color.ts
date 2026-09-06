import { converter, formatHex, parse, type Color } from 'culori';

const toOklch = converter('oklch');

function round(n: number, places: number): number {
	const f = 10 ** places;
	return Math.round(n * f) / f;
}

/**
 * Normalise any CSS colour culori understands (hex, rgb, hsl, oklch, named…) to
 * a canonical `oklch()` string, so presets and user-authored themes share one
 * format. Returns `null` when the input is not a parseable colour.
 */
export function toOklchString(value: string): string | null {
	const parsed = parse(value.trim());
	if (!parsed) return null;
	const o = toOklch(parsed);
	if (!o) return null;
	const l = round(o.l ?? 0, 4);
	const c = round(o.c ?? 0, 4);
	const h = o.h == null || Number.isNaN(o.h) ? 0 : round(o.h, 3);
	const a = o.alpha ?? 1;
	return a < 1 ? `oklch(${l} ${c} ${h} / ${round(a, 3)})` : `oklch(${l} ${c} ${h})`;
}

/** Best-effort conversion to `#rrggbb` for `<input type="color">`. */
export function toHex(value: string): string {
	const parsed = parse(value.trim());
	if (!parsed) return '#000000';
	return formatHex(parsed as Color) ?? '#000000';
}

/**
 * Rough WCAG-style relative luminance from an oklch/any colour, for cheap
 * contrast hints in the editor. Not a substitute for a real contrast check.
 */
export function relativeLuminance(value: string): number | null {
	const parsed = parse(value.trim());
	if (!parsed) return null;
	const rgb = converter('rgb')(parsed);
	if (!rgb) return null;
	const lin = (v: number) => (v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4);
	return 0.2126 * lin(rgb.r) + 0.7152 * lin(rgb.g) + 0.0722 * lin(rgb.b);
}

/** Contrast ratio (1–21) between two colours, or `null` if either won't parse. */
export function contrastRatio(a: string, b: string): number | null {
	const la = relativeLuminance(a);
	const lb = relativeLuminance(b);
	if (la == null || lb == null) return null;
	const [hi, lo] = la > lb ? [la, lb] : [lb, la];
	return (hi + 0.05) / (lo + 0.05);
}
