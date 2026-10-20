<script lang="ts">
	import type { ReaderSettings } from '$lib/types';
	import ImageContextMenu from '$lib/components/ImageContextMenu.svelte';
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';
	import { ImagePlus } from '@lucide/svelte';
	import { pageImageStyle, pageUrl } from './reader-utils';

	let {
		archiveId,
		pageCount,
		currentPage,
		settings,
		coverPage,
		canSetCover,
		onSetCover,
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
		onSetCover?: (index: number) => void;
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
<div class="flex h-dvh w-full items-center justify-center gap-1" onclick={handleClick}>
	{#each displayIndexes as index (index)}
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
					draggable="false"
				/>
			{/snippet}
		</ImageContextMenu>
	{/each}
</div>

<div class="hidden">
	{#each preloadIndexes as index (index)}
		<img src={pageUrl(archiveId, index)} alt="" />
	{/each}
</div>
