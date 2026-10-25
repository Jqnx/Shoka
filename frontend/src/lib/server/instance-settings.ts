import { db } from '$lib/server/db';
import { instanceSettings } from '$lib/server/db/schema';

/** Whether public sign-ups are allowed. A missing row means open - the pre-feature default. */
export async function isSignupEnabled(): Promise<boolean> {
	const [row] = await db.select().from(instanceSettings).limit(1);
	return row?.signupEnabled ?? true;
}

export async function setSignupEnabled(enabled: boolean): Promise<void> {
	await db
		.insert(instanceSettings)
		.values({ id: 1, signupEnabled: enabled })
		.onConflictDoUpdate({ target: instanceSettings.id, set: { signupEnabled: enabled } });
}

export type MaintenanceJob = 'covers' | 'phashes';

/** When each maintenance action was last triggered. Null means never. */
export async function getMaintenanceRuns(): Promise<Record<MaintenanceJob, Date | null>> {
	const [row] = await db.select().from(instanceSettings).limit(1);

	return { covers: row?.coversLastRun ?? null, phashes: row?.phashesLastRun ?? null };
}

/** Stamps a maintenance action as triggered at `at` (now by default). */
export async function markMaintenanceRun(job: MaintenanceJob, at = new Date()): Promise<void> {
	const column = job === 'covers' ? { coversLastRun: at } : { phashesLastRun: at };

	await db
		.insert(instanceSettings)
		.values({ id: 1, ...column })
		.onConflictDoUpdate({ target: instanceSettings.id, set: column });
}
