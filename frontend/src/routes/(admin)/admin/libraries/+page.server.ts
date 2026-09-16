import { error, fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import type { Library, LibraryType } from '$lib/types';
import { errorMessage } from '$lib/server/api';
import { can } from '$lib/auth/permissions';

export const load: PageServerLoad = async ({ fetch, locals }) => {
	if (!can(locals.user, { library: ['read'] })) error(403, 'You cannot manage libraries.');

	const [librariesRes, typesRes] = await Promise.all([
		fetch('/api/libraries'),
		fetch('/api/libraries/types')
	]);

	const libraries: Library[] = librariesRes.ok ? await librariesRes.json() : [];
	const libraryTypes: LibraryType[] = typesRes.ok ? await typesRes.json() : [];

	return { libraries, libraryTypes };
};

export const actions: Actions = {
	create: async ({ request, fetch, locals }) => {
		if (!can(locals.user, { library: ['create'] })) {
			return fail(403, { action: 'create' as const, error: 'You cannot create libraries.' });
		}

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

	scan: async ({ request, fetch, locals }) => {
		if (!can(locals.user, { library: ['scan'] })) {
			return fail(403, { action: 'scan' as const, error: 'You cannot scan libraries.' });
		}

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

	regenerateCovers: async ({ request, fetch, locals }) => {
		if (!can(locals.user, { maintenance: ['run'] })) {
			return fail(403, {
				action: 'regenerateCovers' as const,
				error: 'You cannot run maintenance jobs.'
			});
		}

		const form = await request.formData();
		const id = String(form.get('id') ?? '');

		const res = await fetch(`/api/admin/libraries/${id}/covers`, { method: 'POST' });

		if (!res.ok) {
			return fail(res.status, {
				action: 'regenerateCovers' as const,
				error: await errorMessage(res, 'Failed to enqueue cover generation.')
			});
		}

		return { action: 'regenerateCovers' as const, success: true };
	},

	delete: async ({ request, fetch, locals }) => {
		if (!can(locals.user, { library: ['delete'] })) {
			return fail(403, { action: 'delete' as const, error: 'You cannot delete libraries.' });
		}

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
