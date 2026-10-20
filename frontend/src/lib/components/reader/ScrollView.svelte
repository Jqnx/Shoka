<script lang="ts">
	import type { ReaderSettings } from '$lib/types';
	import ImageContextMenu from '$lib/components/ImageContextMenu.svelte';
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';
	import { ImagePlus } from '@lucide/svelte';
	import { pageImageStyle, pageUrl } from './reader-utils';

	let {
		archiveId,
		pageCount,
		settings,
		initialPage,
		coverPage,
		canSetCover,
		onSetCover,
		onCurrentPageChange,
		onToggleToolbar
	}: {
		archiveId: string;
		pageCount: number;
		settings: ReaderSettings;
		initialPage: number;
		coverPage?: number;
		canSetCover?: boolean;
		onSetCover?: (index: number) => void;
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
			{#snippet setCoverActions()}
				<ContextMenu.Item
					disabled={index === coverPage}
					aria-label={index === coverPage ? 'Current page is already the cover' : undefined}
					onSelect={() => onSetCover?.(index)}
				>
					<ImagePlus class="size-4" />
					Set as cover
				</ContextMenu.Item>
			{/snippet}
			<ImageContextMenu
				src={pageUrl(archiveId, index)}
				downloadName="{archiveId}-page-{index + 1}"
				actions={onSetCover !== undefined && canSetCover ? setCoverActions : undefined}
			>
				{#snippet trigger({ props })}
					<img
						{...props}
						src={pageUrl(archiveId, index)}
						alt="Page {index + 1}"
						style={pageImageStyle(settings.fit_mode)}
						class="object-contain select-none"
						loading={Math.abs(index - initialPage) <= 2 ? 'eager' : 'lazy'}
						draggable="false"
					/>
				{/snippet}
			</ImageContextMenu>
		</div>
	{/each}
</div>
