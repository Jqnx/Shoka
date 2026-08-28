<script lang="ts">
	import * as Command from '$lib/components/ui/command';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { Search, Loader2 } from '@lucide/svelte';
	import type { Library, SearchEntityResult, SearchResponse } from '$lib/types';

	let { libraries }: { libraries: Library[] } = $props();

	// Results are scoped to a single library (the backend has no cross-library
	// search) - use whichever one is being browsed, else the first.
	const libraryId = $derived(page.params.libraryId ?? libraries[0]?.id ?? '');
	const targetPath = $derived(libraryId ? `/${libraryId}` : '/');

	const DEBOUNCE_MS = 300;
	const PREVIEW_LIMIT = 6;
	// The backend's FTS index is trigram-tokenized, so anything shorter than 3
	// characters can't match and is silently ignored as a filter - fetching
	// those would show an arbitrary unfiltered page as if it were results.
	const MIN_QUERY_LENGTH = 3;

	const EMPTY_RESULTS: SearchResponse = {
		archives: { items: [], total: 0 },
		artists: { items: [], total: 0 },
		circles: { items: [], total: 0 },
		tags: { items: [], total: 0 },
		parodies: { items: [], total: 0 },
		characters: { items: [], total: 0 }
	};

	let open = $state(false);
	let query = $state('');
	let results = $state<SearchResponse | null>(null);
	let loading = $state(false);
	let debounceTimer: ReturnType<typeof setTimeout> | undefined;
	// Guards against an in-flight request for a since-superseded query landing
	// after a newer one and clobbering fresher results.
	let requestId = 0;

	const trimmed = $derived(query.trim());
	const queryTooShort = $derived(trimmed.length > 0 && trimmed.length < MIN_QUERY_LENGTH);

	const hasResults = $derived(
		!!results &&
			(results.archives.items.length > 0 ||
				results.artists.items.length > 0 ||
				results.circles.items.length > 0 ||
				results.tags.items.length > 0 ||
				results.parodies.items.length > 0 ||
				results.characters.items.length > 0)
	);

	// The full-results page, carrying over the current page's other params
	// (sort, filters, ...) only when it's the same route search lands on -
	// searching from elsewhere shouldn't drag along filters that route never
	// displayed a UI for.
	const resultsHref = $derived.by(() => {
		const params = new URLSearchParams(
			page.url.pathname === targetPath ? page.url.searchParams : undefined
		);
		params.set('q', trimmed);
		params.delete('page');
		return `${targetPath}?${params}`;
	});

	const fromParam = $derived(encodeURIComponent(page.url.pathname + page.url.search));

	async function fetchResults(value: string) {
		const q = value.trim();
		if (q.length < MIN_QUERY_LENGTH || !libraryId) {
			results = null;
			return;
		}

		const id = ++requestId;
		loading = true;
		try {
			const params = new URLSearchParams({
				library_id: libraryId,
				q,
				limit: String(PREVIEW_LIMIT)
			});
			const res = await fetch(`/api/search?${params}`);
			if (id !== requestId) return;
			const data: SearchResponse = res.ok ? await res.json() : EMPTY_RESULTS;
			results = data;
		} finally {
			if (id === requestId) loading = false;
		}
	}

	function handleInput(value: string) {
		query = value;
		clearTimeout(debounceTimer);
		if (value.trim().length < MIN_QUERY_LENGTH) {
			// Abandon any in-flight request so its results can't land after
			// the user has already cleared/shortened the query.
			requestId++;
			results = null;
			loading = false;
			return;
		}
		debounceTimer = setTimeout(() => fetchResults(value), DEBOUNCE_MS);
	}

	function openDialog() {
		query = '';
		results = null;
		loading = false;
		open = true;
	}

	function handleShortcut(e: KeyboardEvent) {
		if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
			e.preventDefault();
			openDialog();
		}
	}

	function archivesCountLabel(count: number) {
		return `in ${count} ${count === 1 ? 'archive' : 'archives'}`;
	}

	// Artists/tags/parodies/characters are global entities with their own
	// browsing routes; circles have neither a dedicated page nor a backend
	// filter, so they fall back to a full-text query against the archive
	// list (the FTS index already covers circle names).
	function entityHref(
		category: 'artist' | 'tag' | 'parody' | 'character',
		item: SearchEntityResult
	) {
		if (category === 'artist') return resolve(`/artist/${item.id}`);
		return resolve(`/${category}/${encodeURIComponent(item.name)}`);
	}

	function circleHref(item: SearchEntityResult) {
		return `${targetPath}?q=${encodeURIComponent(item.name)}`;
	}
</script>

<svelte:window onkeydown={handleShortcut} />

<button
	type="button"
	onclick={openDialog}
	class="flex h-10 w-full max-w-lg items-center gap-2 rounded-lg border border-input bg-transparent px-3 text-sm text-muted-foreground transition-colors hover:bg-accent/50 dark:bg-input/30"
>
	<Search class="size-4 shrink-0" />
	<span class="truncate">{page.url.searchParams.get('q') || 'Search'}</span>
	<kbd
		class="ms-auto hidden h-5 shrink-0 items-center rounded border border-border px-1.5 pt-[2px] font-mono text-[10px] leading-none text-muted-foreground sm:inline-flex"
	>
		Ctrl+K
	</kbd>
</button>

