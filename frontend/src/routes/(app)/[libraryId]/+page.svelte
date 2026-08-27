<script lang="ts">
	import * as Pagination from '$lib/components/ui/pagination';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as Select from '$lib/components/ui/select';
	import ArchiveCard from '$lib/components/ArchiveCard.svelte';
	import ArchivesSidebar from '$lib/components/ArchivesSidebar.svelte';
	import FiltersTrigger from '$lib/components/FiltersTrigger.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import BulkEditDialog from '$lib/components/BulkEditDialog.svelte';
	import { Search, CheckSquare, Square, Ellipsis, Check, RotateCcw, Sparkles } from '@lucide/svelte';
	import { enhance } from '$app/forms';
	import type { ActionResult } from '@sveltejs/kit';
	import { cn } from '$lib/utils.js';
	import { goto, afterNavigate } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { SvelteSet } from 'svelte/reactivity';

	let { data, form } = $props();

	let selectionMode = $state(false);
	let bulkPending = $state(false);
	const selectedIds = new SvelteSet<string>();

	// Selection is per-view: ids from a page you've navigated away from would
	// stay selected but invisible, so any navigation (paging, filtering,
	// sorting) resets it rather than hiding state from the user.
	afterNavigate((nav) => {
		if (nav.type === 'goto') window.scrollTo({ top: 0, behavior: 'smooth' });
		exitSelection();
	});

	const selectedCount = $derived(selectedIds.size);
	// Ordered by how the grid shows them, so the hidden inputs (and the
	// backend's per-id failure list) line up with what the user sees.
	const selectedArchives = $derived(data.archives.filter((a) => selectedIds.has(a.id)));
	const selectedList = $derived(selectedArchives.map((a) => a.id));

	let identifyOpen = $state(false);

	// '' submits no source at all, which is what the backend reads as "run the
	// normal priority-ordered pipeline".
	const ALL_SOURCES = '';
	let identifySource = $state(ALL_SOURCES);

	// Lowest priority number first, matching the order the pipeline itself
	// would try them in.
	const sourceOptions = $derived(
		[...data.metadataSources].sort((a, b) => a.priority - b.priority)
	);
	const identifySourceLabel = $derived(
		identifySource === ALL_SOURCES ? 'All sources' : identifySource
	);

	const BULK_LABELS: Record<string, string> = {
		markRead: 'marked as read',
		markUnread: 'marked as unread',
		identify: 'queued for identification',
		edit: 'updated'
	};
	const allSelected = $derived(
		data.archives.length > 0 && selectedCount === data.archives.length
	);

	function exitSelection() {
		selectionMode = false;
		selectedIds.clear();
	}

	function startSelection(id: string) {
		selectionMode = true;
		selectedIds.add(id);
	}

	function toggleSelection(id: string) {
		if (selectedIds.has(id)) selectedIds.delete(id);
		else selectedIds.add(id);
		// Emptying the selection leaves selection mode with nothing to act on,
		// so drop straight back to normal browsing. Handled here rather than in
		// an $effect so it can't race startSelection, which flips the mode on
		// before the first id lands.
		if (selectedIds.size === 0) exitSelection();
	}

	function selectAll() {
		selectedIds.clear();
		for (const a of data.archives) selectedIds.add(a.id);
	}

	// Shared enhance handler for the dropdown's bulk forms: every one of them
	// acts on the current selection and is done with it afterwards - but only
	// on success. A failed action keeps bulkPending set until the list has
	// actually refreshed, and leaves the selection intact so the user can
	// retry without rebuilding it.
	function bulkSubmit() {
		bulkPending = true;
		return async ({
			result,
			update
		}: {
			result: ActionResult;
			update: () => Promise<void>;
		}) => {
			await update();
			bulkPending = false;
			if (result.type !== 'failure') {
				exitSelection();
			}
		};
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && selectionMode) exitSelection();
	}

	const DEFAULT_SORT = 'created_at_desc';

	const sortValue = $derived(data.filters.sort || DEFAULT_SORT);
	const sortLabel = $derived(
		data.sortOptions.find((opt) => opt.value === sortValue)?.display_name ?? 'Sort'
	);

	function updateParam(name: string, value: string) {
		const params = new URLSearchParams(page.url.searchParams);
		params.set(name, value);
		params.delete('page');
		goto(resolve(`${page.url.pathname}?${params}`), { noScroll: true, keepFocus: true });
	}
