import type { Actions, PageServerLoad } from './$types';
import { fail } from '@sveltejs/kit';
import { auth } from '$lib/server/auth';
import { APIError } from 'better-auth';
import { setError, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { profileSchema } from '$lib/schemas/profile';
import { changePasswordSchema } from '$lib/schemas/change-password';

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
	}
};
