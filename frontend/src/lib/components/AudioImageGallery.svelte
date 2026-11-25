<script lang="ts">
	import { SvelteSet } from 'svelte/reactivity';
	import { invalidateAll } from '$app/navigation';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import ImageContextMenu from '$lib/components/ImageContextMenu.svelte';
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';
	import { audioImagePreviewSrc, audioImageThumbSrc } from '$lib/audio';
	import type { AudioImage } from '$lib/types';
	import { cn } from '$lib/utils.js';
	import { Heart, ImageOff, ImagePlus, Loader2 } from '@lucide/svelte';

	let {
		audioId,
		updatedAt,
		images,
		coverIndex: initialCoverIndex,
		canSetCover = false
	}: {
		audioId: string;
		updatedAt: string;
		images: AudioImage[];
		coverIndex: number;
		canSetCover?: boolean;
	} = $props();

	const readyIndexes = new SvelteSet<number>();
	let previewIndex = $state<number | null>(null);
	let genError = $state<string | null>(null);

	// The server prop is the source of truth, but a successful set-cover can
	// override it locally until the invalidated load lands, which re-derives
	// the value (and a navigation to another work starts from that work's).
	let coverIndex = $derived(initialCoverIndex);
	let coverError = $state<string | null>(null);
	let settingCover = $state<number | null>(null);

	// Per-user favorite state for this work's gallery images. Loaded
	// client-side because the context menu only needs it once interactive.
	const favoriteImages = new SvelteSet<number>();
	// Toggles made while the initial list is in flight, re-applied over the
	// fetched snapshot so a late response cannot undo them.
	const pendingFavoriteToggles = new Map<number, boolean>();
	// One request per work/image at a time; keyed by work because the
	// component instance is reused across /audio/[id] navigations.
	const favoritePending = new SvelteSet<string>();
	let favoriteLoad = 0;

	function favoriteKey(index: number) {
		return `${audioId}:${index}`;
	}

	async function toggleFavorite(index: number) {
		const id = audioId;
		const key = favoriteKey(index);
		if (favoritePending.has(key)) return;

		favoritePending.add(key);

		const wasFavorited = favoriteImages.has(index);
		pendingFavoriteToggles.set(index, !wasFavorited);

		if (wasFavorited) favoriteImages.delete(index);
		else favoriteImages.add(index);

		try {
			let ok = false;
			try {
				const res = await fetch(`/api/audio/${id}/images/${index}/favorite`, {
					method: wasFavorited ? 'DELETE' : 'PUT'
				});
				ok = res.ok;
			} catch {
				// Network failure is treated like a rejected request.
			}

			if (ok) return;

			// A newer toggle (or a work change) owns this index now.
			if (audioId !== id) return;
			if (pendingFavoriteToggles.get(index) !== !wasFavorited) return;
			pendingFavoriteToggles.delete(index);

			if (wasFavorited) favoriteImages.add(index);
			else favoriteImages.delete(index);
		} finally {
			favoritePending.delete(key);
		}
	}

	async function setCover(index: number) {
		if (!canSetCover || settingCover !== null || index === coverIndex) return;

		// This component survives /audio/[id] navigations, so later writes
		// must be tied to the work the request was made for.
		const id = audioId;
		coverError = null;
		settingCover = index;

		let ok = false;
		try {
			const res = await fetch(`/api/audio/${id}/cover`, {
				method: 'PUT',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ index })
			});
			ok = res.ok;
		} catch {
			// Network failure leaves ok false.
		}

		// A response for the previous work must not mark this one's cover.
		if (audioId !== id) return;

		settingCover = null;

		if (!ok) {
			coverError = 'Failed to set cover.';
			return;
		}

		coverIndex = index;

		// The top-of-page cover URL carries cover_updated_at, so the new render
		// only appears once the load is re-run. A rejected load is a page-load
		// error, not a cover failure - the change itself already succeeded.
		try {
			await invalidateAll();
		} catch {
			// The local coverIndex still reflects the change; a later reload
			// fetches the fresh cover_updated_at.
		}
	}

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

	// ImageContextMenu wraps its trigger props with a guard that swallows the
	// synthetic click after a touch long-press. This component supplies its own
	// button, so the guarded handler has to run first: a suppressed click must
	// not open the preview.
	function handleThumbnailClick(event: MouseEvent, props: Record<string, unknown>, index: number) {
		const guarded = props.onclick as ((event: MouseEvent) => void) | undefined;
		guarded?.(event);
		if (!event.defaultPrevented) openPreview(index);
	}

	// Same for pointerdown: the menu tracks it to tell a touch long-press from
	// a mouse right-click, and it must still see the event before prefetching.
	function handleThumbnailPointerDown(
		event: PointerEvent,
		props: Record<string, unknown>,
		index: number
	) {
		const guarded = props.onpointerdown as ((event: PointerEvent) => void) | undefined;
		guarded?.(event);
		prefetchPreview(index);
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
		let cancelled = false;
		let source: EventSource | undefined;

		readyIndexes.clear();
		genError = null;
		previewIndex = null;
		previewStatus = null;
		prefetched.clear();
		settingCover = null;
		coverError = null;

		function connect() {
			if (cancelled) return;

			const stream = new EventSource(`/api/audio/${id}/thumbnails/events`);
			source = stream;

			stream.addEventListener('ready', (e) => {
				const payload = JSON.parse((e as MessageEvent).data) as { index: number };
				readyIndexes.add(payload.index);
			});

			stream.addEventListener('done', (e) => {
				const payload = JSON.parse((e as MessageEvent).data) as {
					done: boolean;
					error?: string;
				};
				if (payload.error) {
					// A failure to start generation is the root cause; the stream's
					// "no job running" message must not replace it.
					genError ??= payload.error;
				} else {
					// The broadcaster drops events beyond its buffer, so a
					// successful done means every thumbnail is on disk.
					genError = null;
					for (const image of images) readyIndexes.add(image.index);
				}
				stream.close();
			});

			stream.onerror = () => {
				// A stream that gives up for good (auth failure, deleted work)
				// never sends done, so otherwise the grid stays on skeletons.
				if (stream.readyState === EventSource.CLOSED && !cancelled) {
					genError ??= 'Lost connection while generating thumbnails.';
					stream.close();
				}
			};
		}

		// The trigger has to land before the stream opens: the stream ends at
		// once when no job is pending, so a GET that overtakes the POST would
		// report a work that is about to generate as "not running". A failed
		// trigger would otherwise leave the grid as silent skeletons, since the
		// stream only carries progress.
		fetch(`/api/audio/${id}/thumbnails`, { method: 'POST' })
			.then((res) => {
				if (!cancelled && !res.ok) genError = 'Failed to start thumbnail generation.';
			})
			.catch(() => {
				if (!cancelled) genError = 'Failed to start thumbnail generation.';
			})
			.finally(connect);

		return () => {
			cancelled = true;
			clearTimeout(prefetchTimer);
			source?.close();
		};
	});

	// The gallery's favorite marks are per user and not part of the work
	// payload, so they are loaded once per work and kept client-side.
	$effect(() => {
		const id = audioId;
		const version = ++favoriteLoad;

		// This instance survives /audio/[id] navigations, so work B must never
		// show work A's marks while its own list is in flight.
		favoriteImages.clear();
		pendingFavoriteToggles.clear();

		fetch(`/api/audio/${id}/favorite-images`)
			.then((res) => (res.ok ? res.json() : { indices: [] }))
			.then((data: { indices?: number[] }) => {
				if (version !== favoriteLoad) return;

				const indices = new Set(data.indices ?? []);
				for (const [index, favorited] of pendingFavoriteToggles) {
					if (favorited) indices.add(index);
					else indices.delete(index);
				}

				favoriteImages.clear();
				for (const index of indices) favoriteImages.add(index);
			})
			.catch(() => {
				// Best effort: the menu simply starts with nothing marked.
			});
	});
