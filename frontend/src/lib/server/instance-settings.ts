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
