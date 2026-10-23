<script lang="ts">
	import { SvelteSet } from 'svelte/reactivity';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import ImageContextMenu from '$lib/components/ImageContextMenu.svelte';
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';
	import { Heart, ImageOff, ImagePlus } from '@lucide/svelte';
	import { can } from '$lib/auth/permissions';

	let {
		archiveId,
		pageCount,
		thumbsReady,
		coverPage: initialCoverPage
	}: {
		archiveId: string;
		pageCount: number;
		thumbsReady: boolean;
		coverPage: number;
	} = $props();

	const readyIndexes = new SvelteSet<number>();
	// Favorite state for this archive's pages, loaded client-side because the
	// context menu only needs it once the gallery is interactive.
	const favoritePages = new SvelteSet<number>();
	// Toggles made while the initial list is in flight. They are re-applied
	// over the fetched snapshot so a late response cannot undo them.
	const pendingFavoriteToggles = new Map<number, boolean>();
	let favoriteLoad = 0;
	let genError = $state<string | null>(null);

	// The server prop is the source of truth, but a successful set-cover can
	// override it locally until the server data changes (reset-cover action, or
	// navigating to another archive), which re-derives the value.
	let coverPage = $derived(initialCoverPage);
	let coverError = $state<string | null>(null);

	async function setCover(index: number) {
		coverError = null;
		try {
			const res = await fetch(`/api/archives/${archiveId}/cover`, {
				method: 'PUT',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ page: index })
			});
			if (!res.ok) {
				coverError = 'Failed to set cover.';
				return;
			}
			coverPage = index;
		} catch {
			coverError = 'Failed to set cover.';
		}
	}

	$effect(() => {
		const archive = archiveId;
		const version = ++favoriteLoad;

		// This component instance survives /a/[id] param changes, so clear
		// synchronously: archive B must never show archive A's marks while its
		// own list is in flight, nor keep them if the fetch fails.
		favoritePages.clear();
		pendingFavoriteToggles.clear();

		fetch(`/api/archives/${archive}/favorite-pages`)
			.then((res) => (res.ok ? res.json() : { indices: [] }))
			.then((data: { indices?: number[] }) => {
				if (version !== favoriteLoad) return;

				const indices = new Set(data.indices ?? []);
				for (const [index, favorited] of pendingFavoriteToggles) {
					if (favorited) indices.add(index);
					else indices.delete(index);
				}

				favoritePages.clear();
				for (const index of indices) favoritePages.add(index);
			})
			.catch(() => {
				// Best-effort: the menu simply starts with nothing marked.
			});
	});

	async function toggleFavorite(index: number) {
		const archive = archiveId;
		const wasFavorited = favoritePages.has(index);
		pendingFavoriteToggles.set(index, !wasFavorited);

		if (wasFavorited) favoritePages.delete(index);
		else favoritePages.add(index);

		try {
			const res = await fetch(`/api/archives/${archive}/pages/${index}/favorite`, {
				method: wasFavorited ? 'DELETE' : 'PUT'
			});
			if (res.ok) return;
		} catch {
			// Fall through to the revert below.
		}

		// A newer toggle (or an archive change) owns this index now.
		if (archive !== archiveId) return;
		if (pendingFavoriteToggles.get(index) !== !wasFavorited) return;
		pendingFavoriteToggles.delete(index);
		revertFavorite(index, wasFavorited);
	}

	function revertFavorite(index: number, wasFavorited: boolean) {
		if (wasFavorited) favoritePages.add(index);
		else favoritePages.delete(index);
	}

	// Triggers generation (idempotent - a no-op if already ready/running) and
	// streams progress via SSE, filling in thumbnails as they're generated
	// instead of waiting for the whole archive to finish.
	$effect(() => {
		readyIndexes.clear();
		genError = null;

		if (thumbsReady) {
			for (let i = 0; i < pageCount; i++) readyIndexes.add(i);
			return;
		}

		// The stream is readable by anyone, so a user without generation rights
		// still fills in whatever an admin has already kicked off.
		if (can(page.data.user, { archive: ['thumbnail'] })) {
			fetch(`/api/archives/${archiveId}/thumbnails`, { method: 'POST' }).catch(() => {});
		}

		const source = new EventSource(`/api/archives/${archiveId}/thumbnails/events`);

		source.addEventListener('ready', (e) => {
			const payload = JSON.parse((e as MessageEvent).data) as { index: number };
			readyIndexes.add(payload.index);
		});

		source.addEventListener('done', (e) => {
			const payload = JSON.parse((e as MessageEvent).data) as { done: boolean; error?: string };
			if (payload.error) genError = payload.error;
			source.close();
		});

		source.onerror = () => {
			if (source.readyState === EventSource.CLOSED) {
				source.close();
			}
		};

		return () => source.close();
	});

	const pages = $derived(Array.from({ length: pageCount }, (_, i) => i));
</script>

<div>
	<h2 class="mb-3 text-sm font-semibold">Pages</h2>

	{#if genError}
		<p class="mb-3 flex items-center gap-1.5 text-sm text-destructive">
			<ImageOff class="size-4" />
			Failed to generate thumbnails: {genError}
		</p>
	{/if}

	{#if coverError}
		<p class="mb-3 flex items-center gap-1.5 text-sm text-destructive">
			<ImageOff class="size-4" />
			{coverError}
		</p>
	{/if}

	<div class="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8">
		{#each pages as index (index)}
			{#snippet pageActions()}
				<ContextMenu.Item onSelect={() => toggleFavorite(index)}>
					<Heart
						class={favoritePages.has(index) ? 'size-4 fill-red-500 stroke-red-500' : 'size-4'}
					/>
					{favoritePages.has(index) ? 'Remove from favorites' : 'Favorite this page'}
				</ContextMenu.Item>
				{#if can(page.data.user, { archive: ['cover'] })}
					<ContextMenu.Item
						disabled={index === coverPage}
						aria-label={index === coverPage ? 'Current page is already the cover' : undefined}
						onSelect={() => setCover(index)}
					>
						<ImagePlus class="size-4" />
						Set as cover
					</ContextMenu.Item>
				{/if}
			{/snippet}
			<ImageContextMenu
				src="/api/archives/{archiveId}/pages/{index}/thumbnail"
				downloadName="{archiveId}-page-{index + 1}.webp"
				actions={pageActions}
			>
				{#snippet trigger({ props })}
					<a
						{...props}
						href={resolve(`/a/${archiveId}/${index + 1}`)}
						tabindex="0"
						class="block aspect-[2/3] overflow-hidden rounded-md border border-border bg-muted transition-colors hover:border-primary"
					>
						{#if readyIndexes.has(index)}
							<img
								src="/api/archives/{archiveId}/pages/{index}/thumbnail"
								alt="Page {index + 1}"
								loading="lazy"
								class="size-full object-cover"
							/>
						{:else}
							<Skeleton class="size-full" />
						{/if}
					</a>
				{/snippet}
			</ImageContextMenu>
		{/each}
	</div>
</div>