</script>

{#snippet bulkItem(action: string, label: string, Icon: typeof Check, read?: string)}
	<form method="POST" {action} use:enhance={bulkSubmit}>
		{#each selectedList as id (id)}
			<input type="hidden" name="ids" value={id} />
		{/each}
		{#if read !== undefined}
			<input type="hidden" name="read" value={read} />
		{/if}
		<DropdownMenu.Item disabled={bulkPending}>
			{#snippet child({ props })}
				<button type="submit" {...props} class={cn(props.class as string, 'w-full')}>
					<Icon />
					{label}
				</button>
			{/snippet}
		</DropdownMenu.Item>
	</form>
{/snippet}

<svelte:window onkeydown={handleKeydown} />

<svelte:head>
	<title>{data.library.name} | Shoka</title>
</svelte:head>

<Sidebar.Provider open={false}>
	<Sidebar.Inset class="bg-transparent">
		<div class="flex flex-wrap items-center justify-between gap-2 px-4 py-3 sm:px-6">
			<div>
				{#if selectionMode}
					<h1 class="text-base font-semibold">
						{selectedCount}
						{selectedCount === 1 ? 'archive' : 'archives'} selected
					</h1>
					<p class="text-sm text-muted-foreground">
						Press Esc to exit.
					</p>
				{:else}
					<h1 class="text-base font-semibold">{data.library.name}</h1>
					<p class="text-sm text-muted-foreground">
						{data.total}
						{data.total === 1 ? 'archive' : 'archives'} found
					</p>
				{/if}
			</div>

			<div class="flex items-center gap-2">
				{#if selectionMode}
					<BulkEditDialog
						ids={selectedList}
						categories={data.categories}
						languages={data.languages}
						allArtists={data.artists}
						allTags={data.tags}
						allCharacters={data.characters}
						allParodies={data.parodies}
						onDone={exitSelection}
					/>

					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							{#snippet child({ props })}
								<Button {...props} variant="outline" size="sm" disabled={bulkPending}>
									<Ellipsis class="size-3.5" />
									<span class="sr-only">More actions</span>
								</Button>
							{/snippet}
						</DropdownMenu.Trigger>
						<DropdownMenu.Content align="end" class="min-w-44">
							{@render bulkItem('?/bulkProgress', 'Mark as read', Check, 'true')}
							{@render bulkItem('?/bulkProgress', 'Mark as unread', RotateCcw, 'false')}
							<DropdownMenu.Separator />
							<DropdownMenu.Item
								disabled={bulkPending}
								onSelect={() => (identifyOpen = true)}
							>
								<Sparkles />
								Identify
							</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Root>

					<AlertDialog.Root
						bind:open={identifyOpen}
						onOpenChange={(v) => {
							if (v) identifySource = ALL_SOURCES;
						}}
					>
						<AlertDialog.Content class="sm:max-w-lg">
							<AlertDialog.Header>
								<AlertDialog.Title>
									Identify {selectedCount}
									{selectedCount === 1 ? 'archive' : 'archives'}?
								</AlertDialog.Title>
								<AlertDialog.Description>
									Metadata is written straight to these archives — there's no per-archive
									confirmation and no undo. Existing values may be overwritten.
								</AlertDialog.Description>
							</AlertDialog.Header>

							<div class="flex flex-col gap-1.5">
								<span class="text-sm font-medium">Source</span>
								<Select.Root
									type="single"
									value={identifySource}
									onValueChange={(v) => (identifySource = v)}
								>
									<Select.Trigger class="w-full">{identifySourceLabel}</Select.Trigger>
									<Select.Content>
										<Select.Item value={ALL_SOURCES} label="All sources">
											All sources
										</Select.Item>
										{#each sourceOptions as source (source.name)}
											<Select.Item
												value={source.name}
												label={source.name}
												disabled={!source.enabled}
											>
												{source.name}{source.enabled ? '' : ' (disabled)'}
											</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
								<p class="text-xs text-muted-foreground">
									{identifySource === ALL_SOURCES
										? 'Runs every enabled source in priority order, each filling gaps the last left.'
										: `Runs only ${identifySource}; nothing else is tried if it finds nothing.`}
								</p>
							</div>

							<div class="max-h-56 overflow-y-auto rounded-lg border border-border">
								<ul class="divide-y divide-border text-sm">
									{#each selectedArchives as archive (archive.id)}
										<li class="truncate px-2.5 py-1.5" title={archive.title}>{archive.title}</li>
									{/each}
								</ul>
							</div>

							<form method="POST" action="?/bulkIdentify" use:enhance={bulkSubmit}>
								{#each selectedList as id (id)}
									<input type="hidden" name="ids" value={id} />
								{/each}
								<input type="hidden" name="source" value={identifySource} />
								<AlertDialog.Footer>
									<AlertDialog.Cancel type="button">Cancel</AlertDialog.Cancel>
									<AlertDialog.Action type="submit" disabled={bulkPending}>
										{bulkPending ? 'Queuing…' : `Identify ${selectedCount}`}
									</AlertDialog.Action>
								</AlertDialog.Footer>
							</form>
						</AlertDialog.Content>
					</AlertDialog.Root>

					<Button
						variant="outline"
						size="sm"
						class="gap-1.5"
						onclick={() => (allSelected ? exitSelection() : selectAll())}
					>
						{#if allSelected}
							<Square class="size-3.5" />
							Unselect all
						{:else}
							<CheckSquare class="size-3.5" />
							Select all
						{/if}
					</Button>
				{:else}
					<Select.Root type="single" value={sortValue} onValueChange={(v) => updateParam('sort', v)}>
						<Select.Trigger class="h-8 w-[190px]">
							{sortLabel}
						</Select.Trigger>
						<Select.Content>
							<Select.Group>
								<Select.Label>Sort</Select.Label>
								{#each data.sortOptions as opt (opt.value)}
									<Select.Item value={opt.value} label={opt.display_name}>
										{opt.display_name}
									</Select.Item>
								{/each}
							</Select.Group>
						</Select.Content>
					</Select.Root>

					<FiltersTrigger />
				{/if}
			</div>
		</div>

		{#if form?.action && BULK_LABELS[form.action]}
			<div class="px-4 sm:px-6">
				{#if form.success}
					<p class="text-sm text-muted-foreground">
						{form.succeeded} of {form.requested}
						{form.requested === 1 ? 'archive' : 'archives'}
						{BULK_LABELS[form.action]}{form.failed > 0 ? ` · ${form.failed} failed` : ''}.
					</p>
				{/if}
			</div>
		{:else if form?.error}
			<div class="px-4 sm:px-6">
				<p class="text-sm text-destructive">{form.error}</p>
			</div>
		{/if}

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
						<ArchiveCard
							{archive}
							{selectionMode}
							selected={selectedIds.has(archive.id)}
							onLongPress={startSelection}
							onToggle={toggleSelection}
						/>
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

	<ArchivesSidebar
		filters={data.filters}
		artists={data.artists}
		tags={data.tags}
		characters={data.characters}
		parodies={data.parodies}
		categories={data.categories}
		languages={data.languages}
	/>
</Sidebar.Provider>
