<script lang="ts">
	import * as Pagination from '$lib/components/ui/pagination';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import ArchiveCard from '$lib/components/ArchiveCard.svelte';
	import FavoritePagesGrid from '$lib/components/favorites/FavoritePagesGrid.svelte';
	import { Search } from '@lucide/svelte';
	import { goto, afterNavigate } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	afterNavigate((nav) => {
		if (nav.type === 'goto') window.scrollTo({ top: 0, behavior: 'smooth' });
	});

	let { data } = $props();

	// Client-side only: the archives tab paginates through the URL (`?page=`),
	// but query-only navigation does not remount this component, so the tab
	// survives.
	let tab = $state('archives');
</script>

<svelte:head>
	<title>Favorites | Shoka</title>
</svelte:head>

<Sidebar.Provider open={false}>
	<Sidebar.Inset class="bg-transparent">
		<div class="flex flex-wrap items-center justify-between gap-2 px-4 py-3 sm:px-6">
			<h1 class="text-base font-semibold">Favorites</h1>
		</div>

		<div class="px-4 py-4 sm:px-6">
			<Tabs.Root bind:value={tab} class="gap-6">
				<Tabs.List class="self-center">
					<Tabs.Trigger value="archives">Archives</Tabs.Trigger>
					<Tabs.Trigger value="pages">Pages</Tabs.Trigger>
				</Tabs.List>

				<Tabs.Content value="archives" class="text-sm">
					{#if data.archives.length === 0}
						<div class="flex flex-col items-center justify-center py-24 text-center">
							<Search class="mb-3 size-10 text-muted-foreground/40" />
							<p class="text-base font-medium text-foreground">No favorites yet</p>
						</div>
					{:else}
						<div
							class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8"
						>
							{#each data.archives as archive (archive.id)}
								<ArchiveCard {archive} />
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
				</Tabs.Content>

				<Tabs.Content value="pages" class="text-sm">
					<!-- bits-ui keeps every tab panel mounted and only hides the
					     inactive ones, so the grid is gated explicitly to stay
					     lazy: no fetch on page load, a fresh one per activation. -->
					{#if tab === 'pages'}
						<FavoritePagesGrid />
					{/if}
				</Tabs.Content>
			</Tabs.Root>
		</div>
	</Sidebar.Inset>
</Sidebar.Provider>
