import { z } from 'zod';

// Admin edit forms on /admin/users/[id].

// The sign-in username, so the rules are the ones signInUsername enforces
// (3-30 characters, the plugin's default alphanumeric/dot/underscore
// validator): a value outside them would save fine but leave the account
// unable to sign in. The admin path bypasses the plugin's own endpoints, so
// these checks live here.
export const adminUsernameSchema = z.object({
	username: z
		.string()
		.trim()
		.min(3, { error: 'Username must be at least 3 characters' })
		.max(30, { error: 'Username must be at most 30 characters' })
		.regex(/^[a-zA-Z0-9_.]+$/, { error: 'Use letters, numbers, dots and underscores only' })
});

export type AdminUsernameSchema = typeof adminUsernameSchema;

// No current-password field: an admin sets a new one directly. The rules match
// $lib/schemas/change-password so an admin can't set a password the user then
// can't re-enter on their own settings form.
export const adminPasswordSchema = z
	.object({
		newPassword: z.string().min(8, { error: 'Password must be at least 8 characters' }).max(50),
		confirmPassword: z.string()
	})
	.refine((data) => data.newPassword === data.confirmPassword, {
		error: "Passwords don't match",
		path: ['confirmPassword']
	});

export type AdminPasswordSchema = typeof adminPasswordSchema;
