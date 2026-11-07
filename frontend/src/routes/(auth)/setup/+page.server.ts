import type { Actions, PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { APIError } from 'better-auth';
import { setError, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { auth } from '$lib/server/auth';
import { errorMessage } from '$lib/server/api';
import { guard } from '$lib/server/guard';
import { userCount } from '$lib/server/first-run';
import { can } from '$lib/auth/permissions';
import { signupSchema } from '$lib/schemas/signup';
import type { Library, LibraryType } from '$lib/types';

export const load: PageServerLoad = async ({ locals, fetch }) => {
	if (locals.user) {
		// Only an admin can do the remaining steps; anyone else has nothing here.
		if (!can(locals.user, { library: ['create'] })) redirect(302, '/');

		const [librariesRes, typesRes] = await Promise.all([
			fetch('/api/libraries'),
			fetch('/api/libraries/types')
		]);
		const libraries: Library[] = librariesRes.ok ? await librariesRes.json() : [];
		const libraryTypes: LibraryType[] = typesRes.ok ? await typesRes.json() : [];

		return {
			user: locals.user,
			libraries,
			libraryTypes,
			form: await superValidate(zod4(signupSchema))
		};
	}

	// Signed out on /setup: the wizard only exists while the instance is empty.
	if ((await userCount()) > 0) redirect(302, '/login');

	return {
		user: null,
		libraries: [] as Library[],
		libraryTypes: [] as LibraryType[],
		form: await superValidate(zod4(signupSchema))
	};
};

export const actions: Actions = {
	createAccount: async (event) => {
		// The load redirects away once the instance has users, but a direct POST
		// to this action skips the load - the check has to live here too.
		if ((await userCount()) > 0) redirect(303, '/');

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
			// Anything that isn't an APIError is a bug, not a bad signup:
			// letting it fall out of the catch would present a genuine failure
			// as a successful signup, and the redirect below would land the
			// user on the home page, which bounces them to /login with no
			// explanation.
			if (!(e instanceof APIError)) throw e;

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

		// Stay on the wizard: the load now sees the session and renders step 2.
		redirect(302, '/setup');
	},

	createLibrary: async ({ request, fetch, locals }) => {
		const denied = guard(
			locals.user,
			{ library: ['create'] },
			'createLibrary',
			'You cannot create libraries.'
		);
		if (denied) return denied;

		const form = await request.formData();
		const name = String(form.get('name') ?? '').trim();
		const path = String(form.get('path') ?? '').trim();
		const type = String(form.get('type') ?? '').trim() || 'doujinshi';

		if (!name || !path) {
			return fail(400, { action: 'createLibrary' as const, error: 'Name and path are required.' });
		}

		const res = await fetch('/api/admin/libraries', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ name, path, type })
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'createLibrary' as const,
				error: await errorMessage(res, 'Failed to create library.')
			});
		}

		return { action: 'createLibrary' as const, success: true as const };
	}
};
