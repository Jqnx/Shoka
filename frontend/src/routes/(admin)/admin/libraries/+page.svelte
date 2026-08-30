<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import AddLibraryDialog from '$lib/components/add-library-dialog.svelte';
	import { enhance } from '$app/forms';
	import { cn } from '$lib/utils.js';
	import type { Library } from '$lib/types';
	import { Ellipsis, FolderOpen, ImageUp, RefreshCw, Settings, Trash2 } from '@lucide/svelte';

	let { data, form } = $props();

	// The dialog lives outside the {#each} so it isn't torn down with the
	// dropdown that opens it; the row is carried here instead.
	let deleteTarget = $state<Library | null>(null);
	let regenerating = $state<string | null>(null);

	function typeLabel(type: string) {
		return type.charAt(0).toUpperCase() + type.slice(1);
	}
</script>

<svelte:head>
	<title>Libraries | Shoka Admin</title>
</svelte:head>

<div class="p-4 sm:p-6">
	<div class="flex flex-wrap items-center justify-between gap-2">
		<div>
			<h1 class="text-2xl font-bold tracking-tight">Libraries</h1>
			<p class="mt-1 text-sm text-muted-foreground">Folders to scan for media.</p>
		</div>

		<AddLibraryDialog types={data.libraryTypes} />
	</div>

	{#if form?.error}
		<p class="mt-4 text-sm text-destructive">{form.error}</p>
	{:else if form?.action === 'regenerateCovers' && form.success}
		<p class="mt-4 text-sm text-muted-foreground">
			Cover generation enqueued for the library's archives.
		</p>
	{/if}

	<div class="mt-6 flex flex-col gap-3">
		{#if data.libraries.length === 0}
			<div
				class="flex flex-col items-center justify-center rounded-xl border border-dashed border-border py-16 text-center"
			>
				<FolderOpen class="mb-3 size-10 text-muted-foreground/40" />
				<p class="text-base font-medium text-foreground">No libraries yet</p>
				<p class="mt-1 text-sm text-muted-foreground">
					Add a folder to start building your library.
				</p>
			</div>
		{:else}
			{#each data.libraries as library (library.id)}
				<div
					class="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border p-4"
				>
					<div class="min-w-0">
						<div class="flex items-center gap-2">
							<span class="truncate font-medium">{library.name}</span>
						</div>
						<p class="mt-0.5 truncate text-sm text-muted-foreground">{typeLabel(library.type)}</p>
					</div>

					<div class="flex shrink-0 items-center gap-2">
						<form method="POST" action="?/scan" use:enhance>
							<input type="hidden" name="id" value={library.id} />
							<Button type="submit" variant="outline" size="sm">
								<RefreshCw />
								Scan
							</Button>
						</form>

						<Button href="/admin/libraries/{library.id}" variant="outline" size="icon-sm">
							<Settings />
							<span class="sr-only">Settings</span>
						</Button>

						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								{#snippet child({ props })}
									<Button {...props} variant="outline" size="icon-sm" aria-label="More actions">
										<Ellipsis />
									</Button>
								{/snippet}
							</DropdownMenu.Trigger>
							<DropdownMenu.Content align="end" class="min-w-48">
								<form
									method="POST"
									action="?/regenerateCovers"
									use:enhance={() => {
										regenerating = library.id;
										return async ({ update }) => {
											regenerating = null;
											await update();
										};
									}}
								>
									<input type="hidden" name="id" value={library.id} />
									<DropdownMenu.Item disabled={regenerating === library.id}>
										{#snippet child({ props })}
											<button type="submit" {...props} class={cn(props.class as string, 'w-full')}>
												<ImageUp />
												{regenerating === library.id ? 'Enqueuing…' : 'Regenerate covers'}
											</button>
										{/snippet}
									</DropdownMenu.Item>
								</form>
								<DropdownMenu.Separator />
								<DropdownMenu.Item
									variant="destructive"
									onSelect={() => {
										deleteTarget = library;
									}}
								>
									<Trash2 />
									Delete
								</DropdownMenu.Item>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					</div>
				</div>
			{/each}
		{/if}
	</div>
</div>

<AlertDialog.Root
	open={deleteTarget !== null}
	onOpenChange={(open) => {
		if (!open) deleteTarget = null;
	}}
>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Delete library "{deleteTarget?.name}"?</AlertDialog.Title>
			<AlertDialog.Description>
				This removes all its archives from Shoka, but not from disk.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<form method="POST" action="?/delete" use:enhance>
			<input type="hidden" name="id" value={deleteTarget?.id ?? ''} />
			<AlertDialog.Footer>
				<AlertDialog.Cancel type="button">Cancel</AlertDialog.Cancel>
				<AlertDialog.Action type="submit" variant="destructive">Delete</AlertDialog.Action>
			</AlertDialog.Footer>
		</form>
	</AlertDialog.Content>
</AlertDialog.Root>
