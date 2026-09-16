import { error, fail } from '@sveltejs/kit';
import { APIError } from 'better-auth';
import { and, count, eq } from 'drizzle-orm';
import type { Actions, PageServerLoad } from './$types';
import { auth } from '$lib/server/auth';
import { db } from '$lib/server/db';
import { user as userTable } from '$lib/server/db/auth.schema';
import { ADMIN_ROLE, can, DEFAULT_ROLE } from '$lib/auth/permissions';

const PAGE_SIZE = 25;

export type ManagedUser = {
	id: string;
	name: string;
	email: string;
	username?: string | null;
	image?: string | null;
	role?: string | null;
	banned?: boolean | null;
	createdAt: string | Date;
};

// Demoting or deleting the last admin would lock everyone out of the admin
// section with no way back short of editing the database by hand.
async function isLastAdmin(userId: string) {
	const [{ total }] = await db
		.select({ total: count() })
		.from(userTable)
		.where(eq(userTable.role, ADMIN_ROLE));

	if (total > 1) return false;

	const [{ total: matches }] = await db
		.select({ total: count() })
		.from(userTable)
		.where(and(eq(userTable.id, userId), eq(userTable.role, ADMIN_ROLE)));

	return matches > 0;
}

function apiFailure(e: unknown, fallback: string) {
	if (e instanceof APIError) return fail(e.statusCode, { error: e.body?.message ?? fallback });
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

	return {
		users: result.users as ManagedUser[],
		total: result.total,
		page,
		pageSize: PAGE_SIZE,
		search
	};
};

export const actions: Actions = {
	create: async ({ request, locals }) => {
		if (!can(locals.user, { user: ['create'] })) {
			return fail(403, { error: 'You cannot create users.' });
		}

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
		if (!can(locals.user, { user: ['set-role'] })) {
			return fail(403, { error: 'You cannot change roles.' });
		}

		const form = await request.formData();
		const userId = String(form.get('id') ?? '');
		const role = String(form.get('role') ?? '');

		if (userId === locals.user?.id) {
			return fail(400, { error: 'You cannot change your own role.' });
		}

		if (role !== ADMIN_ROLE && (await isLastAdmin(userId))) {
			return fail(409, { error: 'This is the last admin - promote someone else first.' });
		}

		try {
			await auth.api.setRole({ headers: request.headers, body: { userId, role } });
		} catch (e) {
			return apiFailure(e, 'Failed to change the role.');
		}

		return { action: 'setRole' as const, success: true };
	},

	setBanned: async ({ request, locals }) => {
		if (!can(locals.user, { user: ['ban'] })) {
			return fail(403, { error: 'You cannot ban users.' });
		}

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
		if (!can(locals.user, { session: ['revoke'] })) {
			return fail(403, { error: 'You cannot revoke sessions.' });
		}

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
		if (!can(locals.user, { user: ['delete'] })) {
			return fail(403, { error: 'You cannot delete users.' });
		}

		const form = await request.formData();
		const userId = String(form.get('id') ?? '');

		if (await isLastAdmin(userId)) {
			return fail(409, { error: 'This is the last admin - promote someone else first.' });
		}

		try {
			await auth.api.removeUser({ headers: request.headers, body: { userId } });
		} catch (e) {
			return apiFailure(e, 'Failed to delete the user.');
		}

		return { action: 'delete' as const, success: true };
	}
};
