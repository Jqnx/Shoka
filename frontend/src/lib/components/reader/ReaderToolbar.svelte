<script lang="ts">
	import type { ReaderSettings } from '$lib/types';
	import { resolve } from '$app/paths';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { ChevronLeft, ImagePlus, Loader2, Settings } from '@lucide/svelte';
	import ReaderSettingsPanel from './ReaderSettingsPanel.svelte';

	let {
		title,
		backHref,
		currentPage,
		pageCount,
		visible,
		settings,
		settingsOpen = $bindable(false),
		onSettingsChange,
		coverPage,
		settingCover,
		coverError,
		onSetCover
	}: {
		title: string;
		backHref: string;
		currentPage: number;
		pageCount: number;
		visible: boolean;
		settings: ReaderSettings;
		settingsOpen?: boolean;
		onSettingsChange: (patch: Partial<ReaderSettings>) => void;
		/** 0-based page currently used as the archive's cover. */
		coverPage: number;
		settingCover: boolean;
		coverError: string | null;
		onSetCover: () => void;
	} = $props();

	const isCover = $derived(currentPage === coverPage);
</script>

<div
	class="fixed inset-x-0 top-0 z-50 flex items-center gap-3 border-b border-white/10 bg-black/70 px-3 py-2 text-white backdrop-blur transition-transform duration-150 {visible
		? 'translate-y-0'
		: '-translate-y-full'}"
>
	<!--
		The outer div is what actually flexes/shrinks to fill the space between
		the page counter and the toolbar edge; the link itself only sizes to
		its content (chevron + title) so its hover state doesn't visually
		balloon into a wide bar - it just stays capped to that content's width
		via max-w-full, truncating the title if it doesn't fit.
	-->
	<div class="min-w-0 flex-1">
		<a
			href={resolve(backHref as `/${string}`)}
			class="inline-flex max-w-full items-center gap-1.5 rounded-md px-2 py-1 text-white/90 transition-colors hover:bg-white/10 hover:text-white"
		>
			<ChevronLeft class="size-5 shrink-0" />
			<span class="min-w-0 truncate text-sm font-medium">{title}</span>
		</a>
	</div>

	<p class="shrink-0 text-xs text-white/70 tabular-nums">
		{currentPage + 1} / {pageCount}
	</p>

	{#if coverError}
		<p class="shrink-0 text-xs text-red-400">{coverError}</p>
	{/if}

	<Button
		variant="ghost"
		size="icon"
		class="shrink-0 text-white hover:bg-white/10 hover:text-white disabled:opacity-40"
		disabled={isCover || settingCover}
		aria-label={isCover ? 'Current page is already the cover' : 'Set current page as cover'}
		title={isCover ? 'Current page is already the cover' : 'Set current page as cover'}
		onclick={onSetCover}
	>
		{#if settingCover}
			<Loader2 class="size-5 animate-spin" />
		{:else}
			<ImagePlus class={isCover ? 'size-5 fill-current' : 'size-5'} />
		{/if}
	</Button>

	<Popover.Root bind:open={settingsOpen}>
		<Popover.Trigger>
			{#snippet child({ props })}
				<Button
					{...props}
					variant="ghost"
					size="icon"
					class="shrink-0 text-white hover:bg-white/10 hover:text-white"
				>
					<Settings class="size-5" />
				</Button>
			{/snippet}
		</Popover.Trigger>
		<Popover.Content align="end" class="w-64">
			<ReaderSettingsPanel {settings} onChange={onSettingsChange} />
		</Popover.Content>
	</Popover.Root>
</div>
