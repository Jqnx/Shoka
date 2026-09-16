import { betterAuth } from 'better-auth/minimal';
import { drizzleAdapter } from 'better-auth/adapters/drizzle';
import { sveltekitCookies } from 'better-auth/svelte-kit';
import { env } from '$env/dynamic/private';
import { getRequestEvent } from '$app/server';
import { db } from '$lib/server/db';
import { user as userTable } from '$lib/server/db/auth.schema';
import { admin, username } from 'better-auth/plugins';
import { ac, ADMIN_ROLE, DEFAULT_ROLE, roles } from '$lib/auth/permissions';
import { count } from 'drizzle-orm';

export const auth = betterAuth({
	baseURL: env.ORIGIN,
	secret: env.BETTER_AUTH_SECRET,
	database: drizzleAdapter(db, { provider: 'sqlite' }),
	emailAndPassword: { enabled: true },
	databaseHooks: {
		user: {
			create: {
				// Root-level hooks run after plugin hooks, so this overrides the
				// admin plugin's defaultRole and the very first account owns the
				// instance. Every later signup keeps `user`.
				before: async (user) => {
					const [{ total }] = await db.select({ total: count() }).from(userTable);
					if (total > 0) return;
					return { data: { ...user, role: ADMIN_ROLE } };
				}
			}
		}
	},
	plugins: [
		username(),
		admin({ ac, roles, defaultRole: DEFAULT_ROLE, adminRoles: [ADMIN_ROLE] }),
		sveltekitCookies(getRequestEvent) // make sure this is the last plugin in the array
	]
});
