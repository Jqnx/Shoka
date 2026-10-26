<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { enhance } from '$app/forms';
	import { ImageUp, Binary, Loader2, Play } from '@lucide/svelte';

	let { data, form } = $props();

	// One shared pending marker rather than a flag per job: only one form can
	// be mid-submit at a time, and this keeps adding a job to the list below
	// a data-only change.
	let pending = $state<string | null>(null);

	const jobs = [
		{
			task: 'cover' as const,
			action: 'regenerateCovers',
			icon: ImageUp,
			title: 'Regenerate covers',
			description: 'Re-queue cover thumbnail generation for every archive in every library.',
			success: 'Cover generation enqueued for all archives.'
		},
		{
			task: 'phash' as const,
			action: 'generatePHashes',
			icon: Binary,
			title: 'Generate perceptual hashes',
			description: 'Generate perceptual hashes for every archive. Used for duplicate detection.',
			success: 'Hash generation enqueued for all archives.'
		}
	];

	const tasksById = $derived(new Map(data.tasks.map((t) => [t.task, t])));

	// "Last run" is the last enqueue attempt - manual or scheduled - not job
	// completion; the queue itself is visible on /admin/jobs.
	function lastRunLabel(at: string | undefined) {
		if (!at) return 'Last run: never';

		return `Last run: ${new Date(at).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })}`;
	}
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
						<p class="mt-1 text-xs text-muted-foreground">
							{lastRunLabel(tasksById.get(job.task)?.last_run_at)}
						</p>
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
					<Button
						type="submit"
						variant="outline"
						size="icon"
						disabled={pending !== null}
						aria-label={`Run now: ${job.title.toLowerCase()}`}
					>
						{#if pending === job.action}
							<Loader2 class="size-4 animate-spin" />
						{:else}
							<Play class="size-4" />
						{/if}
					</Button>
				</form>
			</div>
		{/each}
	</div>
</div>
