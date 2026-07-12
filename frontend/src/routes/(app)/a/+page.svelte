<script lang="ts">
	import * as Pagination from '$lib/components/ui/pagination';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as Select from '$lib/components/ui/select';
	import ArchiveCard from '$lib/components/ArchiveCard.svelte';
	import ArchivesSidebar from '$lib/components/ArchivesSidebar.svelte';
	import FiltersTrigger from '$lib/components/FiltersTrigger.svelte';
	import { Search } from '@lucide/svelte';
	import { goto, afterNavigate } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	afterNavigate((nav) => {
		if (nav.type === 'goto') window.scrollTo({ top: 0, behavior: 'smooth' });
	});

	let { data } = $props();

	const SORT_OPTIONS = [
		{ value: 'created_at_desc', label: 'Latest' },
		{ value: 'created_at_asc', label: 'Oldest' },
		{ value: 'title_asc', label: 'Title (A–Z)' },
		{ value: 'title_desc', label: 'Title (Z–A)' },
		{ value: 'release_date_desc', label: 'Release date (newest)' },
		{ value: 'release_date_asc', label: 'Release date (oldest)' },
		{ value: 'page_count_desc', label: 'Page count (most)' },
		{ value: 'page_count_asc', label: 'Page count (least)' }
	];

	const sortValue = $derived(data.filters.sort || 'created_at_desc');
	const sortLabel = $derived(
		SORT_OPTIONS.find((opt) => opt.value === sortValue)?.label ?? 'Latest'
	);

	function updateParam(name: string, value: string) {
		const params = new URLSearchParams(page.url.searchParams);
		params.set(name, value);
		params.delete('page');
		goto(`${resolve('/a')}?${params}`, { noScroll: true, keepFocus: true });
	}
</script>

<svelte:head>
	<title>Archives | Shoka</title>
</svelte:head>

<Sidebar.Provider open={false}>
	<Sidebar.Inset class="bg-transparent">
		<div class="flex flex-wrap items-center justify-between gap-2 px-4 py-3 sm:px-6">
			<p class="text-sm text-muted-foreground">
				{data.total}
				{data.total === 1 ? 'archive' : 'archives'} found
			</p>

			<div class="flex items-center gap-2">
				<Select.Root type="single" value={sortValue} onValueChange={(v) => updateParam('sort', v)}>
					<Select.Trigger class="h-8 w-[190px]">
						{sortLabel}
					</Select.Trigger>
					<Select.Content>
						<Select.Group>
							<Select.Label>Sort</Select.Label>
							{#each SORT_OPTIONS as opt (opt.value)}
								<Select.Item value={opt.value} label={opt.label}>{opt.label}</Select.Item>
							{/each}
						</Select.Group>
					</Select.Content>
				</Select.Root>

				<FiltersTrigger />
			</div>
		</div>

		<div class="px-4 py-4 sm:px-6">
			{#if data.archives.length === 0}
				<div class="flex flex-col items-center justify-center py-24 text-center">
					<Search class="mb-3 size-10 text-muted-foreground/40" />
					<p class="text-base font-medium text-foreground">No archives found</p>
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
							goto(`${resolve('/a')}?${params}`, { noScroll: true });
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

	<ArchivesSidebar
		filters={data.filters}
		tags={data.tags}
		characters={data.characters}
		parodies={data.parodies}
	/>
</Sidebar.Provider>
