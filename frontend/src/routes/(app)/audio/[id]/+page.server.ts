import { error, fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { errorMessage } from '$lib/server/api';
import { guard } from '$lib/server/guard';
import type { ArchiveLanguage, Artist, Audio, Circle, Tag, VoiceActor } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, params }) => {
	const [audioRes, artistsRes, tagsRes, circlesRes, voiceActorsRes, languagesRes] =
		await Promise.all([
			fetch(`/api/audio/${params.id}`),
			fetch('/api/artists/all'),
			fetch('/api/tags/all'),
			fetch('/api/circles/all'),
			fetch('/api/voice-actors/all'),
			fetch('/api/archives/languages')
		]);

	if (audioRes.status === 404) error(404, 'Audio not found');
	if (!audioRes.ok) error(500, 'Failed to load audio');

	const audio: Audio = await audioRes.json();
	const allArtists: Artist[] = artistsRes.ok ? await artistsRes.json() : [];
	const allTags: Tag[] = tagsRes.ok ? await tagsRes.json() : [];
	const allCircles: Circle[] = circlesRes.ok ? await circlesRes.json() : [];
	const allVoiceActors: VoiceActor[] = voiceActorsRes.ok ? await voiceActorsRes.json() : [];
	const languages: ArchiveLanguage[] = languagesRes.ok ? await languagesRes.json() : [];

	return { audio, allArtists, allTags, allCircles, allVoiceActors, languages };
};

export const actions: Actions = {
	toggleFavorite: async ({ request, fetch, params }) => {
		const form = await request.formData();
		// PUT/DELETE are both idempotent, so this only picks the direction.
		const wasFavorited = form.get('favorited') === 'true';

		const res = await fetch(`/api/audio/${params.id}/favorite`, {
			method: wasFavorited ? 'DELETE' : 'PUT'
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'toggleFavorite' as const,
				error: await errorMessage(res, 'Failed to update favorite.')
			});
		}

		return { action: 'toggleFavorite' as const, success: true };
	},

	setRating: async ({ request, fetch, params }) => {
		const form = await request.formData();
		const rating = Number(form.get('rating'));
		if (!Number.isInteger(rating) || rating < 1 || rating > 5) {
			return fail(400, { action: 'setRating' as const, error: 'Rating must be between 1 and 5.' });
		}

		const res = await fetch(`/api/audio/${params.id}/rating`, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ rating })
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'setRating' as const,
				error: await errorMessage(res, 'Failed to set rating.')
			});
		}

		return { action: 'setRating' as const, success: true };
	},

	clearRating: async ({ fetch, params }) => {
		const res = await fetch(`/api/audio/${params.id}/rating`, { method: 'DELETE' });

		if (!res.ok) {
			return fail(res.status, {
				action: 'clearRating' as const,
				error: await errorMessage(res, 'Failed to clear rating.')
			});
		}

		return { action: 'clearRating' as const, success: true };
	},

	update: async ({ request, fetch, params, locals }) => {
		const denied = guard(locals.user, { audio: ['update'] }, 'update', 'You cannot edit audio.');
		if (denied) return denied;

		const form = await request.formData();
		const title = String(form.get('title') ?? '').trim();
		if (!title) {
			return fail(400, { action: 'update' as const, error: 'Title is required.' });
		}

		const releaseDate = String(form.get('release_date') ?? '');
		const language = String(form.get('language') ?? '');
		const circle = String(form.get('circle') ?? '');

		// Every array is a full replacement; the sheet round-trips the current
		// lists, so an emptied picker genuinely clears that relation.
		const res = await fetch(`/api/audio/${params.id}`, {
			method: 'PATCH',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				title,
				release_date: releaseDate || null,
				language: language || null,
				circle: circle || null,
				artists: form.getAll('artists').map(String),
				tags: form.getAll('tags').map(String),
				voice_actors: form.getAll('voice_actors').map(String)
			})
		});

		if (!res.ok) {
			return fail(res.status, {
				action: 'update' as const,
				error: await errorMessage(res, 'Failed to update audio.')
			});
		}

		return { action: 'update' as const, success: true };
	}
};
