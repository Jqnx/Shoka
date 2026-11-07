<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import { ArrowUp, ChevronRight, Folder, Loader2 } from '@lucide/svelte';
	import { errorMessage } from '$lib/api';
	import type { LibraryBrowseEntry, LibraryBrowseResponse } from '$lib/types';

	let {
		open = $bindable(false),
		path = $bindable('')
	}: {
		open?: boolean;
		/** The form's path value: the initial directory, and what "Select this folder" writes. */
		path?: string;
	} = $props();

	let current = $state('/');
	let parent = $state<string | undefined>(undefined);
	let entries = $state<LibraryBrowseEntry[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);

	// Incremented per request so a slow reply can't overwrite a newer directory.
	let requestID = 0;

	async function browse(dir: string) {
		const request = ++requestID;
		loading = true;
		error = null;

		try {
			const res = await fetch(`/api/admin/libraries/browse?path=${encodeURIComponent(dir)}`);
			if (request !== requestID) return;

			if (!res.ok) {
				const message = await errorMessage(res, 'Failed to browse folders.');
				if (request !== requestID) return;

				error = message;

				return;
			}

			const body: LibraryBrowseResponse = await res.json();
			if (request !== requestID) return;

			current = body.path;
			parent = body.parent;
			entries = body.entries ?? [];
		} catch {
			if (request === requestID) error = 'Failed to browse folders.';
		} finally {
			if (request === requestID) loading = false;
		}
	}

	$effect(() => {
		if (!open) return;

		// Start from the requested directory: a failed listing must not leave
		// current (and therefore Select) pointing at a previously browsed one.
		// Use a local so current is not an effect dependency - browse() writes
		// it on every navigation.
		const requested = path.trim() || '/';
		current = requested;

		void browse(requested);
	});

	const crumbs = $derived.by(() => {
		const trail = [{ name: '/', path: '/' }];
		let acc = '';

		for (const segment of current.split('/').filter(Boolean)) {
			acc += `/${segment}`;
			trail.push({ name: segment, path: acc });
		}

		return trail;
	});

	function select() {
		path = current;
		open = false;
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-2xl">
		<Dialog.Header>
			<Dialog.Title>Choose a folder</Dialog.Title>
			<Dialog.Description>Pick the folder Shoka should scan.</Dialog.Description>
		</Dialog.Header>

		<div class="flex items-center gap-1">
			<Button
				type="button"
				variant="ghost"
				size="icon-xs"
				aria-label="Parent folder"
				disabled={!parent}
				onclick={() => parent && browse(parent)}
			>
				<ArrowUp />
			</Button>
			<Breadcrumb.Root class="min-w-0">
				<Breadcrumb.List>
					{#each crumbs as crumb, i (crumb.path)}
						{#if i > 0}
							<Breadcrumb.Separator />
						{/if}
						<Breadcrumb.Item>
							{#if i === crumbs.length - 1 && !error}
								<Breadcrumb.Page>{crumb.name}</Breadcrumb.Page>
							{:else}
								<button
									type="button"
									class="transition-colors hover:text-foreground"
									onclick={() => browse(crumb.path)}
								>
									{crumb.name}
								</button>
							{/if}
						</Breadcrumb.Item>
					{/each}
				</Breadcrumb.List>
			</Breadcrumb.Root>
		</div>

		<ScrollArea class="h-[50vh] border border-border">
			<div class="p-1">
				{#if loading}
					<div class="flex items-center justify-center gap-2 py-8 text-xs text-muted-foreground">
						<Loader2 class="size-4 animate-spin" />
						Loading folders…
					</div>
				{:else if error}
					<p class="px-2 py-4 text-sm text-destructive">{error}</p>
				{:else if entries.length === 0}
					<p class="px-2 py-4 text-xs text-muted-foreground">No subfolders.</p>
				{:else}
					{#each entries as entry (entry.path)}
						<Button
							type="button"
							variant="ghost"
							class="h-auto w-full justify-start gap-2 px-2 py-1.5 font-normal"
							onclick={() => browse(entry.path)}
						>
							<Folder class="text-muted-foreground" />
							<span class="truncate">{entry.name}</span>
							<ChevronRight class="ms-auto text-muted-foreground" />
						</Button>
					{/each}
				{/if}
			</div>
		</ScrollArea>

		<Dialog.Footer class="flex-col items-stretch sm:flex-row sm:items-center sm:justify-between">
			<p class="truncate font-mono text-xs text-muted-foreground" title={current}>{current}</p>
			<Button type="button" disabled={loading || error !== null} onclick={select}>
				Select this folder
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
