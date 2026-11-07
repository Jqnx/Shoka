import { z } from 'zod';
import { TOKEN_KEYS, isValidTokenValue } from '$lib/themes/tokens';

const tokenEntries = TOKEN_KEYS.map((key) => [
	key,
	z.string().refine((v) => isValidTokenValue(key, v), { error: `Invalid value for --${key}` })
]);

/** Exactly the known token keys, each an allowlisted CSS value. */
export const themeTokensSchema = z.object(Object.fromEntries(tokenEntries)).strict();

export const appearanceSchema = z.enum(['light', 'dark']);

export const saveThemeSchema = z.object({
	// Absent = create, present = update an existing row.
	id: z.string().min(1).max(80).optional(),
	name: z.string().trim().min(1, { error: 'Name is required' }).max(40),
	appearance: appearanceSchema,
	tokens: themeTokensSchema
});

export const selectThemeSchema = z.object({
	themeId: z.string().min(1).max(80)
});

export const deleteThemeSchema = z.object({
	id: z.string().min(1).max(80)
});

export type SaveThemeInput = z.infer<typeof saveThemeSchema>;
