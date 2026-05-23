import type { Handle } from '@sveltejs/kit';
import { building } from '$app/environment';
import { auth } from '$lib/server/auth';
import { svelteKitHandler } from 'better-auth/svelte-kit';

const BACKEND = 'http://localhost:8080';

export const handle: Handle = async ({ event, resolve }) => {
	const { pathname } = event.url;

	if (pathname.startsWith('/api/') && !pathname.startsWith('/api/auth/')) {
		const url = new URL(pathname + event.url.search, BACKEND);
		const headers = new Headers(event.request.headers);
		headers.delete('host');

		const response = await fetch(url, {
			method: event.request.method,
			headers,
			body: ['GET', 'HEAD'].includes(event.request.method) ? null : event.request.body,
			// @ts-expect-error duplex is required for streaming request bodies in Node
			duplex: 'half'
		});

		return new Response(response.body, {
			status: response.status,
			statusText: response.statusText,
			headers: response.headers
		});
	}

	const session = await auth.api.getSession({ headers: event.request.headers });
	if (session) {
		event.locals.session = session.session;
		event.locals.user = session.user;
	}

	return svelteKitHandler({ event, resolve, auth, building });
};
