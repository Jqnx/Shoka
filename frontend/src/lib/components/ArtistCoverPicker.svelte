<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import ArchiveCover from '$lib/components/ArchiveCover.svelte';
	import { enhance, applyAction } from '$app/forms';
	import type { SubmitFunction } from '@sveltejs/kit';
	import { ImagePlus } from '@lucide/svelte';
	import type { Archive, ArchiveListResponse } from '$lib/types';

	let {
		artistId,
		currentArchiveId
	}: {
		artistId: number;
		currentArchiveId?: string;
	} = $props();

	const LIMIT = 24;

	let open = $state(false);
	let page = $state(1);
	let items = $state<Archive[]>([]);
	let total = $state(0);
	let loading = $state(false);
	let saving = $state(false);
	let error = $state<string | null>(null);

	// Reload whenever the dialog opens or the page changes. Aborting on
	// cleanup means a fast page flip can't have an older response land last.
	$effect(() => {
		if (!open) return;
		const current = page;
		const ctrl = new AbortController();
		loading = true;
		error = null;
		fetch(`/api/artists/${artistId}/archives?page=${current}&limit=${LIMIT}`, {
			signal: ctrl.signal
		})
			.then(async (res) => {
				if (!res.ok) throw new Error('Failed to load archives.');
				const data: ArchiveListResponse = await res.json();
				items = data.items ?? [];
				total = Number(data.total);
			})
			.catch((e: unknown) => {
				if (!ctrl.signal.aborted) {
					error = e instanceof Error ? e.message : 'Failed to load archives.';
				}
			})
			.finally(() => {
				if (!ctrl.signal.aborted) loading = false;
			});
		return () => ctrl.abort();
	});

	const hasPrev = $derived(page > 1);
	const hasNext = $derived(page * LIMIT < total);

	const submitCover: SubmitFunction = () => {
		saving = true;
		error = null;
		return async ({ result, update }) => {
			saving = false;
			if (result.type === 'failure') {
				error = 'Failed to change the artist image.';
				return;
			}
			if (result.type === 'success') {
				await update();
				open = false;
				return;
			}
			// Anything else (an unhandled action exception) must still be
			// applied; a custom callback suppresses the default handling.
			await applyAction(result);
		};
	};
</script>

<Dialog.Root bind:open>
	<Dialog.Trigger>
		{#snippet child({ props })}
			<Button {...props} variant="outline" class="mt-3 w-full gap-2">
				<ImagePlus class="size-4" />
				Change image
			</Button>
		{/snippet}
	</Dialog.Trigger>
	<Dialog.Content class="sm:max-w-2xl">
		<Dialog.Header>
			<Dialog.Title>Artist image</Dialog.Title>
			<Dialog.Description>
				Pick which of this artist's archives supplies the thumbnail.
			</Dialog.Description>
		</Dialog.Header>

		{#if loading}
			<p class="py-6 text-center text-sm text-muted-foreground">Loading…</p>
		{:else if items.length === 0}
			<p class="py-6 text-center text-sm text-muted-foreground">No archives found.</p>
		{:else}
			<div class="grid max-h-[60vh] grid-cols-3 gap-3 overflow-y-auto py-2 sm:grid-cols-4">
				{#each items as archive (archive.id)}
					<form method="POST" action="?/setCover" use:enhance={submitCover}>
						<input type="hidden" name="archive_id" value={archive.id} />
						<button
							type="submit"
							disabled={saving}
							class="group relative aspect-[2/3] w-full overflow-hidden rounded-md border {archive.id ===
							currentArchiveId
								? 'border-primary ring-2 ring-primary'
								: 'border-border'}"
						>
							<ArchiveCover
								{archive}
								alt={archive.title}
								class="absolute inset-0 size-full object-cover"
							/>
						</button>
						<p class="mt-1 truncate text-xs text-muted-foreground">{archive.title}</p>
					</form>
				{/each}
			</div>
		{/if}

		{#if error}
			<p class="text-sm text-destructive">{error}</p>
		{/if}

		<Dialog.Footer class="flex items-center justify-between gap-2">
			<div class="flex gap-2">
				<Button
					variant="outline"
					size="sm"
					disabled={!hasPrev || loading}
					onclick={() => (page -= 1)}
				>
					Previous
				</Button>
				<Button
					variant="outline"
					size="sm"
					disabled={!hasNext || loading}
					onclick={() => (page += 1)}
				>
					Next
				</Button>
			</div>
			<form method="POST" action="?/setCover" use:enhance={submitCover}>
				<input type="hidden" name="archive_id" value="" />
				<Button type="submit" variant="secondary" disabled={saving}>Use default</Button>
			</form>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
