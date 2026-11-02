import type { AudioListItem, AudioListResponse, Library } from '$lib/types';

// One audio section on a taxonomy page, scoped to a single audio library so
// the section's count and its "view all" link always describe the same set.
export type AudioSectionData = {
	library: Library;
	works: AudioListItem[];
	total: number;
};

// loadAudioSections fetches the first page of an audio filter separately per
// accessible audio library. Audio libraries are browsed per library now, so an
// aggregate section could advertise more works than its link would show.
export async function loadAudioSections(
	fetch: typeof globalThis.fetch,
	libraries: Library[],
	filter: { name: 'artist' | 'tag' | 'circle' | 'voice_actor'; value: string }
): Promise<AudioSectionData[]> {
	const sections = await Promise.all(
		libraries
			.filter((library) => library.type === 'audio')
			.map(async (library) => {
				const params = new URLSearchParams({ library_id: library.id, limit: '12' });
				params.set(filter.name, filter.value);

				const res = await fetch(`/api/audio?${params}`);
				if (!res.ok) return { library, works: [], total: 0 };

				const data: AudioListResponse = await res.json();

				return { library, works: data.items ?? [], total: Number(data.total) };
			})
	);

	// A section with no matching works would render nothing; drop it here so
	// the page does not emit empty separators or headings.
	return sections.filter((section) => section.total > 0);
}
