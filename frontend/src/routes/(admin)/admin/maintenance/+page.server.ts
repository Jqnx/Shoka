import { fail } from '@sveltejs/kit';
import type { Actions } from './$types';
import { errorMessage } from '$lib/server/api';

export const actions: Actions = {
	regenerateCovers: async ({ fetch }) => {
		const res = await fetch('/api/admin/covers', { method: 'POST' });

		if (!res.ok) {
			return fail(res.status, {
				action: 'regenerateCovers' as const,
				error: await errorMessage(res, 'Failed to enqueue cover generation.')
			});
		}

		return { action: 'regenerateCovers' as const, success: true };
	},

	generatePHashes: async ({ fetch }) => {
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
