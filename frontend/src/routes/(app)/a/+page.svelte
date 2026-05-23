<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as Pagination from '$lib/components/ui/pagination';
	import { Button } from '$lib/components/ui/button';
	import ArchiveCard from '$lib/components/ArchiveCard.svelte';
	import ArchivesSidebar from '$lib/components/ArchivesSidebar.svelte';
	import { Search } from '@lucide/svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import { goto } from '$app/navigation';
	import type { SortOption, Category, Language } from '$lib/types';

	let { data } = $props();

	const allTags = $derived([...new Set(data.archives.flatMap((a) => a.tags))].sort());

	// --- Filter state ---
	let artistQuery = $state('');
	let selectedTags = new SvelteSet<string>();
	let selectedCategories = new SvelteSet<Category>();
	let selectedLanguages = new SvelteSet<Language>();
	let sortBy = $state<SortOption>('latest');

	function clearFilters() {
		artistQuery = '';
		selectedTags.clear();
		selectedCategories.clear();
		selectedLanguages.clear();
		sortBy = 'latest';
	}

	const hasActiveFilters = $derived(
		artistQuery.trim() !== '' ||
			selectedTags.size > 0 ||
			selectedCategories.size > 0 ||
			selectedLanguages.size > 0 ||
			sortBy !== 'latest'
	);

	const filteredArchives = $derived.by(() => {
		let results = data.archives;

		if (artistQuery.trim()) {
			const q = artistQuery.toLowerCase();
			results = results.filter((a) => a.artists.some((artist) => artist.toLowerCase().includes(q)));
		}

		if (selectedTags.size > 0) {
			results = results.filter((a) => [...selectedTags].every((t) => a.tags.includes(t)));
		}

		if (selectedCategories.size > 0) {
			results = results.filter((a) => a.category !== null && selectedCategories.has(a.category as Category));
		}

		if (selectedLanguages.size > 0) {
			results = results.filter((a) => a.language !== null && selectedLanguages.has(a.language as Language));
		}

		return [...results].sort((a, b) => {
			if (sortBy === 'title') return a.title.localeCompare(b.title);
			if (sortBy === 'release') {
				const da = a.release_date ? new Date(a.release_date).getTime() : 0;
				const db = b.release_date ? new Date(b.release_date).getTime() : 0;
				return db - da;
			}
			return b.created_at.localeCompare(a.created_at);
		});
	});
</script>

<Sidebar.Provider>
	<ArchivesSidebar
		bind:artistQuery
		bind:sortBy
		{selectedTags}
		{selectedCategories}
		{selectedLanguages}
		{allTags}
		{hasActiveFilters}
		{clearFilters}
	/>

	<Sidebar.Inset>
		<div class="sticky top-14 z-10 flex items-center gap-2 border-b border-border bg-background/95 px-4 py-2 backdrop-blur supports-[backdrop-filter]:bg-background/80 sm:px-6">
			<Sidebar.Trigger />
			<p class="text-sm text-muted-foreground">
				{filteredArchives.length}
				{filteredArchives.length === 1 ? 'archive' : 'archives'}
			</p>
		</div>
		<div class="px-4 py-4 sm:px-6">
			{#if filteredArchives.length === 0}
				<div class="flex flex-col items-center justify-center py-24 text-center">
					<Search class="mb-3 size-10 text-muted-foreground/40" />
					<p class="text-base font-medium text-foreground">No archives found</p>
					<p class="mt-1 text-sm text-muted-foreground">Try adjusting your filters</p>
					<Button variant="outline" size="sm" onclick={clearFilters} class="mt-4">
						Clear filters
					</Button>
				</div>
			{:else}
				<div class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8">
					{#each filteredArchives as archive (archive.id)}
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
						onPageChange={(p) => goto(`?page=${p}`)}
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
