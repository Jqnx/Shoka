import { invalidateAll } from '$app/navigation';
import { SvelteSet } from 'svelte/reactivity';

// Covers are generated in the background: a freshly scanned library opens with
// whatever the thumbnail job has written so far and the cover endpoint 404s for
// the rest. Nothing would ever ask again - the list a card renders is only as
// fresh as the last navigation - so an incomplete cover registers here and this
// re-runs the current page's loads (one refetch for the whole grid, rather than
// a stream per card) until either the cover arrives or we stop expecting it.

/** How often an outstanding cover gets the page's loads re-run for it. */
const POLL_MS = 10_000;

/**
 * How many consecutive polls may find the watched set unchanged before we stop
 * asking. Any change at all - an archive finishing, a grid being left behind -
 * resets the count, so a busy library keeps being polled for as long as it is
 * making progress, while a cover whose job keeps failing stops being asked for
 * after roughly a minute instead of pulsing a placeholder forever.
 */
const MAX_STALLED_POLLS = 6;

/**
 * Archive ids whose cover is still outstanding, and how many views are waiting
 * on each: the same archive can be rendered by a grid card and by a carousel on
 * the home page at once, and the first of the two to go away must not cancel
 * the other's wait.
 */
const watching = new Map<string, number>();

/** Ids we have stopped waiting on - their placeholders stop pulsing. */
const gaveUp = new SvelteSet<string>();

/**
 * How many polls have completed. A view re-reads this so that a cover written
 * since the last poll is requested again, even though the list it came from
 * still says the archive is incomplete (page 0 is written before the marker
 * that clears thumbs_ready).
 */
const polls = $state({ done: 0 });

let stalledPolls = 0;
let lastWaitingSize = 0;
let scheduled = false;

/** True once we have stopped asking for this archive's cover. */
export function coverGaveUp(archiveID: string): boolean {
	return gaveUp.has(archiveID);
}

/** How many times the watched lists have been re-fetched; changes every poll. */
export function coverPollCount(): number {
	return polls.done;
}

/**
 * watchCover reports an archive as still waiting for its cover, and returns the
 * function that stops reporting it (its cover arriving, or its view going away).
 */
export function watchCover(archiveID: string): () => void {
	// Registering is new evidence: whatever we concluded earlier about this
	// archive - a give-up recorded for an earlier page, most likely - no longer
	// applies, and the budget starts over rather than carrying over from
	// whichever ids happened to be watched before.
	gaveUp.delete(archiveID);
	watching.set(archiveID, (watching.get(archiveID) ?? 0) + 1);
	stalledPolls = 0;
	lastWaitingSize = watching.size;
	schedule();

	return () => {
		const remaining = (watching.get(archiveID) ?? 1) - 1;

		// Another view is still showing this archive's cover.
		if (remaining > 0) watching.set(archiveID, remaining);
		else watching.delete(archiveID);
	};
}

function schedule() {
	if (scheduled) return;

	scheduled = true;
	setTimeout(poll, POLL_MS);
}

/**
 * giveUp retires everything still being watched: the placeholders stop pulsing
 * and the poller stops asking. Used both when the budget runs out and when the
 * page's own loads fail - either way, re-running them is not getting us anywhere.
 */
function giveUp() {
	for (const archiveID of watching.keys()) gaveUp.add(archiveID);

	watching.clear();
	lastWaitingSize = 0;
}

async function poll() {
	scheduled = false;

	// Nothing outstanding any more: every cover arrived, or the views that were
	// waiting on them went away.
	if (watching.size === 0) {
		stalledPolls = 0;

		return;
	}

	if (watching.size === lastWaitingSize) stalledPolls++;
	else stalledPolls = 0;

	lastWaitingSize = watching.size;

	if (stalledPolls >= MAX_STALLED_POLLS) {
		giveUp();

		return;
	}

	try {
		await invalidateAll();
	} catch {
		// A load that throws is a problem for the page to report, not something
		// to retry from here.
		giveUp();

		return;
	}

	polls.done++;
	schedule();
}
