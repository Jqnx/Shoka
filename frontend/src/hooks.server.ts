import type { Handle } from '@sveltejs/kit';
import { sequence } from '@sveltejs/kit/hooks';
import { building } from '$app/environment';
import { auth } from '$lib/server/auth';
import { svelteKitHandler } from 'better-auth/svelte-kit';
import { resolveTheme } from '$lib/server/themes';
import { serializeThemeCss } from '$lib/themes/css';

const BACKEND = 'http://localhost:8080';

// Every /api/* request (except /api/auth/*) is proxied to the Go backend. Any
// frontend-only endpoint must NOT live under /api/ - use a form action instead.
const handleProxy: Handle = async ({ event, resolve }) => {
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

	return resolve(event);
};

const handleAuth: Handle = async ({ event, resolve }) => {
	const session = await auth.api.getSession({ headers: event.request.headers });
	if (session) {
		event.locals.session = session.session;
		event.locals.user = session.user;
	}
	return resolve(event);
};

// Resolve the active theme (needs locals.user, set by handleAuth) and inject
// its CSS before the first byte. better-auth's svelteKitHandler calls
// resolve(event) itself with no options, so wrapping resolve is how we pass
// transformPageChunk through it.
const handleTheme: Handle = async ({ event, resolve }) => {
	const active = await resolveTheme(event.locals.user?.id, event.cookies);
	event.locals.theme = active;

	const style = serializeThemeCss(active.tokens);
	const htmlClass = active.appearance === 'dark' ? 'dark' : '';

	const themedResolve: typeof resolve = (ev) =>
		resolve(ev, {
			// Runs per streamed chunk, not once per document - guard it.
			transformPageChunk: ({ html }) =>
				html.includes('%shoka.')
					? html
							.replace('%shoka.style%', style)
							.replace('%shoka.class%', htmlClass)
							.replace('%shoka.scheme%', active.appearance)
					: html
		});

	return svelteKitHandler({ event, resolve: themedResolve, auth, building });
};

export const handle = sequence(handleProxy, handleAuth, handleTheme);
