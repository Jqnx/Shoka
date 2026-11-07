import type { FitMode, ReaderBackground } from '$lib/types';

// Inline styles rather than Tailwind classes since the value is a runtime
// per-user setting, not a fixed set of utility classes pickable at build
// time.
export function pageImageStyle(fitMode: FitMode): string {
	switch (fitMode) {
		case 'width':
			return 'width: 100%; height: auto; max-width: 100%;';
		case 'height':
			return 'height: 100%; width: auto; max-height: 100%;';
		case 'original':
			return 'width: auto; height: auto; max-width: none; max-height: none;';
	}
}

export function backgroundClass(background: ReaderBackground): string {
	switch (background) {
		case 'white':
			return 'bg-white';
		case 'gray':
			return 'bg-neutral-500';
		case 'black':
			return 'bg-black';
	}
}

export function pageUrl(archiveId: string, index: number): string {
	return `/api/archives/${archiveId}/pages/${index}`;
}
