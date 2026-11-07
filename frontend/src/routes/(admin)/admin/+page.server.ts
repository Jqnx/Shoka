import type { PageServerLoad } from './$types';
import type { JobSnapshot, Library } from '$lib/types';
import { count } from 'drizzle-orm';
import { can } from '$lib/auth/permissions';
import { db } from '$lib/server/db';
import { user as userTable } from '$lib/server/db/auth.schema';

export const load: PageServerLoad = async ({ fetch, locals }) => {
	// Same permissions that hide these entries in admin-sidebar.svelte: a card
	// must not link somewhere its viewer would only ever get a 403 from.
	const canReadJobs = can(locals.user, { job: ['read'] });
	const canListUsers = can(locals.user, { user: ['list'] });

	const [res, jobsRes, accountRows] = await Promise.all([
		fetch('/api/libraries'),
		canReadJobs ? fetch('/api/admin/jobs') : null,
		canListUsers ? db.select({ total: count() }).from(userTable) : null
	]);

	const libraries: Library[] = res.ok ? await res.json() : [];

	// null counts mean the queue could not be read - the card says so instead
	// of reporting an idle queue.
	const jobCounts: JobSnapshot['counts'] | null = jobsRes?.ok
		? ((await jobsRes.json()) as JobSnapshot).counts
		: null;

	return {
		libraries,
		jobCounts,
		accounts: accountRows?.[0]?.total ?? 0,
		canReadJobs,
		canListUsers
	};
};
