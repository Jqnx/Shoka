import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { Archive } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, params }) => {
	const res = await fetch(`/api/archives/${params.id}`);
	if (res.status === 404) error(404, 'Archive not found');
	if (!res.ok) error(500, 'Failed to load archive');
	const archive: Archive = await res.json();
	return { archive };
};
