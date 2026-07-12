import { redirect } from '@sveltejs/kit';
import { auth } from '$lib/server/auth';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	// Logging out is a POST-only action; visiting this route directly just bounces home.
	redirect(302, '/');
};

export const actions: Actions = {
	default: async (event) => {
		await auth.api.signOut({ headers: event.request.headers });
		redirect(303, '/login');
	}
};
