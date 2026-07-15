import { fail } from '@sveltejs/kit';
import type { Actions } from './$types';

async function errorMessage(res: Response, fallback: string) {
	try {
		const body = await res.json();
		return typeof body?.message === 'string' ? body.message : fallback;
	} catch {
		return fallback;
	}
}

export const actions: Actions = {
	regenerateCovers: async ({ fetch }) => {
		const res = await fetch('/api/admin/covers', { method: 'POST' });

		if (!res.ok) {
			return fail(res.status, { error: await errorMessage(res, 'Failed to enqueue cover generation.') });
		}

		return { success: true };
	}
};
