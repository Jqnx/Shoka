import { z } from 'zod';

export const profileSchema = z.object({
	name: z.string().min(1, { error: 'Name is required' }).max(100),
	username: z.string().min(1, { error: 'Username is required' }).max(50)
});

export type ProfileSchema = typeof profileSchema;
