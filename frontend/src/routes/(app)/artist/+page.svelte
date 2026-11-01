<script lang="ts">
	import * as Pagination from '$lib/components/ui/pagination';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import ArtistCard from '$lib/components/ArtistCard.svelte';
	import ArtistsSidebar from '$lib/components/ArtistsSidebar.svelte';
	import FiltersTrigger from '$lib/components/FiltersTrigger.svelte';
	import { Paintbrush } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	let { data, form } = $props();

	const hasFilters = $derived(data.favorites || data.rating > 0);
	// Bound so a filter change (which re-runs the load) does not snap the
	// panel shut between star clicks.
	let filtersOpen = $state(false);
</script>

<svelte:head>
	<title>Artists | Shoka</title>
</svelte:head>

<Sidebar.Provider bind:open={filtersOpen} style="--sidebar-width: 13rem;">
	<Sidebar.Inset class="bg-transparent">
		<div class="mx-auto w-full max-w-screen-2xl px-4 py-8 sm:px-6">
			<div class="mb-6 flex items-center gap-2">
				<Paintbrush class="size-5 text-muted-foreground" />
				<h1 class="text-2xl font-bold">Artists</h1>
				<span class="text-sm text-muted-foreground">({data.total})</span>
				<div class="ml-auto">
					<FiltersTrigger />
				</div>
			</div>

			{#if form && 'error' in form && form.error}
				<p class="mb-4 text-sm text-destructive">{form.error}</p>
			{/if}

			{#if data.artists.length === 0}
				<p class="text-muted-foreground">
					{hasFilters ? 'No artists match these filters.' : 'No artists found.'}
				</p>
			{:else}
				<div
					class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8"
				>
					{#each data.artists as artist (artist.id)}
						<ArtistCard {artist} />
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

	<ArtistsSidebar favorites={data.favorites} rating={data.rating} />
</Sidebar.Provider>
