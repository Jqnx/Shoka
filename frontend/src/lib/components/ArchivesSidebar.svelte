<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { Combobox, MultiCombobox } from '$lib/components/ui/combobox';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import { X } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Artist, ArchiveLanguage, Character, Parody, Tag } from '$lib/types';

	const ANY = 'any';

	type Filters = {
		// '' means unset - the backend treats has_phash as tri-state (omit to
		// include both hashed and unhashed archives).
		has_phash: string;
		category: string;
		language: string;
		tag: string[];
		character: string[];
		parody: string[];
		artist: string[];
	};

	let {
		filters,
		artists,
		tags,
		characters,
		parodies,
		categories,
		languages
	}: {
		filters: Filters;
		artists: Artist[];
		tags: Tag[];
		characters: Character[];
		parodies: Parody[];
		categories: string[];
		languages: ArchiveLanguage[];
	} = $props();

	const hasActiveFilters = $derived(
		filters.has_phash !== '' ||
			filters.category !== '' ||
			filters.language !== '' ||
			filters.artist.length > 0 ||
			filters.tag.length > 0 ||
			filters.character.length > 0 ||
			filters.parody.length > 0
	);

	const phashOptions = [
		{ value: ANY, label: 'Any' },
		{ value: 'true', label: 'Hashed' },
		{ value: 'false', label: 'Not hashed' }
	];
	const phashLabel = $derived(
		phashOptions.find((o) => o.value === (filters.has_phash || ANY))?.label ?? 'Any'
	);

	const categoryOptions = $derived([
		{ value: ANY, label: 'Any category' },
		...categories.map((cat) => ({ value: cat, label: cat }))
	]);
	// The filter matches the raw stored language code exactly (see
	// ArchiveFilter.Language in the backend), so the option value must be the
	// code - only the label is the human-readable name.
	const languageOptions = $derived([
		{ value: ANY, label: 'Any language' },
		...languages.map((lang) => ({ value: lang.code, label: lang.name }))
	]);

	const artistOptions = $derived(artists.map((artist) => ({ value: artist.name, label: artist.name })));
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
		goto(`${page.url.pathname}?${params}`, { noScroll: true, keepFocus: true });
	}

	function setMultiParam(name: string, values: string[]) {
		const params = new URLSearchParams(page.url.searchParams);
		params.delete(name);
		for (const value of values) {
			if (value) params.append(name, value);
		}
		params.delete('page');
		goto(`${page.url.pathname}?${params}`, { noScroll: true, keepFocus: true });
	}

	function clearFilters() {
		const params = new URLSearchParams(page.url.searchParams);
		for (const key of [
			'has_phash',
			'category',
			'language',
			'tag',
			'character',
			'parody',
			'artist',
			'page'
		]) {
			params.delete(key);
		}
		goto(`${page.url.pathname}?${params}`, { noScroll: true });
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
				<MultiCombobox
					options={artistOptions}
					value={filters.artist}
					onValueChange={(v) => setMultiParam('artist', v)}
					placeholder="Add artist..."
					searchPlaceholder="Search artists..."
					emptyText="No artists found."
				/>
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

		<Separator />

		<!-- Perceptual hash -->
		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Perceptual hash</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<Select.Root
					type="single"
					value={filters.has_phash || ANY}
					onValueChange={(v) => setParam('has_phash', v)}
				>
					<Select.Trigger class="w-full">{phashLabel}</Select.Trigger>
					<Select.Content>
						{#each phashOptions as opt (opt.value)}
							<Select.Item value={opt.value} label={opt.label}>{opt.label}</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>
</Sidebar.Root>
