import { count } from 'drizzle-orm';
import { db } from '$lib/server/db';
import { user as userTable } from '$lib/server/db/auth.schema';

/** Number of registered users — 0 means the instance has never been set up. */
export async function userCount(): Promise<number> {
	const [{ total }] = await db.select({ total: count() }).from(userTable);
	return total;
}
