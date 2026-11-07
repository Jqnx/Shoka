import adapter from '@sveltejs/adapter-node';
import { relative, sep } from 'node:path';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		// defaults to rune mode for the project, except for `node_modules`. Can be removed in svelte 6.
		runes: ({ filename }) => {
			const relativePath = relative(import.meta.dirname, filename);
			const pathSegments = relativePath.toLowerCase().split(sep);
			const isExternalLibrary = pathSegments.includes('node_modules');

			return isExternalLibrary ? undefined : true;
		}
	},
	kit: {
		// Shoka is self-hosted, so the build target is always a plain Node server -
		// `node build` after `vite build`. adapter-auto detects the host from CI env
		// vars (VERCEL, CF_PAGES, ...); none are ever set here, so it would emit a
		// warning and produce no runnable output at all.
		adapter: adapter()
	}
};

export default config;
