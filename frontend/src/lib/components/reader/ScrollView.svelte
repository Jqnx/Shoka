<script lang="ts">
	import type { ReaderSettings } from '$lib/types';
	import { pageImageStyle, pageUrl } from './reader-utils';

	let {
		archiveId,
		pageCount,
		settings,
		initialPage,
		onCurrentPageChange,
		onToggleToolbar
	}: {
		archiveId: string;
		pageCount: number;
		settings: ReaderSettings;
		initialPage: number;
		onCurrentPageChange: (index: number) => void;
		onToggleToolbar: () => void;
	} = $props();

	const pages = $derived(Array.from({ length: pageCount }, (_, i) => i));

	let container = $state<HTMLDivElement | null>(null);
	let pageEls = $state<(HTMLDivElement | null)[]>([]);

	// Jump straight to wherever paged mode left off - this only needs to run
	// once, right after this view is created (Reader.svelte recreates it from
	// scratch each time view_mode switches to "scroll", so there's no
	// stale-prop risk from reading initialPage just on mount).
	$effect(() => {
		pageEls[initialPage]?.scrollIntoView({ block: 'start' });
	});

	$effect(() => {
		if (!container) return;

		const observer = new IntersectionObserver(
			(entries) => {
				let best: IntersectionObserverEntry | null = null;
				for (const entry of entries) {
					if (entry.isIntersecting && (!best || entry.intersectionRatio > best.intersectionRatio)) {
						best = entry;
					}
				}
				if (best) {
					const index = Number((best.target as HTMLElement).dataset.index);
					if (!Number.isNaN(index)) onCurrentPageChange(index);
				}
			},
			{ root: container, threshold: [0.5] }
		);

		for (const el of pageEls) if (el) observer.observe(el);
		return () => observer.disconnect();
	});
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div bind:this={container} class="h-dvh w-full overflow-y-auto" onclick={onToggleToolbar}>
	{#each pages as index (index)}
		<div
			data-index={index}
			bind:this={pageEls[index]}
			class="flex w-full items-center justify-center"
		>
			<img
				src={pageUrl(archiveId, index)}
				alt="Page {index + 1}"
				style={pageImageStyle(settings.fit_mode)}
				class="select-none object-contain"
				loading={Math.abs(index - initialPage) <= 2 ? 'eager' : 'lazy'}
				draggable="false"
			/>
		</div>
	{/each}
</div>
