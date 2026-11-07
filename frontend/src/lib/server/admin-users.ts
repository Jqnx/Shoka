import { fail, type ActionFailure } from '@sveltejs/kit';
import { APIError } from 'better-auth';
import { and, eq, ne } from 'drizzle-orm';
import { db } from '$lib/server/db';
import { user as userTable } from '$lib/server/db/auth.schema';
import { ADMIN_ROLE } from '$lib/auth/permissions';

export const LAST_ADMIN_ERROR = 'This is the last admin - promote someone else first.';

// better-auth's username plugin only enforces uniqueness on /sign-up/email and
// /update-user, so an admin-set username would otherwise be guarded by the
// table's unique index alone - which surfaces as a 500, not a field error.
// `excludeUserId` lets a user keep their own name when the form is saved
// unchanged (including a case-only edit, since usernames are stored lowercase).
export async function isUsernameTaken(username: string, excludeUserId?: string) {
	const where = excludeUserId
		? and(eq(userTable.username, username.toLowerCase()), ne(userTable.id, excludeUserId))
		: eq(userTable.username, username.toLowerCase());

	const taken = await db.select({ id: userTable.id }).from(userTable).where(where).limit(1);

	return taken.length > 0;
}

// Demoting or deleting the last admin would lock everyone out of the admin
// section with no way back short of editing the database by hand. Asking for
// two admin ids answers that in a single query - getting both back means
// someone else can still get in.
export async function isLastAdmin(userId: string) {
	const admins = await db
		.select({ id: userTable.id })
		.from(userTable)
		.where(eq(userTable.role, ADMIN_ROLE))
		.limit(2);

	return admins.length === 1 && admins[0].id === userId;
}

// better-auth performs the write on its own connection, so the last-admin
// check and the mutation it guards can't share a transaction. This module is
// evaluated once per server, so the queue is process-wide: chaining the guarded
// mutations through it is what keeps two concurrent demotions - from the list
// or from a user's own page - from both passing the check and leaving the
// instance with no admin. A multi-process deployment would need a
// database-level guard instead.
let adminMutationQueue: Promise<unknown> = Promise.resolve();

export function serializeAdminMutation<T>(mutation: () => Promise<T>): Promise<T> {
	const result = adminMutationQueue.then(mutation, mutation);

	adminMutationQueue = result.then(
		() => undefined,
		() => undefined
	);

	return result;
}

/**
 * Maps a better-auth APIError to a failure tagged with the action and user that
 * triggered it, so both the list and the user page can render the message next
 * to the control. Anything else is a bug, not a bad request, and is rethrown so
 * the action fails loudly.
 */
export function apiFailure(
	e: unknown,
	fallback: string,
	action: string,
	id: string
): ActionFailure<{ action: string; id: string; error: string }> {
	if (!(e instanceof APIError)) throw e;

	return fail(e.statusCode, { action, id, error: e.body?.message ?? fallback });
}
