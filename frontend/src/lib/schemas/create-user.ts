import { z } from 'zod';
import { ADMIN_ROLE, DEFAULT_ROLE } from '$lib/auth/permissions';

// The admin-created account is the same credential account /signup makes, so
// the field rules match $lib/schemas/signup.ts - minus the password
// confirmation, plus the role the admin picks. The one addition is the trim:
// signup is covered by the username plugin, which rejects a whitespace-only
// username, but that plugin only validates /sign-up/email and /update-user -
// so without it here an admin could create an account that can never sign in.
export const createUserSchema = z.object({
	username: z.string().trim().min(1, { error: 'Username is required' }).max(50),
	email: z.email('Not an email'),
	password: z.string().min(8, { error: 'Password must be at least 8 characters' }).max(50),
	role: z.enum([DEFAULT_ROLE, ADMIN_ROLE], { error: 'Pick a role' })
});

export type CreateUserSchema = typeof createUserSchema;
