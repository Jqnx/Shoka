import { error, fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import type { Library, LibrarySource } from '$lib/types';
import { errorMessage } from '$lib/server/api';
import { guard } from '$lib/server/guard';
import { can } from '$lib/auth/permissions';

export const load: PageServerLoad = async ({ fetch, params, locals }) => {
	if (!can(locals.user, { library: ['read'] })) error(403, 'You cannot manage libraries.');

	const [librariesRes, sourcesRes] = await Promise.all([
		fetch('/api/libraries'),
		fetch(`/api/admin/libraries/${params.id}/sources`)
	]);

	const libraries: Library[] = librariesRes.ok ? await librariesRes.json() : [];
	const library = libraries.find((l) => l.id === params.id);
	if (!library) error(404, 'Library not found');

	const sources: LibrarySource[] = sourcesRes.ok ? await sourcesRes.json() : [];

	return { library, sources };
};

export const actions: Actions = {
	updateGeneral: async ({ request, fetch, params, locals }) => {
		const denied = guard(
			locals.user,
			{ library: ['update'] },
			'updateGeneral',
			'You cannot update libraries.'
		);
		if (denied) return denied;

		const form = await request.formData();
		const name = String(form.get('name') ?? '').trim();

		if (!name) {
			return fail(400, { action: 'updateGeneral' as const, error: 'Name cannot be empty.' });
		}

		const scanInterval = Number(form.get('scan_interval_minutes') ?? 0);

		if (!Number.isInteger(scanInterval) || scanInterval < 0) {
			return fail(400, { action: 'updateGeneral' as const, error: 'Invalid scan interval.' });
		}

		const res = await fetch(`/api/admin/libraries/${params.id}`, {
			method: 'PATCH',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				name,
				scan_interval_minutes: scanInterval,
				// Carried by a hidden input holding String(boolean), same as
				// the "enabled" toggle in library-source-card.svelte.
				watch_enabled: form.get('watch_enabled') === 'true',
				default_audio_folder: String(form.get('default_audio_folder') ?? '').trim()
			})
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'updateGeneral' as const,
				error: await errorMessage(res, 'Failed to update library.')
			});
		}

		return { action: 'updateGeneral' as const, success: true };
	},

	updateSource: async ({ request, fetch, params, locals }) => {
		const form = await request.formData();
		const source = String(form.get('source') ?? '');

		// Not using `guard`: the page renders this error on the matching source
		// card, so the payload has to carry the source name.
		if (!can(locals.user, { library: ['update'] })) {
			return fail(403, {
				action: 'updateSource' as const,
				source,
				error: 'You cannot update libraries.'
			});
		}

		const enabled = form.get('enabled') === 'true';
		const cookies = String(form.get('cookies') ?? '');
		const apiKey = String(form.get('api_key') ?? '');
		const magazineBlocklist = form.getAll('magazine_blocklist').map(String).filter(Boolean);
		const miscBlocklist = form.getAll('misc_blocklist').map(String).filter(Boolean);

		if (!source) {
			return fail(400, { action: 'updateSource' as const, source, error: 'Missing source.' });
		}

		const res = await fetch(`/api/admin/libraries/${params.id}/sources/${source}`, {
			method: 'PATCH',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				enabled,
				cookies,
				api_key: apiKey,
				magazine_blocklist: magazineBlocklist,
				misc_blocklist: miscBlocklist
			})
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'updateSource' as const,
				source,
				error: await errorMessage(res, 'Failed to update source.')
			});
		}

		return { action: 'updateSource' as const, source, success: true };
	},

	updateSourceOrder: async ({ request, fetch, params, locals }) => {
		const denied = guard(
			locals.user,
			{ library: ['update'] },
			'updateSourceOrder',
			'You cannot update libraries.'
		);
		if (denied) return denied;

		const form = await request.formData();
		const sources = String(form.get('order') ?? '')
			.split(',')
			.map((name) => name.trim())
			.filter(Boolean);

		const res = await fetch(`/api/admin/libraries/${params.id}/sources/order`, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ sources })
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'updateSourceOrder' as const,
				error: await errorMessage(res, 'Failed to update source order.')
			});
		}

		return { action: 'updateSourceOrder' as const, success: true };
	}
};
