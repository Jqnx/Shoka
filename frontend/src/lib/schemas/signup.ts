import { z } from 'zod';

export const signupSchema = z
	.object({
		username: z.string().min(1, { error: 'Username is required' }).max(50),
		email: z.email('Not an email'),
		password: z.string().min(8, { error: 'Password must be at least 8 characters' }).max(50),
		confirmPassword: z.string()
	})
	.refine((data) => data.password === data.confirmPassword, {
		error: "Passwords don't match",
		path: ['confirmPassword']
	});

export type SignupSchema = typeof signupSchema;
