import type { Actions, PageServerLoad } from './$types';
import { fail, type Cookies } from '@sveltejs/kit';
import { and, eq } from 'drizzle-orm';
import { auth } from '$lib/server/auth';
import { APIError } from 'better-auth';
import { setError, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { profileSchema } from '$lib/schemas/profile';
import { changePasswordSchema } from '$lib/schemas/change-password';
import { db } from '$lib/server/db';
import { userPreferences, userTheme } from '$lib/server/db/schema';
import { deleteThemeSchema, saveThemeSchema, selectThemeSchema } from '$lib/schemas/theme';
import type { ThemeTokens } from '$lib/themes/tokens';
import { COLOR_TOKENS } from '$lib/themes/tokens';
import { toOklchString } from '$lib/themes/color';
import { DEFAULT_THEME_ID, PRESET_MAP } from '$lib/themes/presets';
import {
	builtinIdForAppearance,
	customRowId,
	isCustomId,
	setThemeCookie
} from '$lib/server/themes';
import type { Appearance } from '$lib/themes/types';

const MAX_CUSTOM_THEMES = 20;

// The (app) layout's own load already redirects to /login when there's no
// session, and layout loads always run before this one - so locals.user is
// guaranteed to be set by the time we get here.
export const load: PageServerLoad = async ({ locals }) => {
	const user = locals.user!;

	return {
		profileForm: await superValidate(
			{ name: user.name, username: user.username ?? '' },
			zod4(profileSchema),
			{ id: 'profile' }
		),
		passwordForm: await superValidate(zod4(changePasswordSchema), { id: 'password' })
	};
};

/** Normalise every colour token to canonical `oklch()`; leave `radius` alone. */
function normalizeTokens(tokens: ThemeTokens): ThemeTokens {
	const out = { ...tokens };
	for (const key of COLOR_TOKENS) {
		out[key] = toOklchString(tokens[key]) ?? tokens[key];
	}
	return out;
}

async function upsertPreference(userId: string, themeId: string) {
	await db
		.insert(userPreferences)
		.values({ userId, themeId })
		.onConflictDoUpdate({
			target: userPreferences.userId,
			set: { themeId, updatedAt: new Date() }
		});
}

/** Keep the logged-out cookie pointed at a sensible built-in for `themeId`. */
async function syncCookie(cookies: Cookies, userId: string, themeId: string) {
	if (isCustomId(themeId)) {
		const [row] = await db
			.select({ appearance: userTheme.appearance })
			.from(userTheme)
			.where(and(eq(userTheme.id, customRowId(themeId)), eq(userTheme.userId, userId)))
			.limit(1);
		setThemeCookie(cookies, builtinIdForAppearance((row?.appearance ?? 'dark') as Appearance));
		return;
	}
	const preset = PRESET_MAP[themeId];
	setThemeCookie(cookies, preset ? preset.id : DEFAULT_THEME_ID);
}

export const actions: Actions = {
	updateProfile: async (event) => {
		const form = await superValidate(event, zod4(profileSchema), { id: 'profile' });

		if (!form.valid) {
			return fail(400, { profileForm: form });
		}

		try {
			await auth.api.updateUser({
				headers: event.request.headers,
				body: {
					name: form.data.name,
					username: form.data.username,
					displayUsername: form.data.username
				}
			});
		} catch (e) {
			if (e instanceof APIError) {
				setError(form, 'username', e.body?.message ?? 'Failed to update profile.');
				return fail(400, { profileForm: form });
			}
			throw e;
		}

		return { profileForm: form };
	},

	changePassword: async (event) => {
		const form = await superValidate(event, zod4(changePasswordSchema), { id: 'password' });

		if (!form.valid) {
			return fail(400, { passwordForm: form });
		}

		try {
			await auth.api.changePassword({
				headers: event.request.headers,
				body: {
					currentPassword: form.data.currentPassword,
					newPassword: form.data.newPassword
				}
			});
		} catch (e) {
			if (e instanceof APIError) {
				setError(form, 'currentPassword', e.body?.message ?? 'Failed to change password.');
				return fail(400, { passwordForm: form });
			}
			throw e;
		}

		// Don't echo the submitted password values back once they've actually
		// been applied - only the (now stale) validated shape is needed to
		// reset the form to its empty state.
		form.data.currentPassword = '';
		form.data.newPassword = '';
		form.data.confirmPassword = '';

		return { passwordForm: form };
	},

	selectTheme: async ({ request, locals, cookies }) => {
		const user = locals.user!;
		const parsed = selectThemeSchema.safeParse({
			themeId: (await request.formData()).get('themeId')
		});
		if (!parsed.success) return fail(400, { error: 'Invalid theme selection.' });

		const { themeId } = parsed.data;

		// Reject an unknown / other-user id rather than storing a dead pointer.
		if (isCustomId(themeId)) {
			const [row] = await db
				.select({ id: userTheme.id })
				.from(userTheme)
				.where(and(eq(userTheme.id, customRowId(themeId)), eq(userTheme.userId, user.id)))
				.limit(1);
			if (!row) return fail(404, { error: 'Theme not found.' });
		} else if (!PRESET_MAP[themeId]) {
			return fail(400, { error: 'Unknown theme.' });
		}

		await upsertPreference(user.id, themeId);
		await syncCookie(cookies, user.id, themeId);
		return { selected: themeId };
	},

	saveTheme: async ({ request, locals }) => {
		const user = locals.user!;
		const raw = (await request.formData()).get('payload');

		let json: unknown;
		try {
			json = JSON.parse(typeof raw === 'string' ? raw : '');
		} catch {
			return fail(400, { error: 'Malformed theme payload.' });
		}

		const parsed = saveThemeSchema.safeParse(json);
		if (!parsed.success) {
			return fail(400, {
				error: 'Theme failed validation.',
				issues: parsed.error.flatten().fieldErrors
			});
		}

		const { id, name, appearance } = parsed.data;
		const tokens = normalizeTokens(parsed.data.tokens as ThemeTokens);

		if (id) {
			const [updated] = await db
				.update(userTheme)
				.set({ name, appearance, tokens, updatedAt: new Date() })
				.where(and(eq(userTheme.id, id), eq(userTheme.userId, user.id)))
				.returning({ id: userTheme.id });
			if (!updated) return fail(404, { error: 'Theme not found.' });
			return { saved: `custom:${updated.id}` };
		}

		const count = await db.$count(userTheme, eq(userTheme.userId, user.id));
		if (count >= MAX_CUSTOM_THEMES) {
			return fail(400, { error: `You can keep at most ${MAX_CUSTOM_THEMES} custom themes.` });
		}

		const [created] = await db
			.insert(userTheme)
			.values({ userId: user.id, name, appearance, tokens })
			.returning({ id: userTheme.id });
		return { saved: `custom:${created.id}` };
	},

	deleteTheme: async ({ request, locals, cookies }) => {
		const user = locals.user!;
		const parsed = deleteThemeSchema.safeParse({
			id: (await request.formData()).get('id')
		});
		if (!parsed.success) return fail(400, { error: 'Invalid theme id.' });

		const [deleted] = await db
			.delete(userTheme)
			.where(and(eq(userTheme.id, parsed.data.id), eq(userTheme.userId, user.id)))
			.returning({ id: userTheme.id });
		if (!deleted) return fail(404, { error: 'Theme not found.' });

		// If the removed theme was selected, fall back to the default now.
		const [pref] = await db
			.select({ themeId: userPreferences.themeId })
			.from(userPreferences)
			.where(eq(userPreferences.userId, user.id))
			.limit(1);
		if (pref?.themeId === `custom:${parsed.data.id}`) {
			await upsertPreference(user.id, DEFAULT_THEME_ID);
			setThemeCookie(cookies, DEFAULT_THEME_ID);
			return { deleted: parsed.data.id, reselected: DEFAULT_THEME_ID };
		}

		return { deleted: parsed.data.id };
	}
};
