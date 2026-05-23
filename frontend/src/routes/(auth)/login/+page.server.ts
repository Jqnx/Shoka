import type { Actions, PageServerLoad } from './$types';
import { redirect, fail } from '@sveltejs/kit';
import { auth } from '$lib/server/auth';
import { APIError } from 'better-auth';
import { setError, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { loginSchema } from '$lib/schemas/login';

export const load: PageServerLoad = async (event) => {
	if (event.locals.user) {
		redirect(302, '/');
	}

	return {
		form: await superValidate(zod4(loginSchema))
	};
};

export const actions: Actions = {
	default: async (event) => {
		const form = await superValidate(event, zod4(loginSchema));
		const data = form.data;

		if (!form.valid) {
			return fail(400, { form });
		}

		try {
			await auth.api.signInUsername({ body: { username: data.username, password: data.password } });
		} catch (e) {
			if (e instanceof APIError) {
				switch (e.body?.code) {
					case 'USERNAME_IS_INVALID':
						setError(form, 'username', 'Invalid username');
						break;
					default:
						setError(form, 'password', 'Invalid username or password');
				}
				return { form };
			}
		}

		redirect(302, '/');
	}
};