<Command.Dialog
	bind:open
	shouldFilter={false}
	title="Search archives"
	description="Search this library by title, artist, tag, parody, character, or circle."
	class="top-[15%] sm:max-w-4xl [&_[data-slot=input-group]]:h-11!"
>
	<Command.Input
		placeholder="Search archives..."
		value={query}
		class="text-base"
		oninput={(e) => handleInput(e.currentTarget.value)}
	/>
	<Command.List class="max-h-[60vh]">
		{#if !trimmed}
			<p class="py-10 text-center text-sm text-muted-foreground">
				Start typing to search this library.
			</p>
		{:else if queryTooShort}
			<p class="py-10 text-center text-sm text-muted-foreground">
				Keep typing - at least {MIN_QUERY_LENGTH} characters.
			</p>
		{:else if loading && !results}
			<div class="flex items-center justify-center gap-2 py-10 text-sm text-muted-foreground">
				<Loader2 class="size-4 animate-spin" />
				Searching...
			</div>
		{:else}
			<Command.Group>
				<!-- First item so it's selected by default: pressing Enter right
				     after typing goes straight to the full results page. Kept
				     outside the grid below so it stays visible even when there
				     are zero archive matches. -->
				<Command.LinkItem
					value="see-all"
					href={resultsHref}
					onSelect={() => (open = false)}
					class="gap-2.5 py-2.5 text-base"
				>
					<Search class="size-4 shrink-0" />
					<span class="truncate">See all results for "{trimmed}"</span>
				</Command.LinkItem>
			</Command.Group>

			<div class="grid grid-cols-1 gap-x-4 sm:grid-cols-2 lg:grid-cols-3">
				{#if results?.archives.items.length}
					<Command.Group heading="Archives">
						{#each results.archives.items as archive (archive.id)}
							<Command.LinkItem
								value={`archive-${archive.id}`}
								href={`${resolve(`/a/${archive.id}`)}?from=${fromParam}`}
								onSelect={() => (open = false)}
								class="gap-3 py-2"
							>
								<img
									src="/api/archives/{archive.id}/cover"
									alt=""
									class="h-16 w-11 shrink-0 rounded object-cover"
								/>
								<div class="min-w-0">
									<p class="truncate text-base font-medium">{archive.title}</p>
									{#if archive.artists?.length}
										<p class="truncate text-sm text-muted-foreground">
											{archive.artists.join(', ')}
										</p>
									{/if}
								</div>
							</Command.LinkItem>
						{/each}
					</Command.Group>
				{/if}

				{#if results?.artists.items.length}
					<Command.Group heading="Artists">
						{#each results.artists.items as artist (artist.id)}
							<Command.LinkItem
								value={`artist-${artist.id}`}
								href={entityHref('artist', artist)}
								onSelect={() => (open = false)}
							>
								<span class="truncate">{artist.name}</span>
								<span class="ms-auto shrink-0 truncate text-xs text-muted-foreground">
									{archivesCountLabel(artist.count)}
								</span>
							</Command.LinkItem>
						{/each}
					</Command.Group>
				{/if}

				{#if results?.tags.items.length}
					<Command.Group heading="Tags">
						{#each results.tags.items as tag (tag.id)}
							<Command.LinkItem
								value={`tag-${tag.id}`}
								href={entityHref('tag', tag)}
								onSelect={() => (open = false)}
							>
								<span class="truncate">{tag.name}</span>
								<span class="ms-auto shrink-0 truncate text-xs text-muted-foreground">
									{archivesCountLabel(tag.count)}
								</span>
							</Command.LinkItem>
						{/each}
					</Command.Group>
				{/if}

				{#if results?.parodies.items.length}
					<Command.Group heading="Parodies">
						{#each results.parodies.items as parody (parody.id)}
							<Command.LinkItem
								value={`parody-${parody.id}`}
								href={entityHref('parody', parody)}
								onSelect={() => (open = false)}
							>
								<span class="truncate">{parody.name}</span>
								<span class="ms-auto shrink-0 truncate text-xs text-muted-foreground">
									{archivesCountLabel(parody.count)}
								</span>
							</Command.LinkItem>
						{/each}
					</Command.Group>
				{/if}

				{#if results?.characters.items.length}
					<Command.Group heading="Characters">
						{#each results.characters.items as character (character.id)}
							<Command.LinkItem
								value={`character-${character.id}`}
								href={entityHref('character', character)}
								onSelect={() => (open = false)}
							>
								<span class="truncate">{character.name}</span>
								<span class="ms-auto shrink-0 truncate text-xs text-muted-foreground">
									{archivesCountLabel(character.count)}
								</span>
							</Command.LinkItem>
						{/each}
					</Command.Group>
				{/if}

				{#if results?.circles.items.length}
					<Command.Group heading="Circles">
						{#each results.circles.items as circle (circle.id)}
							<Command.LinkItem
								value={`circle-${circle.id}`}
								href={circleHref(circle)}
								onSelect={() => (open = false)}
							>
								<span class="truncate">{circle.name}</span>
								<span class="ms-auto shrink-0 truncate text-xs text-muted-foreground">
									{archivesCountLabel(circle.count)}
								</span>
							</Command.LinkItem>
						{/each}
					</Command.Group>
				{/if}
			</div>

			{#if !hasResults && !loading}
				<p class="py-10 text-center text-sm text-muted-foreground">No results found.</p>
			{/if}
		{/if}
	</Command.List>
</Command.Dialog>
