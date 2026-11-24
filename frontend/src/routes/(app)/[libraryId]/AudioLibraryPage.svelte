<script lang="ts">
	import * as Pagination from '$lib/components/ui/pagination';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import AudioCard from '$lib/components/AudioCard.svelte';
	import AudioSidebar from '$lib/components/AudioSidebar.svelte';
	import FiltersTrigger from '$lib/components/FiltersTrigger.svelte';
	import BulkEditAudioDialog from '$lib/components/BulkEditAudioDialog.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { AudioLines, CheckSquare, Square, Ellipsis, Check, RotateCcw } from '@lucide/svelte';
	import { enhance } from '$app/forms';
	import type { ActionResult } from '@sveltejs/kit';
	import type { ActionData, PageData } from './$types';
	import { cn } from '$lib/utils.js';
	import { goto, afterNavigate } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { can } from '$lib/auth/permissions';
	import { SvelteSet } from 'svelte/reactivity';

	type AudioPageData = Extract<PageData, { kind: 'audio' }>;

	let { data, form }: { data: AudioPageData; form: ActionData } = $props();

	const canEdit = $derived(can(page.data.user, { audio: ['update'] }));

	let selectionMode = $state(false);
	let bulkPending = $state(false);
	const selectedIds = new SvelteSet<string>();

	// Selection is per-view: ids from a page you've navigated away from would
	// stay selected but invisible, so any navigation (paging, filtering)
	// resets it rather than hiding state from the user.
	afterNavigate((nav) => {
		if (nav.type === 'goto') window.scrollTo({ top: 0, behavior: 'smooth' });
		exitSelection();
	});

	const selectedCount = $derived(selectedIds.size);
	// Ordered by how the grid shows them, so the hidden inputs (and the
	// backend's per-id failure list) line up with what the user sees.
	const selectedWorks = $derived(data.audio.filter((w) => selectedIds.has(w.id)));
	const selectedList = $derived(selectedWorks.map((w) => w.id));

	const BULK_LABELS: Record<string, string> = {
		markPlayed: 'marked as played',
		markUnplayed: 'marked as unplayed',
		audioEdit: 'updated'
	};
	const allSelected = $derived(data.audio.length > 0 && selectedCount === data.audio.length);

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
		for (const work of data.audio) selectedIds.add(work.id);
	}

	// Shared enhance handler for the dropdown's bulk forms: every one of them
	// acts on the current selection and is done with it afterwards - but only
	// on success. A failed action keeps bulkPending set until the list has
	// actually refreshed, and leaves the selection intact so the user can
	// retry without rebuilding it.
	function bulkSubmit() {
		bulkPending = true;
		return async ({ result, update }: { result: ActionResult; update: () => Promise<void> }) => {
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
</script>

{#snippet bulkItem(action: string, label: string, Icon: typeof Check, played?: string)}
	<form method="POST" {action} use:enhance={bulkSubmit}>
		{#each selectedList as id (id)}
			<input type="hidden" name="ids" value={id} />
		{/each}
		{#if played !== undefined}
			<input type="hidden" name="played" value={played} />
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
						{selectedCount === 1 ? 'work' : 'works'} selected
					</h1>
					<p class="text-sm text-muted-foreground">Press Esc to exit.</p>
				{:else}
					<h1 class="text-base font-semibold">{data.library.name}</h1>
					<p class="text-sm text-muted-foreground">
						{data.total}
						{data.total === 1 ? 'work' : 'works'} found
					</p>
				{/if}
			</div>

			<div class="flex items-center gap-2">
				{#if selectionMode}
					{#if canEdit}
						<BulkEditAudioDialog
							ids={selectedList}
							languages={data.languages}
							allArtists={data.artists}
							allTags={data.tags}
							allCircles={data.allCircles}
							allVoiceActors={data.voiceActors}
							onDone={exitSelection}
						/>
					{/if}

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
							{@render bulkItem('?/bulkAudioPlayed', 'Mark as played', Check, 'true')}
							{@render bulkItem('?/bulkAudioPlayed', 'Mark as unplayed', RotateCcw, 'false')}
						</DropdownMenu.Content>
					</DropdownMenu.Root>

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
					<FiltersTrigger />
				{/if}
			</div>
		</div>

		{#if form?.action && BULK_LABELS[form.action]}
			<div class="px-4 sm:px-6">
				{#if form.success}
					<p class="text-sm text-muted-foreground">
						{form.succeeded} of {form.requested}
						{form.requested === 1 ? 'work' : 'works'}
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
			{#if data.audio.length === 0}
				<div class="flex flex-col items-center justify-center py-24 text-center">
					<AudioLines class="mb-3 size-10 text-muted-foreground/40" />
					<p class="text-base font-medium text-foreground">No audio found</p>
				</div>
			{:else}
				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
					{#each data.audio as work (work.id)}
						<AudioCard
							audio={work}
							{selectionMode}
							selected={selectedIds.has(work.id)}
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

	<AudioSidebar
		filters={data.filters}
		artists={data.artists}
		tags={data.tags}
		circles={data.allCircles}
		voiceActors={data.voiceActors}
	/>
</Sidebar.Provider>
