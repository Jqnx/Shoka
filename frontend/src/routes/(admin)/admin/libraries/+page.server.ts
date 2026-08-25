import { fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import type { Library, LibraryType } from '$lib/types';
import { errorMessage } from '$lib/server/api';

export const load: PageServerLoad = async ({ fetch }) => {
	const [librariesRes, typesRes] = await Promise.all([
		fetch('/api/libraries'),
		fetch('/api/libraries/types')
	]);

	const libraries: Library[] = librariesRes.ok ? await librariesRes.json() : [];
	const libraryTypes: LibraryType[] = typesRes.ok ? await typesRes.json() : [];

	return { libraries, libraryTypes };
};

export const actions: Actions = {
	create: async ({ request, fetch }) => {
		const form = await request.formData();
		const name = String(form.get('name') ?? '').trim();
		const path = String(form.get('path') ?? '').trim();
		const type = String(form.get('type') ?? '').trim() || 'doujinshi';

		if (!name || !path) {
			return fail(400, { action: 'create' as const, error: 'Name and path are required.' });
		}

		const res = await fetch('/api/admin/libraries', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ name, path, type })
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'create' as const,
				error: await errorMessage(res, 'Failed to create library.')
			});
		}

		return { action: 'create' as const, success: true };
	},

	scan: async ({ request, fetch }) => {
		const form = await request.formData();
		const id = String(form.get('id') ?? '');

		const res = await fetch(`/api/admin/libraries/${id}/scan`, { method: 'POST' });

		if (!res.ok) {
			return fail(res.status, {
				action: 'scan' as const,
				error: await errorMessage(res, 'Failed to start scan.')
			});
		}

		return { action: 'scan' as const, success: true };
	},

	delete: async ({ request, fetch }) => {
		const form = await request.formData();
		const id = String(form.get('id') ?? '');

		const res = await fetch(`/api/admin/libraries/${id}`, { method: 'DELETE' });

		if (!res.ok) {
			return fail(res.status, {
				action: 'delete' as const,
				error: await errorMessage(res, 'Failed to delete library.')
			});
		}

		return { action: 'delete' as const, success: true };
	}
};
