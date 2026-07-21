import { z } from 'zod';

export const changePasswordSchema = z
	.object({
		currentPassword: z.string().min(1, { error: 'Current password is required' }),
		newPassword: z.string().min(8, { error: 'Password must be at least 8 characters' }).max(50),
		confirmPassword: z.string()
	})
	.refine((data) => data.newPassword === data.confirmPassword, {
		error: "Passwords don't match",
		path: ['confirmPassword']
	});

export type ChangePasswordSchema = typeof changePasswordSchema;
