import type { User, Session } from 'better-auth/minimal';

// The base `User` type doesn't know about the `username` plugin's extra
// columns (see auth.schema.ts / lib/server/auth.ts) - intersect them in here
// so `locals.user.username` etc. type-checks wherever it's read.
type AppUser = User & {
	username?: string | null;
	displayUsername?: string | null;
};

// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		interface Locals {
			user?: AppUser;
			session?: Session;
		}

		// interface Error {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
