import { error, fail } from '@sveltejs/kit';
import type { ActionFailure } from '@sveltejs/kit';
import { APIError } from 'better-auth';
import { eq } from 'drizzle-orm';
import type { Actions, PageServerLoad } from './$types';
import { auth } from '$lib/server/auth';
import { db } from '$lib/server/db';
import { user as userTable } from '$lib/server/db/auth.schema';
import { guard } from '$lib/server/guards';
import { ADMIN_ROLE, can, DEFAULT_ROLE } from '$lib/auth/permissions';

const PAGE_SIZE = 25;

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

// better-auth performs the write on its own connection, so the last-admin
// check and the mutation it guards can't share a transaction. Chaining the two
// guarded actions through this queue is what keeps two concurrent demotions
// from both passing the check and leaving the instance with no admin; a
// multi-process deployment would need a database-level guard instead.
let adminMutations: Promise<unknown> = Promise.resolve();

function serializeAdminMutation<T>(mutation: () => Promise<T>): Promise<T> {
	const result = adminMutations.then(mutation, mutation);

	adminMutations = result.then(
		() => undefined,
		() => undefined
	);

	return result;
}

/**
 * Maps a better-auth APIError to a form failure. Any other exception is a bug,
 * not a bad request, and is rethrown so the action fails loudly.
 */
function apiFailure(e: unknown, fallback: string): ActionFailure<{ error: string }> {
	if (e instanceof APIError) {
		return fail(e.statusCode, { error: e.body?.message ?? fallback });
	}

	throw e;
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
		search
	};
};

export const actions: Actions = {
	create: async ({ request, locals }) => {
		const denied = guard(locals.user, { user: ['create'] }, 'create', 'You cannot create users.');
		if (denied) return denied;

		const form = await request.formData();
		const username = String(form.get('username') ?? '').trim();
		const email = String(form.get('email') ?? '').trim();
		const password = String(form.get('password') ?? '');
		const role = String(form.get('role') ?? DEFAULT_ROLE);

		if (!username || !email) {
			return fail(400, { error: 'Username and email are required.' });
		}

		if (password.length < 8) {
			return fail(400, { error: 'Password must be at least 8 characters.' });
		}

		if (role !== DEFAULT_ROLE && role !== ADMIN_ROLE) {
			return fail(400, { error: 'Unknown role.' });
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
			return apiFailure(e, 'Failed to create the user.');
		}

		return { action: 'create' as const, success: true };
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
			return fail(400, { action: 'setRole' as const, error: 'You cannot change your own role.' });
		}

		return serializeAdminMutation(async () => {
			if (role !== ADMIN_ROLE && (await isLastAdmin(userId))) {
				return fail(409, {
					action: 'setRole' as const,
					error: 'This is the last admin - promote someone else first.'
				});
			}

			try {
				await auth.api.setRole({ headers: request.headers, body: { userId, role } });
			} catch (e) {
				return apiFailure(e, 'Failed to change the role.');
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
			return apiFailure(e, 'Failed to update the ban.');
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

		try {
			await auth.api.revokeUserSessions({
				headers: request.headers,
				body: { userId: String(form.get('id') ?? '') }
			});
		} catch (e) {
			return apiFailure(e, 'Failed to revoke sessions.');
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
				return fail(409, {
					action: 'delete' as const,
					error: 'This is the last admin - promote someone else first.'
				});
			}

			try {
				await auth.api.removeUser({ headers: request.headers, body: { userId } });
			} catch (e) {
				return apiFailure(e, 'Failed to delete the user.');
			}

			return { action: 'delete' as const, success: true };
		});
	}
};
