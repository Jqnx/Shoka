/** Extracts the `message` field from a backend `response.Error` JSON body, falling back if the body isn't JSON or doesn't have one. */
export async function errorMessage(res: Response, fallback: string): Promise<string> {
	try {
		const body = await res.json();
		return typeof body?.message === 'string' ? body.message : fallback;
	} catch {
		return fallback;
	}
}
