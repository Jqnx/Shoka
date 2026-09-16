// Canonical role/permission definition. The Go backend enforces the same
// matrix independently in backend/internal/auth/rbac.go - keep both in sync;
// a Go test (rbac_test.go) locks the statement list down.
//
// Deliberately not under $lib/server/: components import `can` to hide
// controls they aren't allowed to use. Only the client-safe better-auth
// access helpers are imported here.
import { createAccessControl, type Role } from 'better-auth/plugins/access';
import { adminAc, defaultStatements } from 'better-auth/plugins/admin/access';

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
export const roles: Record<string, Role> = {
	user: ac.newRole({
		user: [],
		session: [],
		archive: [],
		metadata: [],
		taxonomy: [],
		library: [],
		job: [],
		maintenance: []
	}),
	admin: ac.newRole({
		...adminAc.statements,
		archive: ['update', 'delete', 'cover', 'thumbnail'],
		metadata: ['fetch', 'apply', 'write'],
		taxonomy: ['create', 'update', 'delete'],
		library: ['read', 'create', 'update', 'delete', 'scan'],
		job: ['read'],
		maintenance: ['read', 'run']
	})
};

export type Statement = typeof statement;
export type Permissions = { [R in keyof Statement]?: Statement[R][number][] };

export const DEFAULT_ROLE = 'user';
export const ADMIN_ROLE = 'admin';

function roleNames(role: string | null | undefined): string[] {
	const names = (role ?? DEFAULT_ROLE)
		.split(',')
		.map((r) => r.trim())
		.filter(Boolean);

	return names.length > 0 ? names : [DEFAULT_ROLE];
}

// Synchronous check against the local role map - no database round-trip, so it
// is safe to call during render. An unknown role grants nothing, which matches
// the Go side treating anything unrecognised as `user`.
export function can(
	user: { role?: string | null } | null | undefined,
	permissions: Permissions
): boolean {
	return roleNames(user?.role).some(
		(name) => roles[name as keyof typeof roles]?.authorize(permissions).success === true
	);
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
