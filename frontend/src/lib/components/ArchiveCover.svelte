<script lang="ts">
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { coverSrc } from '$lib/types';
	import { coverGaveUp, coverPollCount, watchCover } from '$lib/components/coverWatch.svelte';

	let {
		archive,
		alt = '',
		watch = true,
		class: className = ''
	}: {
		/** The fields the cover URL and its placeholder depend on, and no more. */
		archive: { id: string; cover_page: number; thumbs_ready?: boolean };
		/** Empty by default: a cover that never arrives must not paint any text. */
		alt?: string;
		/**
		 * Whether this cover should keep the page's lists re-fetching while it is
		 * missing. Views whose data isn't a page load (the search dialog fetches
		 * its own results) would only cost the poller without ever being
		 * refreshed by it.
		 */
		watch?: boolean;
		class?: string;
	} = $props();

	let coverEl = $state<HTMLImageElement | undefined>(undefined);
	let coverLoaded = $state(false);
	let coverFailed = $state(false);

	// The cover endpoint answers 404 until the thumbnail job has written the
	// page, and a failed <img> paints the browser's broken-image glyph plus its
	// alt text - which is what an un-generated cover used to look like. The
	// image therefore stays invisible until it has actually loaded, and what
	// stands in for it is a pulsing block while the backend says the archive has
	// pages without thumbnails, and nothing but its box's muted background once
	// it doesn't - either because the cover arrived, or because we stopped
	// waiting on it (see coverWatch). An archive that doesn't report readiness
	// at all gets no placeholder: we have no reason to think a cover is coming.
	const coverPlaceholder = $derived(
		!coverLoaded && archive.thumbs_ready === false && !coverGaveUp(archive.id)
	);

	// The result of one request, so it is re-armed whenever that request
	// changes: a different cover_page is a different URL, and a refreshed list
	// entry that no longer reports the archive as incomplete means the cover was
	// written since. An incomplete archive also re-arms on every completed poll:
	// its list entry says nothing new, but the poll re-fetched it, and that is
	// the only signal that the page has since been written - page 0 is generated
	// before the marker that clears thumbs_ready. Dropping the failed element is
	// what makes the re-armed cover ask again; without this it would sit on an
	// empty box (or keep pulsing) until a full page load, however stale the
	// result it kept.
	let coverRequest = '';
	$effect(() => {
		const waiting = archive.thumbs_ready === false;
		const request = `${coverSrc(archive)}:${archive.thumbs_ready}:${waiting ? coverPollCount() : ''}`;
		if (request === coverRequest) return;

		coverRequest = request;
		coverLoaded = false;
		coverFailed = false;
	});

	// A cover served from the cache can finish loading before hydration has
	// attached the load listener, in which case that event never reaches us and
	// the image would sit on its placeholder, invisible. The element's own state
	// settles that case.
	$effect(() => {
		if (!coverEl?.complete) return;

		coverLoaded = coverEl.naturalWidth > 0;
		coverFailed = coverEl.naturalWidth === 0;
	});

	// Watched from here rather than from each grid, so the views that render a
	// single cover (the archive page) get a live one too.
	$effect(() => {
		if (archive.thumbs_ready !== false || !watch) return;

		return watchCover(archive.id);
	});
</script>

{#if coverPlaceholder}
	<!-- The default Skeleton fill is bg-muted, i.e. the same colour as the box
	     it sits on, which would hide the pulse entirely. -->
	<Skeleton class="absolute inset-0 size-full bg-muted-foreground/10" />
{/if}

<!-- Anchors and images are natively draggable: a mousedown plus the slightest
     wobble starts a drag, which fires pointercancel and kills an in-progress
     hold. -->
{#if !coverFailed}
	<img
		bind:this={coverEl}
		src={coverSrc(archive)}
		{alt}
		draggable="false"
		onload={() => (coverLoaded = true)}
		onerror={() => (coverFailed = true)}
		class="transition-opacity duration-300 {coverLoaded ? 'opacity-100' : 'opacity-0'} {className}"
	/>
{/if}
