<script lang="ts">
	import * as Pagination from '$lib/components/ui/pagination';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import AudioCard from '$lib/components/AudioCard.svelte';
	import AudioSidebar from '$lib/components/AudioSidebar.svelte';
	import FiltersTrigger from '$lib/components/FiltersTrigger.svelte';
	import { Input } from '$lib/components/ui/input';
	import { AudioLines, Search } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	let { data } = $props();

	// Bound so a filter change (which re-runs the load) does not snap the
	// panel shut between clicks.
	let filtersOpen = $state(false);
	// Writable derived: typing overrides it locally, while a same-route
	// navigation (sidebar library link, back/forward) that changes q resets it
	// to what the backend actually applied.
	let search = $derived(data.filters.q);

	const activeFilters = $derived.by(() => {
		const values: string[] = [];
		for (const tag of data.filters.tag) values.push(`tag: ${tag}`);
		for (const artist of data.filters.artist) values.push(`artist: ${artist}`);
		for (const circle of data.filters.circle) values.push(`circle: ${circle}`);
		for (const actor of data.filters.voice_actor) values.push(`voice actor: ${actor}`);
		return values;
	});

	function submitSearch(event: SubmitEvent) {
		event.preventDefault();
		const params = new URLSearchParams(page.url.searchParams);
		const q = search.trim();
		if (q) params.set('q', q);
		else params.delete('q');
		params.delete('page');
		goto(resolve(`${page.url.pathname}?${params}`), { noScroll: true });
	}
</script>

<svelte:head>
	<title>Audio | Shoka</title>
</svelte:head>

<Sidebar.Provider bind:open={filtersOpen} style="--sidebar-width: 13rem;">
	<Sidebar.Inset class="bg-transparent">
		<div class="mx-auto w-full max-w-screen-2xl px-4 py-8 sm:px-6">
			<div class="mb-6 flex items-center gap-2">
				<AudioLines class="size-5 text-muted-foreground" />
				<h1 class="text-2xl font-bold">Audio</h1>
				{#if data.library}
					<span class="text-sm text-muted-foreground">{data.library.name}</span>
				{/if}
				<span class="text-sm text-muted-foreground">({data.total})</span>
				<div class="ml-auto">
					<FiltersTrigger />
				</div>
			</div>

			<form class="mb-4 flex max-w-md items-center gap-2" onsubmit={submitSearch}>
				<div class="relative flex-1">
					<Search class="absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
					<Input bind:value={search} placeholder="Search audio…" class="pl-8" />
				</div>
			</form>

			{#if activeFilters.length > 0}
				<p class="mb-4 text-sm text-muted-foreground">{activeFilters.join(' · ')}</p>
			{/if}

			{#if data.audio.length === 0}
				<p class="text-muted-foreground">No audio found.</p>
			{:else}
				<div
					class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8"
				>
					{#each data.audio as work (work.id)}
						<AudioCard audio={work} />
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

	<AudioSidebar
		favorites={data.filters.favorites === 'true'}
		rating={Number(data.filters.rating) || 0}
	/>
</Sidebar.Provider>
