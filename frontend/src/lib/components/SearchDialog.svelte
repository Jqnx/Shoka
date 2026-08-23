<script lang="ts">
	import * as Command from '$lib/components/ui/command';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { Search, Loader2 } from '@lucide/svelte';
	import type { Archive, ArchiveListResponse, Library } from '$lib/types';

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

	let open = $state(false);
	let query = $state('');
	let results = $state<Archive[]>([]);
	let loading = $state(false);
	let debounceTimer: ReturnType<typeof setTimeout> | undefined;
	// Guards against an in-flight request for a since-superseded query landing
	// after a newer one and clobbering fresher results.
	let requestId = 0;

	const trimmed = $derived(query.trim());
	const queryTooShort = $derived(trimmed.length > 0 && trimmed.length < MIN_QUERY_LENGTH);

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
			results = [];
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
			const res = await fetch(`/api/archives?${params}`);
			if (id !== requestId) return;
			const data: ArchiveListResponse = res.ok
				? await res.json()
				: { items: [], total: 0, page: 1, limit: PREVIEW_LIMIT };
			results = data.items ?? [];
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
			results = [];
			loading = false;
			return;
		}
		debounceTimer = setTimeout(() => fetchResults(value), DEBOUNCE_MS);
	}

	function openDialog() {
		query = '';
		results = [];
		loading = false;
		open = true;
	}

	function handleShortcut(e: KeyboardEvent) {
		if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
			e.preventDefault();
			openDialog();
		}
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
	class="top-[15%] sm:max-w-2xl [&_[data-slot=input-group]]:h-11!"
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
		{:else if loading && results.length === 0}
			<div class="flex items-center justify-center gap-2 py-10 text-sm text-muted-foreground">
				<Loader2 class="size-4 animate-spin" />
				Searching...
			</div>
		{:else}
			<Command.Group>
				<!-- First item so it's selected by default: pressing Enter right
				     after typing goes straight to the full results page. -->
				<Command.LinkItem
					value="see-all"
					href={resultsHref}
					onSelect={() => (open = false)}
					class="gap-2.5 py-2.5 text-base"
				>
					<Search class="size-4 shrink-0" />
					<span class="truncate">See all results for "{trimmed}"</span>
				</Command.LinkItem>

				{#each results as archive (archive.id)}
					<Command.LinkItem
						value={archive.id}
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

			{#if results.length === 0 && !loading}
				<p class="py-10 text-center text-sm text-muted-foreground">No archives found.</p>
			{/if}
		{/if}
	</Command.List>
</Command.Dialog>
