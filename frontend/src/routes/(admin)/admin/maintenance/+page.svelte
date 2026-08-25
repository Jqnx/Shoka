<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { enhance } from '$app/forms';
	import { ImageUp, Binary } from '@lucide/svelte';

	let { form } = $props();

	// One shared pending marker rather than a flag per job: only one form can
	// be mid-submit at a time, and this keeps adding a job to the list below
	// a data-only change.
	let pending = $state<string | null>(null);

	const jobs = [
		{
			action: 'regenerateCovers',
			icon: ImageUp,
			title: 'Regenerate covers',
			description: 'Re-queue cover thumbnail generation for every archive in every library.',
			label: 'Regenerate covers',
			success: 'Cover generation enqueued for all archives.'
		},
		{
			action: 'generatePHashes',
			icon: Binary,
			title: 'Generate perceptual hashes',
			description:
				"Generate perceptual hashes for every archive. Used for duplicate detection.",
			label: 'Generate hashes',
			success: 'Hash generation enqueued for all archives.'
		}
	];
</script>

<svelte:head>
	<title>Maintenance | Shoka Admin</title>
</svelte:head>

<div class="p-4 sm:p-6">
	<h1 class="text-2xl font-bold tracking-tight">Maintenance</h1>
	<p class="mt-1 text-sm text-muted-foreground">Housekeeping tasks for your Shoka instance.</p>

	<div class="mt-6 flex flex-col gap-3">
		{#each jobs as job (job.action)}
			<div
				class="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border p-4"
			>
				<div class="flex min-w-0 flex-1 items-start gap-3">
					<job.icon class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
					<div class="min-w-0">
						<p class="font-medium">{job.title}</p>
						<p class="mt-0.5 text-sm text-muted-foreground">{job.description}</p>
						{#if form?.action === job.action}
							{#if form.success}
								<p class="mt-1.5 text-sm text-muted-foreground">{job.success}</p>
							{:else if form.error}
								<p class="mt-1.5 text-sm text-destructive">{form.error}</p>
							{/if}
						{/if}
					</div>
				</div>

				<form
					method="POST"
					action="?/{job.action}"
					use:enhance={() => {
						pending = job.action;
						return async ({ update }) => {
							await update();
							pending = null;
						};
					}}
					class="shrink-0"
				>
					<Button type="submit" variant="outline" size="sm" disabled={pending !== null}>
						{pending === job.action ? 'Enqueuing…' : job.label}
					</Button>
				</form>
			</div>
		{/each}
	</div>
</div>
