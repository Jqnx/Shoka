import type { PageServerLoad } from './$types';
import { firstArchiveLibrary } from '$lib/types';
import type { Archive, AudioListItem } from '$lib/types';

const CAROUSEL_LIMIT = 10;

// A failed carousel request degrades to an empty row instead of failing the
// whole home page - one broken feed should not hide the others.
async function listItems<T>(res: Response | null): Promise<T[]> {
	if (!res?.ok) return [];

	const data: { items?: T[] } = await res.json();
	return data.items ?? [];
}

export const load: PageServerLoad = async ({ fetch, parent }) => {
	const { libraries } = await parent();
	// GetArchives is scoped to a single library; default to the first
	// doujinshi library until there's a "current library" concept for the home
	// page to draw from. Recently-read is its own endpoint and deliberately
	// spans every library, so it doesn't need libraryId at all.
	const libraryId = firstArchiveLibrary(libraries)?.id;
	// The audio list spans every audio library the caller can access, so
	// these carousels need no library id.
	const hasAudioLibraries = libraries.some((library) => library.type === 'audio');

	const [addedRes, releasedRes, readRes, audioPlayedRes, audioAddedRes, audioReleasedRes] =
		await Promise.all([
			libraryId
				? fetch(
						`/api/archives?library_id=${libraryId}&sort=created_at_desc&limit=${CAROUSEL_LIMIT}`
					)
				: null,
			libraryId
				? fetch(
						`/api/archives?library_id=${libraryId}&sort=release_date_desc&limit=${CAROUSEL_LIMIT}`
					)
				: null,
			fetch(`/api/archives/recently-read?limit=${CAROUSEL_LIMIT}`),
			hasAudioLibraries
				? fetch(`/api/audio?sort=played_at_desc&played=true&limit=${CAROUSEL_LIMIT}`)
				: null,
			hasAudioLibraries ? fetch(`/api/audio?sort=created_at_desc&limit=${CAROUSEL_LIMIT}`) : null,
			hasAudioLibraries ? fetch(`/api/audio?sort=release_date_desc&limit=${CAROUSEL_LIMIT}`) : null
		]);

	const [
		recentlyAdded,
		recentlyReleased,
		recentlyRead,
		audioRecentlyPlayed,
		audioRecentlyAdded,
		audioRecentlyReleased
	] = await Promise.all([
		listItems<Archive>(addedRes),
		listItems<Archive>(releasedRes),
		listItems<Archive>(readRes),
		listItems<AudioListItem>(audioPlayedRes),
		listItems<AudioListItem>(audioAddedRes),
		listItems<AudioListItem>(audioReleasedRes)
	]);

	return {
		recentlyAdded,
		recentlyReleased: recentlyReleased.filter((archive) => archive.release_date !== null),
		recentlyRead,
		audioRecentlyPlayed,
		audioRecentlyAdded,
		audioRecentlyReleased: audioRecentlyReleased.filter((audio) => audio.release_date !== null)
	};
};
