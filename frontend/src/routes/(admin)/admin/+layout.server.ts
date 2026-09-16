import { error, redirect } from '@sveltejs/kit';
import { canAccessAdmin } from '$lib/auth/permissions';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ locals }) => {
	if (!locals.user) redirect(302, '/login');
	if (!canAccessAdmin(locals.user)) error(403, 'You do not have access to the admin area.');
	return { user: locals.user };
};
