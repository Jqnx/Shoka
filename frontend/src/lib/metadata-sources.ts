import { FileCode, Globe, Regex } from '@lucide/svelte';
import type { Component } from 'svelte';

/** Extra per-source config fields a library source's settings form should expose. */
export type SourceField = 'cookies' | 'api_key' | 'blocklists';

export type SourceInfo = {
	label: string;
	/** Optional help shown in the source's settings dialog. */
	help?: {
		text: string;
		href: string;
	};
	fields: SourceField[];
	/** True for filesystem-only sources: pinned ahead of every remote source. */
	local: boolean;
	/** Fallback tile look when no `image` is set: an icon on a plain color background. */
	icon: Component;
	color: string;
	/**
	 * Path to a logo/banner image for this source's tile, if one has been added
	 * under static/. None are bundled yet, so every source currently falls back
	 * to its icon + color tile.
	 */
	image?: string;
};

/**
 * Display metadata for Shoka's known metadata sources, mirroring what each
 * Source implementation in the backend (internal/metadata/sources) actually
 * reads off SourceSettings: ComicInfo uses none of it, filename uses only
 * the blocklists, e-hentai uses only cookies, nhentai uses only an API key.
 */
export const METADATA_SOURCE_INFO: Record<string, SourceInfo> = {
	comicinfo: {
		label: 'ComicInfo.xml',
		fields: [],
		local: true,
		icon: FileCode,
		color: 'bg-sky-600'
	},
	filename: {
		label: 'Filename',
		fields: ['blocklists'],
		local: true,
		icon: Regex,
		color: 'bg-amber-600'
	},
	'e-hentai': {
		label: 'E-Hentai',
		fields: ['cookies'],
		local: false,
		icon: Globe,
		color: 'bg-violet-600'
	},
	nhentai: {
		label: 'nHentai',
		help: {
			text: 'An API key can be created at this page:',
			href: 'https://nhentai.net/user/settings#apikeys'
		},
		fields: ['api_key'],
		local: false,
		icon: Globe,
		color: 'bg-rose-600'
	}
};

export function sourceInfo(source: string): SourceInfo {
	return (
		METADATA_SOURCE_INFO[source] ?? {
			label: source,
			fields: ['cookies', 'api_key', 'blocklists'],
			local: false,
			icon: Globe,
			color: 'bg-neutral-600'
		}
	);
}
