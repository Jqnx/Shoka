<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import { ChevronDown, ChevronUp, X } from '@lucide/svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import type { SortOption, Category, Language } from '$lib/types';

	const CATEGORIES: Category[] = ['Doujinshi', 'Manga', 'Artist CG', 'Game CG', 'Other'];
	const LANGUAGES: Language[] = ['English', 'Japanese'];
	const SORT_OPTIONS: { value: SortOption; label: string }[] = [
		{ value: 'latest', label: 'Latest' },
		{ value: 'title', label: 'Title (A–Z)' },
		{ value: 'release', label: 'Release Date' }
	];

	let {
		artistQuery = $bindable(''),
		sortBy = $bindable<SortOption>('latest'),
		selectedTags,
		selectedCategories,
		selectedLanguages,
		allTags,
		hasActiveFilters,
		clearFilters
	}: {
		artistQuery?: string;
		sortBy?: SortOption;
		selectedTags: SvelteSet<string>;
		selectedCategories: SvelteSet<Category>;
		selectedLanguages: SvelteSet<Language>;
		allTags: string[];
		hasActiveFilters: boolean;
		clearFilters: () => void;
	} = $props();

	let tagsExpanded = $state(true);
	let categoriesExpanded = $state(true);
	let languagesExpanded = $state(true);
	let sortExpanded = $state(true);

	function toggleTag(tag: string) {
		if (selectedTags.has(tag)) selectedTags.delete(tag);
		else selectedTags.add(tag);
	}

	function toggleCategory(cat: Category) {
		if (selectedCategories.has(cat)) selectedCategories.delete(cat);
		else selectedCategories.add(cat);
	}

	function toggleLanguage(lang: Language) {
		if (selectedLanguages.has(lang)) selectedLanguages.delete(lang);
		else selectedLanguages.add(lang);
	}
</script>

<Sidebar.Root collapsible="offcanvas" class="top-14 h-[calc(100svh_-_3.5rem)]">
	{#if hasActiveFilters}
		<Sidebar.Header class="p-3">
			<Button
				variant="ghost"
				size="sm"
				onclick={clearFilters}
				class="justify-start gap-1.5 text-muted-foreground"
			>
				<X class="size-3.5" />
				Clear filters
			</Button>
		</Sidebar.Header>
	{/if}

	<Sidebar.Content class="px-3 py-2">
		<!-- Tags -->
		<Sidebar.Group class="p-0">
			<button
				onclick={() => (tagsExpanded = !tagsExpanded)}
				class="flex w-full items-center justify-between py-1.5 text-xs font-semibold tracking-wider text-muted-foreground uppercase transition-colors hover:text-foreground"
			>
				Tags
				{#if tagsExpanded}<ChevronUp class="size-3.5" />{:else}<ChevronDown class="size-3.5" />{/if}
			</button>
			{#if tagsExpanded}
				<Sidebar.GroupContent class="mt-1.5 flex flex-wrap gap-1.5">
					{#each allTags as tag (tag)}
						{@const active = selectedTags.has(tag)}
						<button onclick={() => toggleTag(tag)}>
							<Badge
								variant={active ? 'default' : 'outline'}
								class="cursor-pointer transition-all {active ? '' : 'hover:bg-muted'}"
							>
								{tag}
							</Badge>
						</button>
					{/each}
				</Sidebar.GroupContent>
			{/if}
		</Sidebar.Group>

		<Separator class="my-2" />

		<!-- Artist -->
		<Sidebar.Group class="p-0">
			<span class="py-1.5 text-xs font-semibold tracking-wider text-muted-foreground uppercase">
				Artist
			</span>
			<Sidebar.GroupContent class="mt-1.5">
				<Sidebar.Input
					type="text"
					placeholder="Filter by artist..."
					bind:value={artistQuery}
				/>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Separator class="my-2" />

		<!-- Category -->
		<Sidebar.Group class="p-0">
			<button
				onclick={() => (categoriesExpanded = !categoriesExpanded)}
				class="flex w-full items-center justify-between py-1.5 text-xs font-semibold tracking-wider text-muted-foreground uppercase transition-colors hover:text-foreground"
			>
				Category
				{#if categoriesExpanded}<ChevronUp class="size-3.5" />{:else}<ChevronDown class="size-3.5" />{/if}
			</button>
			{#if categoriesExpanded}
				<Sidebar.GroupContent class="mt-1 flex flex-col gap-0.5">
					{#each CATEGORIES as cat (cat)}
						{@const active = selectedCategories.has(cat)}
						<Button
							variant={active ? 'default' : 'ghost'}
							size="sm"
							onclick={() => toggleCategory(cat)}
							class="justify-start"
						>
							{cat}
						</Button>
					{/each}
				</Sidebar.GroupContent>
			{/if}
		</Sidebar.Group>

		<Separator class="my-2" />

		<!-- Language -->
		<Sidebar.Group class="p-0">
			<button
				onclick={() => (languagesExpanded = !languagesExpanded)}
				class="flex w-full items-center justify-between py-1.5 text-xs font-semibold tracking-wider text-muted-foreground uppercase transition-colors hover:text-foreground"
			>
				Language
				{#if languagesExpanded}<ChevronUp class="size-3.5" />{:else}<ChevronDown class="size-3.5" />{/if}
			</button>
			{#if languagesExpanded}
				<Sidebar.GroupContent class="mt-1 flex gap-1.5">
					{#each LANGUAGES as lang (lang)}
						{@const active = selectedLanguages.has(lang)}
						<Button
							variant={active ? 'default' : 'outline'}
							size="sm"
							onclick={() => toggleLanguage(lang)}
							class="flex-1"
						>
							{lang}
						</Button>
					{/each}
				</Sidebar.GroupContent>
			{/if}
		</Sidebar.Group>

		<Separator class="my-2" />

		<!-- Sort -->
		<Sidebar.Group class="p-0">
			<button
				onclick={() => (sortExpanded = !sortExpanded)}
				class="flex w-full items-center justify-between py-1.5 text-xs font-semibold tracking-wider text-muted-foreground uppercase transition-colors hover:text-foreground"
			>
				Sort by
				{#if sortExpanded}<ChevronUp class="size-3.5" />{:else}<ChevronDown class="size-3.5" />{/if}
			</button>
			{#if sortExpanded}
				<Sidebar.GroupContent class="mt-1 flex flex-col gap-0.5">
					{#each SORT_OPTIONS as opt (opt.value)}
						{@const active = sortBy === opt.value}
						<Button
							variant={active ? 'default' : 'ghost'}
							size="sm"
							onclick={() => (sortBy = opt.value)}
							class="justify-start"
						>
							{opt.label}
						</Button>
					{/each}
				</Sidebar.GroupContent>
			{/if}
		</Sidebar.Group>
	</Sidebar.Content>
</Sidebar.Root>
