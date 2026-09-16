import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { JobSnapshot } from '$lib/types';
import { errorMessage } from '$lib/server/api';
import { can } from '$lib/auth/permissions';

// Rendered until the SSE stream delivers its first snapshot (and whenever the
// initial fetch fails), so the page always has a well-formed shape to render.
const EMPTY_SNAPSHOT: JobSnapshot = {
	counts: { pending: 0, running: 0, done: 0, failed: 0 },
	active: [],
	recent_failures: []
};

export const load: PageServerLoad = async ({ fetch, locals }) => {
	if (!can(locals.user, { job: ['read'] })) error(403, 'You cannot view jobs.');

	const res = await fetch('/api/admin/jobs');

	return {
		snapshot: res.ok ? ((await res.json()) as JobSnapshot) : EMPTY_SNAPSHOT,
		error: res.ok ? null : await errorMessage(res, 'Failed to load jobs.')
	};
};
