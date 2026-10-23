<script lang="ts">
	import * as Pagination from '$lib/components/ui/pagination';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';
	import ImageContextMenu from '$lib/components/ImageContextMenu.svelte';
	import { Button } from '$lib/components/ui/button';
	import { BookOpen, HeartOff, Images } from '@lucide/svelte';
	import { afterNavigate, goto } from '$app/navigation';
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	afterNavigate((nav) => {
		if (nav.type === 'goto') window.scrollTo({ top: 0, behavior: 'smooth' });
	});

	let { data } = $props();
</script>

<svelte:head>
	<title>Favorite Pages | Shoka</title>
</svelte:head>

<Sidebar.Provider open={false}>
	<Sidebar.Inset class="bg-transparent">
		<div class="flex flex-wrap items-center justify-between gap-2 px-4 py-3 sm:px-6">
			<div>
				<h1 class="text-base font-semibold">Favorite Pages</h1>
				<p class="text-sm text-muted-foreground">
					{data.total}
					{data.total === 1 ? 'page' : 'pages'} found
				</p>
			</div>
		</div>

		<div class="px-4 py-4 sm:px-6">
			{#if data.items.length === 0}
				<div class="flex flex-col items-center justify-center py-24 text-center">
					<Images class="mb-3 size-10 text-muted-foreground/40" />
					<p class="text-base font-medium text-foreground">No favorite pages yet</p>
					<p class="mt-1 text-sm text-muted-foreground">
						Favorite a page from the reader to see it here.
					</p>
				</div>
			{:else}
				<div
					class="columns-2 gap-2 sm:columns-3 md:columns-4 lg:columns-5 xl:columns-6 [&>div]:mb-2"
				>
					{#each data.items as item, i (item.archive_id + ':' + item.page_index)}
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
										href={resolve(`/favorites/pages/viewer?at=${(data.page - 1) * data.limit + i}`)}
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
									<form
										method="POST"
										action="?/removeFavorite"
										use:enhance={() => {
											return async ({ update }) => {
												await update();
											};
										}}
									>
										<input type="hidden" name="archive_id" value={item.archive_id} />
										<input type="hidden" name="page_index" value={item.page_index} />
										<Button
											type="submit"
											variant="secondary"
											size="icon-sm"
											aria-label="Remove {item.title} page {item.page_index + 1} from favorites"
											title="Remove from favorites"
										>
											<HeartOff class="size-4" />
										</Button>
									</form>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}

			{#if data.total > data.limit}
				<div class="mt-8">
					<Pagination.Root
						count={data.total}
						perPage={data.limit}
						page={data.page}
						onPageChange={(p) => {
							const params = new URLSearchParams(page.url.searchParams);
							params.set('page', String(p));
							goto(resolve(`${page.url.pathname}?${params}`), { noScroll: true });
						}}
					>
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
											<Pagination.Link page={p} isActive={data.page === p.value} />
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
		</div>
	</Sidebar.Inset>
</Sidebar.Provider>
