<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { Combobox, MultiCombobox } from '$lib/components/ui/combobox';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import { X } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import type { Category, Character, Language, Parody, Tag } from '$lib/types';

	const CATEGORIES: Category[] = ['Doujinshi', 'Manga', 'Artist CG', 'Game CG', 'Other'];
	const LANGUAGES: Language[] = ['English', 'Japanese'];
	const ANY = 'any';

	type Filters = {
		category: string;
		language: string;
		tag: string[];
		character: string[];
		parody: string[];
		artist: string;
	};

	let {
		filters,
		tags,
		characters,
		parodies
	}: {
		filters: Filters;
		tags: Tag[];
		characters: Character[];
		parodies: Parody[];
	} = $props();

	// svelte-ignore state_referenced_locally
	let artistInput = $state(filters.artist);

	const hasActiveFilters = $derived(
		filters.category !== '' ||
			filters.language !== '' ||
			filters.artist !== '' ||
			filters.tag.length > 0 ||
			filters.character.length > 0 ||
			filters.parody.length > 0
	);

	const categoryOptions = [
		{ value: ANY, label: 'Any category' },
		...CATEGORIES.map((cat) => ({ value: cat, label: cat }))
	];
	const languageOptions = [
		{ value: ANY, label: 'Any language' },
		...LANGUAGES.map((lang) => ({ value: lang, label: lang }))
	];

	const tagOptions = $derived(tags.map((tag) => ({ value: tag.name, label: tag.name })));
	const characterOptions = $derived(
		characters.map((character) => ({ value: character.name, label: character.name }))
	);
	const parodyOptions = $derived(
		parodies.map((parody) => ({ value: parody.name, label: parody.name }))
	);

	function setParam(name: string, value: string) {
		const params = new URLSearchParams(page.url.searchParams);
		if (value && value !== ANY) params.set(name, value);
		else params.delete(name);
		params.delete('page');
		goto(`${resolve('/a')}?${params}`, { noScroll: true, keepFocus: true });
	}

	function setMultiParam(name: string, values: string[]) {
		const params = new URLSearchParams(page.url.searchParams);
		params.delete(name);
		for (const value of values) {
			if (value) params.append(name, value);
		}
		params.delete('page');
		goto(`${resolve('/a')}?${params}`, { noScroll: true, keepFocus: true });
	}

	function submitArtist(e: SubmitEvent) {
		e.preventDefault();
		setParam('artist', artistInput.trim());
	}

	function clearFilters() {
		const params = new URLSearchParams(page.url.searchParams);
		for (const key of ['category', 'language', 'tag', 'character', 'parody', 'artist', 'page']) {
			params.delete(key);
		}
		artistInput = '';
		goto(`${resolve('/a')}?${params}`, { noScroll: true });
	}
</script>

<Sidebar.Root side="right" collapsible="offcanvas" class="top-14 h-[calc(100svh-3.5rem)]">
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

	<Sidebar.Content class="gap-4 px-3 py-2">
		<!-- Artist -->
		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Artist</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<form onsubmit={submitArtist}>
					<Sidebar.Input type="text" placeholder="Exact artist name..." bind:value={artistInput} />
				</form>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Separator />

		<!-- Category -->
		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Category</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<Combobox
					options={categoryOptions}
					value={filters.category || ANY}
					onValueChange={(v) => setParam('category', v)}
					placeholder="Any category"
					searchPlaceholder="Search category..."
				/>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Separator />

		<!-- Language -->
		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Language</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<Combobox
					options={languageOptions}
					value={filters.language || ANY}
					onValueChange={(v) => setParam('language', v)}
					placeholder="Any language"
					searchPlaceholder="Search language..."
				/>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Separator />

		<!-- Tags -->
		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Tags</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<MultiCombobox
					options={tagOptions}
					value={filters.tag}
					onValueChange={(v) => setMultiParam('tag', v)}
					placeholder="Add tag..."
					searchPlaceholder="Search tags..."
					emptyText="No tags found."
				/>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Separator />

		<!-- Characters -->
		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Characters</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<MultiCombobox
					options={characterOptions}
					value={filters.character}
					onValueChange={(v) => setMultiParam('character', v)}
					placeholder="Add character..."
					searchPlaceholder="Search characters..."
					emptyText="No characters found."
				/>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Separator />

		<!-- Parodies -->
		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Parodies</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<MultiCombobox
					options={parodyOptions}
					value={filters.parody}
					onValueChange={(v) => setMultiParam('parody', v)}
					placeholder="Add parody..."
					searchPlaceholder="Search parodies..."
					emptyText="No parodies found."
				/>
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>
</Sidebar.Root>
