// Canonical role/permission definition. The Go backend enforces the same
// matrix independently in backend/internal/auth/rbac.go; rbac_test.go parses
// the `statement` literal below and fails if the two disagree.
//
// Deliberately not under $lib/server/: components import `can` to hide
// controls they aren't allowed to use. Only the client-safe better-auth
// access helpers are imported here.
import { createAccessControl, type Role } from 'better-auth/plugins/access';
import { defaultStatements } from 'better-auth/plugins/admin/access';

export const statement = {
	...defaultStatements,
	archive: ['update', 'delete', 'cover', 'thumbnail'],
	metadata: ['fetch', 'apply', 'write'],
	taxonomy: ['create', 'update', 'delete'],
	library: ['read', 'create', 'update', 'delete', 'scan'],
	job: ['read'],
	maintenance: ['read', 'run']
} as const;

export const ac = createAccessControl(statement);

// Reading the library is the baseline of being authenticated, so `user` grants
// nothing: read routes are ungated and per-user data (progress, favourites,
// ratings, reader settings, themes) is scoped by user id rather than by role.
// The entry itself is still required - better-auth validates setRole and
// createUser against the keys of this map, so dropping it would make demoting
// an admin to `user` fail.
export const roles: Record<string, Role> = {
	user: ac.newRole({ user: [], session: [] }),
	// Admin holds the whole statement set, so derive it instead of repeating
	// the matrix.
	admin: ac.newRole({ ...statement })
};

export type Statement = typeof statement;
export type Permissions = { [R in keyof Statement]?: Statement[R][number][] };

export const DEFAULT_ROLE = 'user';
export const ADMIN_ROLE = 'admin';

// An empty or missing role yields no names, so `can` falls through to the same
// `false` an unknown role gets - `user` and `librarian` are both denied.
function roleNames(role: string | null | undefined): string[] {
	return (role ?? '')
		.split(',')
		.map((r) => r.trim())
		.filter(Boolean);
}

// Synchronous check against the local role map - no database round-trip, so it
// is safe to call during render. An unknown role grants nothing, which matches
// the Go side treating anything unrecognised as `user`.
export function can(
	user: { role?: string | null } | null | undefined,
	permissions: Permissions
): boolean {
	return roleNames(user?.role).some((name) => roles[name]?.authorize(permissions).success === true);
}

// Whether the user should see the admin section at all. Each admin page
// asserts its own permission on top of this.
export function canAccessAdmin(user: { role?: string | null } | null | undefined): boolean {
	return (
		can(user, { library: ['read'] }) ||
		can(user, { job: ['read'] }) ||
		can(user, { maintenance: ['read'] }) ||
		can(user, { user: ['list'] })
	);
}
