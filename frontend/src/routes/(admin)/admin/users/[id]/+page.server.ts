import { error, fail } from '@sveltejs/kit';
import { APIError } from 'better-auth';
import type { UserWithRole } from 'better-auth/plugins';
import type { Actions, PageServerLoad } from './$types';
import type { Library } from '$lib/types';
import { auth } from '$lib/server/auth';
import { errorMessage } from '$lib/server/api';
import { guard } from '$lib/server/guard';
import { can } from '$lib/auth/permissions';

type LibraryAccess = { library_ids: string[] };

export const load: PageServerLoad = async ({ fetch, locals, params, request }) => {
	if (!can(locals.user, { user: ['get'] })) error(403, 'You cannot view users.');

	let user: UserWithRole;
	try {
		user = await auth.api.getUser({ query: { id: params.id }, headers: request.headers });
	} catch (e) {
		if (e instanceof APIError && e.statusCode === 404) error(404, 'User not found');
		throw e;
	}

	const [librariesRes, accessRes] = await Promise.all([
		fetch('/api/libraries'),
		fetch(`/api/admin/users/${params.id}/library-access`)
	]);

	if (!librariesRes.ok) error(librariesRes.status, 'Failed to load library access.');
	if (!accessRes.ok) error(accessRes.status, 'Failed to load library access.');

	const libraries: Library[] = await librariesRes.json();
	const access: LibraryAccess = await accessRes.json();

	// `target`, not `user`: the admin layout already exposes the signed-in
	// admin as `user`, and a same-named key here would shadow it.
	return { target: user, libraries, grantedLibraryIds: access.library_ids };
};

export const actions: Actions = {
	saveAccess: async ({ fetch, locals, params, request }) => {
		const denied = guard(
			locals.user,
			{ user: ['update'] },
			'saveAccess',
			'You cannot manage user access.'
		);
		if (denied) return denied;

		const form = await request.formData();
		const libraryIds = form.getAll('libraryIds').map(String);

		const res = await fetch(`/api/admin/users/${params.id}/library-access`, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ library_ids: libraryIds })
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'saveAccess' as const,
				error: await errorMessage(res, 'Failed to save library access.')
			});
		}

		return { action: 'saveAccess' as const, success: true };
	}
};
