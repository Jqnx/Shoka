<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import AddLibraryDialog from '$lib/components/add-library-dialog.svelte';
	import { enhance } from '$app/forms';
	import { FolderOpen, RefreshCw, Settings, Trash2 } from '@lucide/svelte';

	let { data, form } = $props();

	function confirmDelete(e: SubmitEvent, name: string) {
		if (!confirm(`Delete library "${name}"? This removes all its archives from Shoka, but not from disk.`)) {
			e.preventDefault();
		}
	}

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
			<p class="mt-1 text-sm text-muted-foreground">
				Folders on disk that Shoka scans for archives.
			</p>
		</div>

		<AddLibraryDialog types={data.libraryTypes} />
	</div>

	{#if form?.error}
		<p class="mt-4 text-sm text-destructive">{form.error}</p>
	{/if}

	<div class="mt-6 flex flex-col gap-3">
		{#if data.libraries.length === 0}
			<div
				class="flex flex-col items-center justify-center rounded-xl border border-dashed border-border py-16 text-center"
			>
				<FolderOpen class="mb-3 size-10 text-muted-foreground/40" />
				<p class="text-base font-medium text-foreground">No libraries yet</p>
				<p class="mt-1 text-sm text-muted-foreground">Add a folder to start building your library.</p>
			</div>
		{:else}
			{#each data.libraries as library (library.id)}
				<div
					class="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border p-4"
				>
					<div class="min-w-0">
						<div class="flex items-center gap-2">
							<span class="truncate font-medium">{library.name}</span>
							<Badge
								variant="outline"
								class={library.enabled
									? 'border-transparent bg-emerald-600/10 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-400'
									: 'border-transparent bg-destructive/10 text-destructive dark:bg-destructive/20'}
							>
								{library.enabled ? 'Enabled' : 'Disabled'}
							</Badge>
						</div>
						<p class="mt-0.5 truncate text-sm text-muted-foreground">{typeLabel(library.type)}</p>
					</div>

					<div class="flex shrink-0 items-center gap-2">
						<Button href="/admin/libraries/{library.id}" variant="outline" size="icon-sm">
							<Settings />
							<span class="sr-only">Settings</span>
						</Button>

						<form method="POST" action="?/scan" use:enhance>
							<input type="hidden" name="id" value={library.id} />
							<Button type="submit" variant="outline" size="sm">
								<RefreshCw />
								Scan
							</Button>
						</form>

						<form method="POST" action="?/toggle" use:enhance>
							<input type="hidden" name="id" value={library.id} />
							<input type="hidden" name="enabled" value={String(library.enabled)} />
							<Button type="submit" variant="outline" size="sm">
								{library.enabled ? 'Disable' : 'Enable'}
							</Button>
						</form>

						<form
							method="POST"
							action="?/delete"
							use:enhance
							onsubmit={(e) => confirmDelete(e, library.name)}
						>
							<input type="hidden" name="id" value={library.id} />
							<Button type="submit" variant="destructive" size="icon-sm">
								<Trash2 />
								<span class="sr-only">Delete</span>
							</Button>
						</form>
					</div>
				</div>
			{/each}
		{/if}
	</div>
</div>
