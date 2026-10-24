import { z } from 'zod';

export const loginSchema = z.object({
	username: z.string().min(1),
	password: z.string().min(1),
	// Unchecked checkboxes submit nothing, so Superforms maps the missing key
	// to this default. `false` keeps the session cookie browser-bound.
	rememberMe: z.boolean().default(false)
});

export type LoginSchema = typeof loginSchema;
