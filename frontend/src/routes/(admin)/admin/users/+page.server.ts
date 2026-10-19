import { error, fail } from '@sveltejs/kit';
import type { ActionFailure } from '@sveltejs/kit';
import { APIError } from 'better-auth';
import { eq } from 'drizzle-orm';
import type { Actions, PageServerLoad } from './$types';
import { auth } from '$lib/server/auth';
import { db } from '$lib/server/db';
import { user as userTable } from '$lib/server/db/auth.schema';
import { guard } from '$lib/server/guard';
import { ADMIN_ROLE, can } from '$lib/auth/permissions';
import { setError, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { createUserSchema } from '$lib/schemas/create-user';
import { isSignupEnabled, setSignupEnabled } from '$lib/server/instance-settings';

const PAGE_SIZE = 25;

const LAST_ADMIN_ERROR = 'This is the last admin - promote someone else first.';

export type ManagedUser = {
	id: string;
	name: string;
	email: string;
	image?: string | null;
	role?: string | null;
	banned?: boolean | null;
	createdAt: Date;
};

// Demoting or deleting the last admin would lock everyone out of the admin
// section with no way back short of editing the database by hand. Asking for
// two admin ids answers that in a single query - getting both back means
// someone else can still get in.
async function isLastAdmin(userId: string) {
	const admins = await db
		.select({ id: userTable.id })
		.from(userTable)
		.where(eq(userTable.role, ADMIN_ROLE))
		.limit(2);

	return admins.length === 1 && admins[0].id === userId;
}

// better-auth's username plugin only enforces uniqueness on /sign-up/email and
// /update-user, so an admin-created username would otherwise be guarded by the
// table's unique index alone - which surfaces as a 500, not a field error.
async function isUsernameTaken(username: string) {
	const taken = await db
		.select({ id: userTable.id })
		.from(userTable)
		.where(eq(userTable.username, username.toLowerCase()))
		.limit(1);

	return taken.length > 0;
}

// better-auth performs the write on its own connection, so the last-admin
// check and the mutation it guards can't share a transaction. Route modules are
// evaluated once per server, so this queue is process-wide: chaining the two
// guarded mutations through it is what keeps two concurrent demotions from both
// passing the check and leaving the instance with no admin. A multi-process
// deployment would need a database-level guard instead.
let adminMutationQueue: Promise<unknown> = Promise.resolve();

function serializeAdminMutation<T>(mutation: () => Promise<T>): Promise<T> {
	const result = adminMutationQueue.then(mutation, mutation);

	adminMutationQueue = result.then(
		() => undefined,
		() => undefined
	);

	return result;
}

/**
 * Maps a better-auth APIError to a failure tagged with the row that triggered
 * it, so the page can put the message on that row rather than at the top of the
 * list. Anything else is a bug, not a bad request, and is rethrown so the action
 * fails loudly.
 */
function apiFailure(
	e: unknown,
	fallback: string,
	action: string,
	id: string
): ActionFailure<{ action: string; id: string; error: string }> {
	if (!(e instanceof APIError)) throw e;

	return fail(e.statusCode, { action, id, error: e.body?.message ?? fallback });
}

export const load: PageServerLoad = async ({ locals, request, url }) => {
	if (!can(locals.user, { user: ['list'] })) error(403, 'You cannot manage users.');

	const search = url.searchParams.get('q')?.trim() ?? '';
	const page = Math.max(1, Number(url.searchParams.get('page') ?? 1) || 1);

	const result = await auth.api.listUsers({
		headers: request.headers,
		query: {
			limit: PAGE_SIZE,
			offset: (page - 1) * PAGE_SIZE,
			sortBy: 'createdAt',
			sortDirection: 'desc',
			...(search ? { searchValue: search, searchField: 'name' as const } : {})
		}
	});

	const users: ManagedUser[] = result.users.map((user) => ({
		id: user.id,
		name: user.name,
		email: user.email,
		image: user.image ?? null,
		role: user.role ?? null,
		banned: user.banned ?? null,
		createdAt: user.createdAt
	}));

	return {
		users,
		total: result.total,
		page,
		pageSize: PAGE_SIZE,
		search,
		signupEnabled: await isSignupEnabled(),
		canToggleSignup: can(locals.user, { user: ['update'] }),
		createForm: await superValidate(zod4(createUserSchema), { id: 'createUser' })
	};
};

export const actions: Actions = {
	create: async ({ request, locals }) => {
		const denied = guard(locals.user, { user: ['create'] }, 'create', 'You cannot create users.');
		if (denied) return denied;

		const form = await superValidate(request, zod4(createUserSchema), { id: 'createUser' });

		if (!form.valid) {
			return fail(400, { createForm: form });
		}

		const { username, email, password, role } = form.data;

		// The uniqueness check and the insert are one step: they aren't a
		// transaction (better-auth writes on its own connection), so without this
		// a second create for the same username landing between the check and the
		// write would reach the user_username_unique index and answer 500 instead
		// of a field error. Queued behind any earlier create, it sees that one in
		// its own check. The index stays the thing that enforces this.
		return serializeAdminMutation(async () => {
			if (await isUsernameTaken(username)) {
				setError(form, 'username', 'Username is already taken. Please try another.');

				return fail(400, { createForm: form });
			}

			try {
				await auth.api.createUser({
					headers: request.headers,
					body: {
						email,
						password,
						name: username,
						role,
						data: { username, displayUsername: username }
					}
				});
			} catch (e) {
				// Unlike the row actions this one reports into the dialog's own
				// superform, so it maps the two conflicts it knows itself rather
				// than going through apiFailure.
				if (!(e instanceof APIError)) throw e;

				switch (e.body?.code) {
					case 'USER_ALREADY_EXISTS_USE_ANOTHER_EMAIL':
						setError(form, 'email', `${e.body.message}`);
						break;
					case 'USERNAME_IS_ALREADY_TAKEN_PLEASE_TRY_ANOTHER':
						setError(form, 'username', `${e.body.message}`);
						break;
					default:
						// Tagged with the action so the page can show it inside
						// the dialog: it stays open on a failure, and the
						// page-level slot is behind the modal.
						return fail(e.statusCode, {
							action: 'create' as const,
							error: e.body?.message ?? 'Failed to create the user.'
						});
				}

				return fail(e.statusCode, { createForm: form });
			}

			return { createForm: form };
		});
	},

	setSignupEnabled: async ({ request, locals }) => {
		const denied = guard(
			locals.user,
			{ user: ['update'] },
			'setSignupEnabled',
			'You cannot change sign-up settings.'
		);
		if (denied) return denied;

		const form = await request.formData();
		await setSignupEnabled(String(form.get('enabled')) === 'true');
		return { success: true };
	},

	setRole: async ({ request, locals }) => {
		const denied = guard(
			locals.user,
			{ user: ['set-role'] },
			'setRole',
			'You cannot change roles.'
		);
		if (denied) return denied;

		const form = await request.formData();
		const userId = String(form.get('id') ?? '');
		const role = String(form.get('role') ?? '');

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

	setBanned: async ({ request, locals }) => {
		const denied = guard(locals.user, { user: ['ban'] }, 'setBanned', 'You cannot ban users.');
		if (denied) return denied;

		const form = await request.formData();
		const userId = String(form.get('id') ?? '');

		try {
			// Banning also deletes the user's sessions, so it takes effect
			// immediately on both the frontend and the Go API.
			if (form.get('banned') === 'true') {
				await auth.api.banUser({ headers: request.headers, body: { userId } });
			} else {
				await auth.api.unbanUser({ headers: request.headers, body: { userId } });
			}
		} catch (e) {
			return apiFailure(e, 'Failed to update the ban.', 'setBanned', userId);
		}

		return { action: 'setBanned' as const, success: true };
	},

	revokeSessions: async ({ request, locals }) => {
		const denied = guard(
			locals.user,
			{ session: ['revoke'] },
			'revokeSessions',
			'You cannot revoke sessions.'
		);
		if (denied) return denied;

		const form = await request.formData();
		const userId = String(form.get('id') ?? '');

		try {
			await auth.api.revokeUserSessions({
				headers: request.headers,
				body: { userId }
			});
		} catch (e) {
			return apiFailure(e, 'Failed to revoke sessions.', 'revokeSessions', userId);
		}

		return { action: 'revokeSessions' as const, success: true };
	},

	delete: async ({ request, locals }) => {
		const denied = guard(locals.user, { user: ['delete'] }, 'delete', 'You cannot delete users.');
		if (denied) return denied;

		const form = await request.formData();
		const userId = String(form.get('id') ?? '');

		return serializeAdminMutation(async () => {
			if (await isLastAdmin(userId)) {
				return fail(409, { action: 'delete' as const, id: userId, error: LAST_ADMIN_ERROR });
			}

			try {
				await auth.api.removeUser({ headers: request.headers, body: { userId } });
			} catch (e) {
				return apiFailure(e, 'Failed to delete the user.', 'delete', userId);
			}

			return { action: 'delete' as const, success: true };
		});
	}
};
