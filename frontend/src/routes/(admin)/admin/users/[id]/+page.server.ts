import { error, fail } from '@sveltejs/kit';
import { APIError } from 'better-auth';
import type { UserWithRole } from 'better-auth/plugins';
import type { Actions, PageServerLoad } from './$types';
import type { Library } from '$lib/types';
import { auth } from '$lib/server/auth';
import { db } from '$lib/server/db';
import { user as userTable } from '$lib/server/db/auth.schema';
import { eq } from 'drizzle-orm';
import { errorMessage } from '$lib/server/api';
import { guard } from '$lib/server/guard';
import {
	apiFailure,
	isLastAdmin,
	isUsernameTaken,
	LAST_ADMIN_ERROR,
	serializeAdminMutation
} from '$lib/server/admin-users';
import { ADMIN_ROLE, DEFAULT_ROLE, can } from '$lib/auth/permissions';
import { setError, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { adminPasswordSchema, adminUsernameSchema } from '$lib/schemas/admin-user';

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

	// The admin plugin's user type carries no username, and the rename form has
	// to start from the stored value.
	const [identity] = await db
		.select({ username: userTable.username })
		.from(userTable)
		.where(eq(userTable.id, params.id))
		.limit(1);

	// `target`, not `user`: the admin layout already exposes the signed-in
	// admin as `user`, and a same-named key here would shadow it.
	return {
		target: user,
		libraries,
		grantedLibraryIds: access.library_ids,
		usernameForm: await superValidate(
			{ username: identity?.username ?? '' },
			zod4(adminUsernameSchema),
			{ id: 'username' }
		),
		passwordForm: await superValidate(zod4(adminPasswordSchema), { id: 'password' })
	};
};

export const actions: Actions = {
	updateUsername: async ({ request, locals, params }) => {
		const denied = guard(
			locals.user,
			{ user: ['update'] },
			'updateUsername',
			'You cannot edit users.'
		);
		if (denied) return denied;

		const form = await superValidate(request, zod4(adminUsernameSchema), { id: 'username' });

		if (!form.valid) {
			return fail(400, { usernameForm: form });
		}

		const { username } = form.data;

		// The uniqueness check and the write are one step: better-auth writes on
		// its own connection, so without the queue a second rename landing between
		// the check and the write would reach the unique index and answer 500
		// instead of a field error.
		return serializeAdminMutation(async () => {
			if (await isUsernameTaken(username, params.id)) {
				setError(form, 'username', 'Username is already taken. Please try another.');
				return fail(400, { usernameForm: form });
			}

			try {
				// The plugin normalises `username` to lowercase on write; the
				// typed casing is kept for display.
				await auth.api.adminUpdateUser({
					headers: request.headers,
					body: { userId: params.id, data: { username, displayUsername: username } }
				});
			} catch (e) {
				if (e instanceof APIError) {
					setError(form, 'username', e.body?.message ?? 'Failed to change the username.');
					return fail(e.statusCode, { usernameForm: form });
				}
				throw e;
			}

			return { usernameForm: form };
		});
	},

	setPassword: async ({ request, locals, params }) => {
		const denied = guard(
			locals.user,
			{ user: ['set-password'] },
			'setPassword',
			'You cannot change passwords.'
		);
		if (denied) return denied;

		const form = await superValidate(request, zod4(adminPasswordSchema), { id: 'password' });

		if (!form.valid) {
			return fail(400, { passwordForm: form });
		}

		try {
			await auth.api.setUserPassword({
				headers: request.headers,
				body: { userId: params.id, newPassword: form.data.newPassword }
			});
		} catch (e) {
			if (e instanceof APIError) {
				setError(form, 'newPassword', e.body?.message ?? 'Failed to set the password.');
				return fail(e.statusCode, { passwordForm: form });
			}
			throw e;
		}

		// Don't echo the applied password back; clearing the fields resets the
		// form on the client after a successful change.
		form.data.newPassword = '';
		form.data.confirmPassword = '';

		return { passwordForm: form };
	},

	setRole: async ({ request, locals, params }) => {
		const denied = guard(
			locals.user,
			{ user: ['set-role'] },
			'setRole',
			'You cannot change roles.'
		);
		if (denied) return denied;

		const userId = params.id;
		const role = String((await request.formData()).get('role') ?? '');

		if (role !== ADMIN_ROLE && role !== DEFAULT_ROLE) {
			return fail(400, { action: 'setRole' as const, id: userId, error: 'Unknown role.' });
		}

		// The page hides the toggle on one's own row, but a crafted POST still
		// gets the same answer - same rule as the users list.
		if (userId === locals.user?.id) {
			return fail(400, {
				action: 'setRole' as const,
				id: userId,
				error: 'You cannot change your own role.'
			});
		}

		return serializeAdminMutation(async () => {
			if (role !== ADMIN_ROLE && (await isLastAdmin(userId))) {
				return fail(409, { action: 'setRole' as const, id: userId, error: LAST_ADMIN_ERROR });
			}

			try {
				await auth.api.setRole({ headers: request.headers, body: { userId, role } });
			} catch (e) {
				return apiFailure(e, 'Failed to change the role.', 'setRole', userId);
			}

			return { action: 'setRole' as const, success: true };
		});
	},

	setBanned: async ({ request, locals, params }) => {
		const denied = guard(locals.user, { user: ['ban'] }, 'setBanned', 'You cannot ban users.');
		if (denied) return denied;

		const userId = params.id;
		const banned = String((await request.formData()).get('banned') ?? '') === 'true';

		// Banning deletes the user's sessions, so doing it to yourself would sign
		// you out of an account you can no longer sign back into.
		if (userId === locals.user?.id) {
			return fail(400, {
				action: 'setBanned' as const,
				id: userId,
				error: 'You cannot disable your own account.'
			});
		}

		try {
			// Banning also deletes the user's sessions, so it takes effect
			// immediately on both the frontend and the Go API.
			if (banned) {
				await auth.api.banUser({ headers: request.headers, body: { userId } });
			} else {
				await auth.api.unbanUser({ headers: request.headers, body: { userId } });
			}
		} catch (e) {
			return apiFailure(e, 'Failed to update the ban.', 'setBanned', userId);
		}

		return { action: 'setBanned' as const, success: true };
	},

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
