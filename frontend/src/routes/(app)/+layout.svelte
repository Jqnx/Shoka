<script lang="ts">
	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Search, Loader2 } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import type { Archive, ArchiveListResponse } from '$lib/types';

	let { data, children } = $props();

	// The archives list route needs a library to scope to - use whichever
	// one is currently being browsed (if any), else default to the first
	// until there's a broader "current library" concept to derive this from.
	const searchAction = $derived(
		page.params.libraryId ? `/${page.params.libraryId}` : data.libraries[0] ? `/${data.libraries[0].id}` : '/'
	);
	const searchLibraryId = $derived(page.params.libraryId ?? data.libraries[0]?.id ?? '');

	let searchValue = $state(page.url.searchParams.get('q') ?? '');

	// Keep the field in sync with the URL when it changes from elsewhere
	// (browser back/forward, clearing filters, switching library) - only
	// reruns when the URL itself changes, not on every keystroke, since
	// typing alone doesn't touch page.url until Enter navigates.
	$effect(() => {
		searchValue = page.url.searchParams.get('q') ?? '';
	});

	const PREVIEW_DEBOUNCE_MS = 300;
	const PREVIEW_LIMIT = 6;
	let debounceTimer: ReturnType<typeof setTimeout> | undefined;

	let previewResults = $state<Archive[]>([]);
	let previewOpen = $state(false);
	let previewLoading = $state(false);
	// Guards against an in-flight request for a since-superseded query
	// landing after a newer one and clobbering fresher results.
	let previewRequestId = 0;

	async function fetchPreview(value: string) {
		const query = value.trim();
		if (!query || !searchLibraryId) {
			previewResults = [];
			previewOpen = false;
			return;
		}

		const requestId = ++previewRequestId;
		previewLoading = true;
		try {
			const params = new URLSearchParams({
				library_id: searchLibraryId,
				q: query,
				limit: String(PREVIEW_LIMIT)
			});
			const res = await fetch(`/api/archives?${params}`);
			if (requestId !== previewRequestId) return;
			const data: ArchiveListResponse = res.ok
				? await res.json()
				: { items: [], total: 0, page: 1, limit: PREVIEW_LIMIT };
			previewResults = data.items ?? [];
			previewOpen = true;
		} finally {
			if (requestId === previewRequestId) previewLoading = false;
		}
	}

	function goToResults(value: string) {
		// Only carry over the current page's other params (sort, filters, ...)
		// when navigating to the same route search lands on - jumping there
		// from a different library/page shouldn't drag along filters that
		// route never displayed a UI for.
		const onTargetPage = page.url.pathname === searchAction;
		const params = new URLSearchParams(onTargetPage ? page.url.searchParams : undefined);
		const query = value.trim();
		if (query) {
			params.set('q', query);
		} else {
			params.delete('q');
		}
		params.delete('page');
		goto(`${searchAction}?${params}`, { keepFocus: true, noScroll: true });
	}

	function handleInput(value: string) {
		searchValue = value;
		clearTimeout(debounceTimer);
		if (!value.trim()) {
			previewResults = [];
			previewOpen = false;
			return;
		}
		debounceTimer = setTimeout(() => fetchPreview(value), PREVIEW_DEBOUNCE_MS);
	}

	function handleFocus() {
		if (searchValue.trim() && previewResults.length) previewOpen = true;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') previewOpen = false;
	}

	function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		clearTimeout(debounceTimer);
		previewOpen = false;
		goToResults(searchValue);
	}

	let formRef: HTMLFormElement | undefined = $state();

	// Closes the dropdown on any click outside the search form. Deliberately
	// not focusout-based: blur fires before a result link's click finishes
	// processing, so closing there would unmount the link mid-click and
	// silently swallow the navigation.
	function handleWindowClick(e: MouseEvent) {
		if (formRef && !formRef.contains(e.target as Node)) previewOpen = false;
	}
</script>

<svelte:window onclick={handleWindowClick} />

<Sidebar.Provider>
	<AppSidebar user={data.user} libraries={data.libraries} />
	<Sidebar.Inset class="bg-muted/40">
		<header
			class="sticky top-0 z-50 flex h-14 shrink-0 items-center gap-2 border-b border-border bg-background/95 px-4 backdrop-blur supports-backdrop-filter:bg-background/80 sm:px-6"
		>
			<Sidebar.Trigger class="-ms-1" />

			<div class="flex flex-1 justify-center">
				<form
					bind:this={formRef}
					method="GET"
					action={searchAction}
					onsubmit={handleSubmit}
					class="relative w-full max-w-lg"
				>
					<Label for="search" class="sr-only">Search</Label>
					<div class="relative">
						<Search
							class="pointer-events-none absolute inset-s-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
						/>
						<Input
							id="search"
							name="q"
							type="search"
							placeholder="Search"
							autocomplete="off"
							bind:value={searchValue}
							oninput={(e) => handleInput(e.currentTarget.value)}
							onfocus={handleFocus}
							onkeydown={handleKeydown}
							class="h-10 ps-9"
						/>
						{#if previewLoading}
							<Loader2
								class="absolute inset-e-3 top-1/2 size-4 -translate-y-1/2 animate-spin text-muted-foreground"
							/>
						{/if}
					</div>

					{#if previewOpen && searchValue.trim()}
						<div
							role="listbox"
							aria-label="Search results"
							class="absolute inset-x-0 top-full mt-1.5 max-h-96 overflow-y-auto rounded-lg border border-border bg-popover p-1.5 text-popover-foreground shadow-lg"
						>
							{#if previewResults.length === 0 && !previewLoading}
								<p class="px-2.5 py-3 text-center text-sm text-muted-foreground">
									No archives found.
								</p>
							{:else}
								{#each previewResults as archive (archive.id)}
									<a
										role="option"
										aria-selected="false"
										href={`${resolve(`/a/${archive.id}`)}?from=${encodeURIComponent(page.url.pathname + page.url.search)}`}
										onclick={() => (previewOpen = false)}
										class="flex items-center gap-2.5 rounded-md p-1.5 hover:bg-accent hover:text-accent-foreground"
									>
										<img
											src="/api/archives/{archive.id}/cover"
											alt=""
											class="h-12 w-9 shrink-0 rounded object-cover"
										/>
										<div class="min-w-0">
											<p class="truncate text-sm font-medium">{archive.title}</p>
											{#if archive.artists?.length}
												<p class="truncate text-xs text-muted-foreground">
													{archive.artists.join(', ')}
												</p>
											{/if}
										</div>
									</a>
								{/each}
								<button
									type="submit"
									class="w-full rounded-md p-1.5 text-left text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground"
								>
									See all results for "{searchValue.trim()}"
								</button>
							{/if}
						</div>
					{/if}
				</form>
			</div>

			<div class="w-6" aria-hidden="true"></div>
		</header>

		<main class="flex-1">
			{@render children()}
		</main>
	</Sidebar.Inset>
</Sidebar.Provider>
