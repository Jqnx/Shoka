import { FileCode, Globe, Regex } from '@lucide/svelte';
import type { Component } from 'svelte';

/** Extra per-source config fields a library source's settings form should expose. */
export type SourceField = 'cookies' | 'api_key' | 'blocklists';

export type SourceInfo = {
	label: string;
	description: string;
	fields: SourceField[];
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
		description: 'Reads metadata already embedded in a ComicInfo.xml file inside each archive.',
		fields: [],
		icon: FileCode,
		color: 'bg-sky-600'
	},
	filename: {
		label: 'Filename',
		description: 'Parses tags, artists, and other metadata out of the archive filename.',
		fields: ['blocklists'],
		icon: Regex,
		color: 'bg-amber-600'
	},
	'e-hentai': {
		label: 'E-Hentai',
		description:
			'Looks up matching galleries on E-Hentai. A logged-in session cookie unlocks more results.',
		fields: ['cookies'],
		icon: Globe,
		color: 'bg-violet-600'
	},
	nhentai: {
		label: 'nHentai',
		description: 'Looks up matching galleries on nHentai using its API.',
		fields: ['api_key'],
		icon: Globe,
		color: 'bg-rose-600'
	}
};

export function sourceInfo(source: string): SourceInfo {
	return (
		METADATA_SOURCE_INFO[source] ?? {
			label: source,
			description: '',
			fields: ['cookies', 'api_key', 'blocklists'],
			icon: Globe,
			color: 'bg-neutral-600'
		}
	);
}
