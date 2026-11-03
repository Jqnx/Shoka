<script lang="ts">
	import { SvelteSet } from 'svelte/reactivity';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { audioImagePreviewSrc, audioImageThumbSrc } from '$lib/audio';
	import type { AudioImage } from '$lib/types';
	import { cn } from '$lib/utils.js';
	import { ImageOff, Loader2 } from '@lucide/svelte';

	let { audioId, updatedAt, images }: { audioId: string; updatedAt: string; images: AudioImage[] } =
		$props();

	const readyIndexes = new SvelteSet<number>();
	let previewIndex = $state<number | null>(null);
	let genError = $state<string | null>(null);

	const previewImage = $derived(
		previewIndex === null ? null : (images.find((image) => image.index === previewIndex) ?? null)
	);

	// Tracks the image the dialog's render actually finished loading, keyed by
	// index so a stale result from a previously opened image cannot mark the
	// current one ready.
	let previewStatus = $state<{ index: number; state: 'ready' | 'error' } | null>(null);
	const previewState = $derived(
		previewStatus !== null && previewStatus.index === previewImage?.index
			? previewStatus.state
			: 'loading'
	);

	// A preview takes a render on the server the first time and a network round
	// trip after that. Starting it while the pointer rests on a thumbnail makes
	// most opens immediate; the dwell keeps a sweep across the grid from
	// fetching every preview.
	const prefetched = new Set<string>();
	let prefetchTimer: ReturnType<typeof setTimeout> | undefined;

	function prefetchPreview(index: number) {
		const key = `${audioId}:${index}`;
		if (prefetched.has(key)) return;

		prefetched.add(key);
		const preload = new Image();
		preload.src = audioImagePreviewSrc(audioId, index, updatedAt);
	}

	function schedulePrefetch(index: number) {
		clearTimeout(prefetchTimer);
		prefetchTimer = setTimeout(() => prefetchPreview(index), 150);
	}

	function cancelPrefetch() {
		clearTimeout(prefetchTimer);
	}

	// Every open starts from the loading state: the browser may still have to
	// fetch or decode the render, and keeping a previous "ready" would render
	// an empty dialog while it does.
	function openPreview(index: number) {
		previewStatus = null;
		previewIndex = index;
	}

	// Triggers generation (idempotent - a no-op if already ready/running) and
	// streams progress via SSE, filling thumbnails in as they are generated.
	// This trigger is ungated, so there is no permission check here.
	$effect(() => {
		const id = audioId;

		readyIndexes.clear();
		genError = null;
		previewIndex = null;
		previewStatus = null;
		prefetched.clear();

		fetch(`/api/audio/${id}/thumbnails`, { method: 'POST' }).catch(() => {});

		const source = new EventSource(`/api/audio/${id}/thumbnails/events`);

		source.addEventListener('ready', (e) => {
			const payload = JSON.parse((e as MessageEvent).data) as { index: number };
			readyIndexes.add(payload.index);
		});

		source.addEventListener('done', (e) => {
			const payload = JSON.parse((e as MessageEvent).data) as { done: boolean; error?: string };
			if (payload.error) {
				genError = payload.error;
			} else {
				// The broadcaster drops events beyond its buffer, so a
				// successful done means every thumbnail is on disk.
				for (const image of images) readyIndexes.add(image.index);
			}
			source.close();
		});

		source.onerror = () => {
			if (source.readyState === EventSource.CLOSED) {
				source.close();
			}
		};

		return () => {
			clearTimeout(prefetchTimer);
			source.close();
		};
	});
</script>

<div>
	{#if genError}
		<p class="mb-3 text-sm text-destructive">Failed to generate thumbnails: {genError}</p>
	{/if}

	<div
		class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8"
	>
		{#each images as image (image.index)}
			<button
				type="button"
				class="aspect-square overflow-hidden rounded-lg border border-border bg-muted"
				onclick={() => openPreview(image.index)}
				onpointerdown={() => prefetchPreview(image.index)}
				onmouseenter={() => schedulePrefetch(image.index)}
				onmouseleave={cancelPrefetch}
				onfocus={() => schedulePrefetch(image.index)}
				onblur={cancelPrefetch}
				aria-label="View {image.name}"
			>
				{#if readyIndexes.has(image.index)}
					<img
						src={audioImageThumbSrc(audioId, image.index, updatedAt)}
						alt={image.name}
						loading="lazy"
						decoding="async"
						class="size-full object-cover transition-transform hover:scale-105"
					/>
				{:else}
					<Skeleton class="size-full" />
				{/if}
			</button>
		{/each}
	</div>
</div>

<Dialog.Root
	open={previewIndex !== null}
	onOpenChange={(isOpen) => {
		if (!isOpen) previewIndex = null;
	}}
>
	<Dialog.Content class="sm:max-w-[min(90vw,72rem)]">
		<Dialog.Header>
			<Dialog.Title class="sr-only">{previewImage?.name ?? 'Image'}</Dialog.Title>
		</Dialog.Header>
		{#if previewImage}
			<div
				class={cn(
					'relative flex w-full justify-center',
					previewState !== 'ready' && 'min-h-[50vh] items-center'
				)}
			>
				{#if previewState === 'loading'}
					<div
						class="absolute inset-0 flex flex-col items-center justify-center gap-3 text-muted-foreground"
						role="status"
					>
						<Loader2 class="size-6 animate-spin" />
						<span class="text-sm">Loading preview…</span>
					</div>
				{:else if previewState === 'error'}
					<div
						class="absolute inset-0 flex flex-col items-center justify-center gap-3 text-destructive"
						role="status"
					>
						<ImageOff class="size-5" />
						<span class="text-sm">Failed to load preview</span>
					</div>
				{/if}
				<img
					src={audioImagePreviewSrc(audioId, previewImage.index, updatedAt)}
					alt={previewImage.name}
					class={cn(
						'max-h-[80vh] w-full object-contain transition-opacity duration-200',
						previewState === 'ready' ? 'opacity-100' : 'h-0 opacity-0'
					)}
					onload={() => (previewStatus = { index: previewImage.index, state: 'ready' })}
					onerror={() => (previewStatus = { index: previewImage.index, state: 'error' })}
				/>
			</div>
		{/if}
	</Dialog.Content>
</Dialog.Root>
