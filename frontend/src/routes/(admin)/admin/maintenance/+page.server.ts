import { error, fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { errorMessage } from '$lib/server/api';
import { guard } from '$lib/server/guards';
import { can } from '$lib/auth/permissions';

export const load: PageServerLoad = async ({ locals }) => {
	if (!can(locals.user, { maintenance: ['read'] })) error(403, 'You cannot access maintenance.');
};

export const actions: Actions = {
	regenerateCovers: async ({ fetch, locals }) => {
		const denied = guard(
			locals.user,
			{ maintenance: ['run'] },
			'regenerateCovers',
			'You cannot run maintenance jobs.'
		);
		if (denied) return denied;

		const res = await fetch('/api/admin/covers', { method: 'POST' });

		if (!res.ok) {
			return fail(res.status, {
				action: 'regenerateCovers' as const,
				error: await errorMessage(res, 'Failed to enqueue cover generation.')
			});
		}

		return { action: 'regenerateCovers' as const, success: true };
	},

	generatePHashes: async ({ fetch, locals }) => {
		const denied = guard(
			locals.user,
			{ maintenance: ['run'] },
			'generatePHashes',
			'You cannot run maintenance jobs.'
		);
		if (denied) return denied;

		const res = await fetch('/api/admin/phashes', { method: 'POST' });

		if (!res.ok) {
			return fail(res.status, {
				action: 'generatePHashes' as const,
				error: await errorMessage(res, 'Failed to enqueue hash generation.')
			});
		}

		return { action: 'generatePHashes' as const, success: true };
	}
};
