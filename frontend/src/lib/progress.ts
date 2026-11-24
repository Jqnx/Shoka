import type { Progress } from '$lib/types';

/**
 * Width (0-100) of an archive's reading-progress bar.
 *
 * `current_page` is a 0-based index, and a finished archive normally sits on
 * its last page (`page_count - 1`, see `MarkArchivesRead` and `UpdateProgress`),
 * so the pages seen is `current_page + 1`. Dividing by `page_count` alone left
 * the bar short of the end even on the last page.
 *
 * `completed` wins over the page: the progress endpoint lets a client mark an
 * archive finished at any position, and such an archive is done, not 12 pages in.
 */
export function progressPercent(progress: Progress, pageCount: number): number {
	if (pageCount <= 0) return 0;
	if (progress.completed) return 100;
	return Math.min(100, Math.round(((progress.current_page + 1) / pageCount) * 100));
}
