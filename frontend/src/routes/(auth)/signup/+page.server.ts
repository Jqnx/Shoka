import type { Actions, PageServerLoad } from './$types';
import { redirect, fail } from '@sveltejs/kit';
import { auth } from '$lib/server/auth';
import { setError, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { signupSchema } from '$lib/schemas/signup';
import { APIError } from 'better-auth';

export const load: PageServerLoad = async (event) => {
	if (event.locals.user) {
		redirect(302, '/');
	}

	return {
		form: await superValidate(zod4(signupSchema))
	};
};

export const actions: Actions = {
	default: async (event) => {
		const form = await superValidate(event, zod4(signupSchema));
		const data = form.data;

		if (!form.valid) {
			return fail(400, { form });
		}

		try {
			await auth.api.signUpEmail({
				body: {
					email: data.email,
					password: data.password,
					name: data.username,
					username: data.username,
					displayUsername: data.username
				}
			});
		} catch (e: unknown) {
			if (e instanceof APIError) {
				switch (e.body?.code) {
					case 'USER_ALREADY_EXISTS_USE_ANOTHER_EMAIL':
						setError(form, 'email', `${e.body?.message}`);
						break;
					case 'USERNAME_IS_ALREADY_TAKEN_PLEASE_TRY_ANOTHER':
						setError(form, 'username', `${e.body?.message}`);
						break;
					default:
						return fail(500, { form });
				}
				return { form };
			}
		}

		redirect(302, '/');
	}
};
