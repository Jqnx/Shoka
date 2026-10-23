<script lang="ts">
	import type { ReaderSettings } from '$lib/types';
	import ImageContextMenu from '$lib/components/ImageContextMenu.svelte';
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';
	import { Heart, ImagePlus } from '@lucide/svelte';
	import { pageImageStyle, pageUrl } from './reader-utils';
	import type { SvelteSet } from 'svelte/reactivity';

	let {
		archiveId,
		pageCount,
		currentPage,
		settings,
		coverPage,
		canSetCover,
		favoritePages,
		onSetCover,
		onToggleFavorite,
		onLeftZone,
		onRightZone,
		onToggleToolbar
	}: {
		archiveId: string;
		pageCount: number;
		currentPage: number;
		settings: ReaderSettings;
		coverPage?: number;
		canSetCover?: boolean;
		favoritePages: SvelteSet<number>;
		onSetCover?: (index: number) => void;
		onToggleFavorite: (index: number) => void;
		onLeftZone: () => void;
		onRightZone: () => void;
		onToggleToolbar: () => void;
	} = $props();

	const isDouble = $derived(settings.page_layout === 'double');
	const secondIndex = $derived(isDouble && currentPage + 1 < pageCount ? currentPage + 1 : null);
	// Two-page spreads read in the same direction as everything else - in rtl
	// the higher (later) page index sits on the visual left.
	const displayIndexes = $derived.by(() => {
		if (secondIndex === null) return [currentPage];
		return settings.reading_direction === 'ltr'
			? [currentPage, secondIndex]
			: [secondIndex, currentPage];
	});

	// Warm the cache for the page(s) a "next" tap would land on, so the turn
	// feels instant - the page endpoint is served with a long immutable
	// Cache-Control, so this is a real win, not just a wasted request.
	const preloadIndexes = $derived.by(() => {
		const step = isDouble ? 2 : 1;
		const indexes: number[] = [];
		for (let i = 0; i < step; i++) {
			const idx = currentPage + step + i;
			if (idx < pageCount) indexes.push(idx);
		}
		return indexes;
	});

	let scroller = $state<HTMLDivElement | null>(null);

	// A tall/wide page can be scrolled before turning it; the next page must
	// start at its top-left rather than inheriting that offset.
	$effect(() => {
		void currentPage;
		scroller?.scrollTo({ top: 0, left: 0 });
	});

	function handleClick(e: MouseEvent) {
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const ratio = (e.clientX - rect.left) / rect.width;
		if (ratio < 0.35) onLeftZone();
		else if (ratio > 0.65) onRightZone();
		else onToggleToolbar();
	}
</script>

<!--
	This surface is a tap/click navigation zone, not a semantic control - all
	keyboard navigation is wired up at the document level in Reader.svelte,
	so there's no keyboard-equivalent to provide here.
-->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div bind:this={scroller} class="h-dvh w-full overflow-auto" onclick={handleClick}>
	<div class="flex h-full min-w-full items-start gap-1">
		{#each displayIndexes as index (index)}
			{#snippet pageActions()}
				<ContextMenu.Item onSelect={() => onToggleFavorite(index)}>
					<Heart
						class={favoritePages.has(index) ? 'size-4 fill-red-500 stroke-red-500' : 'size-4'}
					/>
					{favoritePages.has(index) ? 'Remove from favorites' : 'Favorite this page'}
				</ContextMenu.Item>
				{#if onSetCover !== undefined && canSetCover}
					<ContextMenu.Item
						disabled={index === coverPage}
						aria-label={index === coverPage ? 'Current page is already the cover' : undefined}
						onSelect={() => onSetCover?.(index)}
					>
						<ImagePlus class="size-4" />
						Set as cover
					</ContextMenu.Item>
				{/if}
			{/snippet}
			<div
				class={settings.fit_mode === 'original'
					? 'relative m-auto flex shrink-0 items-center justify-center'
					: settings.fit_mode === 'height'
						? 'relative flex h-full min-w-0 flex-1 items-center justify-center'
						: 'relative my-auto flex min-w-0 flex-1 items-center justify-center'}
			>
				<ImageContextMenu
					src={pageUrl(archiveId, index)}
					downloadName="{archiveId}-page-{index + 1}"
					actions={pageActions}
				>
					{#snippet trigger({ props })}
						<img
							{...props}
							src={pageUrl(archiveId, index)}
							alt="Page {index + 1}"
							style={pageImageStyle(settings.fit_mode)}
							class="object-contain select-none"
							draggable="false"
						/>
					{/snippet}
				</ImageContextMenu>
			</div>
		{/each}
	</div>
</div>

<div class="hidden">
	{#each preloadIndexes as index (index)}
		<img src={pageUrl(archiveId, index)} alt="" />
	{/each}
</div>
