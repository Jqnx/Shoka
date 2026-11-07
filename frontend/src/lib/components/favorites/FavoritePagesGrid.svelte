<script lang="ts">
	import { onMount } from 'svelte';
	import * as Pagination from '$lib/components/ui/pagination';
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';
	import ImageContextMenu from '$lib/components/ImageContextMenu.svelte';
	import { Button } from '$lib/components/ui/button';
	import { BookOpen, Heart, Images } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import type { FavoritePageItem, FavoritePageListResponse } from '$lib/types';

	const DEFAULT_LIMIT = 40;

	let items = $state<FavoritePageItem[]>([]);
	let total = $state(0);
	let currentPage = $state(1);
	let loading = $state(true);

	onMount(() => {
		void load(1);
	});

	async function load(p: number) {
		try {
			const res = await fetch(`/api/archives/favorites/pages?page=${p}&limit=${DEFAULT_LIMIT}`);
			if (!res.ok) {
				items = [];
				total = 0;
				return;
			}

			const data: FavoritePageListResponse = await res.json();
			items = data.items ?? [];
			total = Number(data.total);
			currentPage = p;
		} catch {
			// Silent degrade, same as the old server load: no error UI.
		} finally {
			loading = false;
		}
	}

	async function unfavorite(item: FavoritePageItem) {
		try {
			const res = await fetch(
				`/api/archives/${item.archive_id}/pages/${item.page_index}/favorite`,
				{ method: 'DELETE' }
			);
			if (!res.ok) return;
		} catch {
			return;
		}

		// Removing the only item on a page past the first would leave that page
		// empty, so step back instead.
		if (items.length === 1 && currentPage > 1) {
			await load(currentPage - 1);
		} else {
			await load(currentPage);
		}
	}

	function onPageChange(p: number) {
		window.scrollTo({ top: 0, behavior: 'smooth' });
		void load(p);
	}
</script>

{#if !loading}
	{#if items.length === 0}
		<div class="flex flex-col items-center justify-center py-24 text-center">
			<Images class="mb-3 size-10 text-muted-foreground/40" />
			<p class="text-base font-medium text-foreground">No favorite pages yet</p>
			<p class="mt-1 text-sm text-muted-foreground">
				Favorite a page from the reader to see it here.
			</p>
		</div>
	{:else}
		<div class="columns-2 gap-2 sm:columns-3 md:columns-4 lg:columns-5 xl:columns-6 [&>div]:mb-2">
			{#each items as item, i (item.archive_id + ':' + item.page_index)}
				{#snippet cellActions()}
					<ContextMenu.Item onSelect={() => goto(resolve(`/a/${item.archive_id}`))}>
						<BookOpen class="size-4" />
						Open in archive
					</ContextMenu.Item>
				{/snippet}
				<div class="group relative break-inside-avoid">
					<ImageContextMenu
						src={`/api/archives/${item.archive_id}/pages/${item.page_index}`}
						downloadName="{item.archive_id}-page-{item.page_index + 1}"
						actions={cellActions}
					>
						{#snippet trigger({ props })}
							<a
								{...props}
								href={resolve(`/favorites/viewer?at=${(currentPage - 1) * DEFAULT_LIMIT + i}`)}
								aria-label="Open {item.title} page {item.page_index + 1} in the viewer"
								class="block"
							>
								<img
									src={`/api/archives/${item.archive_id}/pages/${item.page_index}`}
									alt="{item.title} page {item.page_index + 1}"
									loading="lazy"
									class="w-full select-none"
									draggable="false"
								/>
							</a>
						{/snippet}
					</ImageContextMenu>

					<div
						class="pointer-events-none invisible absolute inset-0 flex flex-col justify-end bg-gradient-to-b from-black/40 via-transparent to-black/60 opacity-0 transition-opacity group-focus-within:visible group-focus-within:opacity-100 group-hover:visible group-hover:opacity-100"
					>
						<p class="truncate px-2 pb-2 text-xs font-medium text-white">
							{item.title}
							<span class="text-white/70">· p. {item.page_index + 1}</span>
						</p>
					</div>

					<!--
						Kept outside the gradient overlay so a touch device, where
						`hover: hover` never matches, can still see and use the
						buttons - `visibility` and `opacity` are independent of the
						hover-only label layer above.
					-->
					<div
						class="pointer-events-none invisible absolute top-1 right-1 z-10 flex gap-1 opacity-0 transition-opacity group-focus-within:visible group-focus-within:opacity-100 group-hover:visible group-hover:opacity-100 [@media(hover:none)]:visible [@media(hover:none)]:opacity-100"
					>
						<div class="pointer-events-auto flex gap-1">
							<Button
								variant="secondary"
								size="icon-sm"
								aria-label="Open {item.title} in the library"
								title="Open in archive"
								onclick={() => goto(resolve(`/a/${item.archive_id}`))}
							>
								<BookOpen class="size-4" />
							</Button>
							<Button
								variant="secondary"
								size="icon-sm"
								aria-label="Remove {item.title} page {item.page_index + 1} from favorites"
								title="Remove from favorites"
								onclick={() => unfavorite(item)}
							>
								<Heart class="size-4 fill-red-500 stroke-red-500" />
							</Button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	{#if total > DEFAULT_LIMIT}
		<div class="mt-8">
			<Pagination.Root count={total} perPage={DEFAULT_LIMIT} page={currentPage} {onPageChange}>
				{#snippet children({ pages })}
					<Pagination.Content>
						<Pagination.Item>
							<Pagination.Previous />
						</Pagination.Item>
						{#each pages as p (p.key)}
							{#if p.type === 'ellipsis'}
								<Pagination.Item>
									<Pagination.Ellipsis />
								</Pagination.Item>
							{:else}
								<Pagination.Item>
									<Pagination.Link page={p} isActive={currentPage === p.value} />
								</Pagination.Item>
							{/if}
						{/each}
						<Pagination.Item>
							<Pagination.Next />
						</Pagination.Item>
					</Pagination.Content>
				{/snippet}
			</Pagination.Root>
		</div>
	{/if}
{/if}