</script>

<div>
	{#if genError}
		<p class="mb-3 text-sm text-destructive">Failed to generate thumbnails: {genError}</p>
	{/if}
	{#if settingCover !== null}
		<p class="mb-3 flex items-center gap-1.5 text-sm text-muted-foreground" role="status">
			<Loader2 class="size-4 animate-spin" />
			Setting cover…
		</p>
	{/if}
	{#if coverError}
		<p class="mb-3 flex items-center gap-1.5 text-sm text-destructive" role="alert">
			<ImageOff class="size-4" />
			{coverError}
		</p>
	{/if}

	<div
		class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8"
	>
		{#each images as image (image.index)}
			{#snippet imageActions()}
				<ContextMenu.Item
					disabled={favoritePending.has(favoriteKey(image.index))}
					onSelect={() => toggleFavorite(image.index)}
				>
					<Heart
						class={favoriteImages.has(image.index)
							? 'size-4 fill-red-500 stroke-red-500'
							: 'size-4'}
					/>
					{favoriteImages.has(image.index) ? 'Remove from favorites' : 'Favorite this image'}
				</ContextMenu.Item>
				{#if canSetCover}
					<ContextMenu.Item
						disabled={image.index === coverIndex || settingCover !== null}
						aria-label={image.index === coverIndex
							? 'Current image is already the cover'
							: undefined}
						onSelect={() => setCover(image.index)}
					>
						<ImagePlus class="size-4" />
						Set as cover
					</ContextMenu.Item>
				{/if}
			{/snippet}
			<ImageContextMenu
				src={audioImageThumbSrc(audioId, image.index, updatedAt)}
				downloadName="{audioId}-image-{image.index + 1}.webp"
				actions={imageActions}
			>
				{#snippet trigger({ props })}
					<button
						{...props}
						tabindex="0"
						type="button"
						class="relative aspect-square overflow-hidden rounded-lg border border-border bg-muted"
						onclick={(event) => handleThumbnailClick(event, props, image.index)}
						onpointerdown={(event) => handleThumbnailPointerDown(event, props, image.index)}
						onmouseenter={() => schedulePrefetch(image.index)}
						onmouseleave={cancelPrefetch}
						onfocus={() => schedulePrefetch(image.index)}
						onblur={cancelPrefetch}
						aria-label="View {image.name}{image.index === coverIndex
							? ' (current cover)'
							: ''}{favoriteImages.has(image.index) ? ' (favorited)' : ''}"
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
						{#if image.index === coverIndex}
							<span
								class="pointer-events-none absolute top-1.5 left-1.5 rounded-md bg-background/90 px-1.5 py-0.5 text-xs font-semibold text-foreground shadow-sm"
							>
								Cover
							</span>
						{/if}
						{#if favoriteImages.has(image.index)}
							<span
								class="pointer-events-none absolute top-1.5 right-1.5 rounded-full bg-background/90 p-1 text-red-500 shadow-sm"
								aria-hidden="true"
							>
								<Heart class="size-3.5 fill-red-500 stroke-red-500" />
							</span>
						{/if}
					</button>
				{/snippet}
			</ImageContextMenu>
		{/each}
	</div>
</div>

<Dialog.Root
	open={previewIndex !== null}
	onOpenChange={(isOpen) => {
		if (!isOpen) previewIndex = null;
	}}
>
	<Dialog.Content class="pt-9 sm:max-w-[min(90vw,72rem)]">
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
					onload={() => {
						if (previewImage) previewStatus = { index: previewImage.index, state: 'ready' };
					}}
					onerror={() => {
						if (previewImage) previewStatus = { index: previewImage.index, state: 'error' };
					}}
				/>
			</div>
		{/if}
	</Dialog.Content>
</Dialog.Root>
