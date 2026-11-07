// Re-exported so existing server-side callers (+page.server.ts / actions)
// don't need to change their import path. The implementation itself has no
// server-only dependencies, so it also lives at $lib/api.ts for components
// that need it directly (SvelteKit forbids importing $lib/server/* from
// client code).
export { errorMessage } from '$lib/api';
