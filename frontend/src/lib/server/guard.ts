import { fail, type ActionFailure } from '@sveltejs/kit';
import { can, type Permissions } from '$lib/auth/permissions';

/**
 * Guards a form action: returns a 403 failure when the current user lacks the
 * permission, or null to proceed.
 *
 * Every guarded action answers with the same `{ action, error }` payload so
 * the pages can render errors uniformly. A load has no error payload to
 * return - guard those with `can` directly and `error(403, …)`.
 */
export function guard<A extends string>(
	user: App.Locals['user'],
	permissions: Permissions,
	action: A,
	message: string
): ActionFailure<{ action: A; error: string }> | null {
	return can(user, permissions) ? null : fail(403, { action, error: message });
}
