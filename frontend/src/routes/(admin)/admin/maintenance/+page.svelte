<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { enhance } from '$app/forms';
	import { ImageUp, Binary } from '@lucide/svelte';

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
			label: 'Regenerate covers',
			success: 'Cover generation enqueued for all archives.'
		},
		{
			task: 'phash' as const,
			action: 'generatePHashes',
			icon: Binary,
			title: 'Generate perceptual hashes',
			description: 'Generate perceptual hashes for every archive. Used for duplicate detection.',
			label: 'Generate hashes',
			success: 'Hash generation enqueued for all archives.'
		}
	];

	// The backend rejects a non-zero interval below 15 minutes; maintenance runs
	// are whole-library passes, so the presets start coarser than the scan list.
	const intervals = [
		{ value: '0', label: 'Off' },
		{ value: '360', label: 'Every 6 hours' },
		{ value: '720', label: 'Every 12 hours' },
		{ value: '1440', label: 'Every 24 hours' },
		{ value: '10080', label: 'Every week' }
	];

	const tasksById = $derived(new Map(data.tasks.map((t) => [t.task, t])));
	// String state because Select.Root carries its value as a string (same as the
	// scan-interval select on the library page). Seeded once and never re-synced
	// from `data`: a successful save leaves it at the value the server now holds,
	// and a failed save keeps the user's pick with the error line above the list.
	// Built from `jobs`, not the response, so a task row deleted by hand still
	// renders as Off instead of an undefined interval.
	let schedule = $state<Record<string, string>>(
		Object.fromEntries(
			jobs.map((job) => [job.task, String(tasksById.get(job.task)?.interval_minutes ?? 0)])
		)
	);

	function scheduleLabel(task: string) {
		const value = schedule[task];
		const preset = intervals.find((i) => i.value === value);

		// A value set through the API that is not one of the presets still has to
		// read truthfully; falling back to "Off" would claim scheduling is disabled.
		return preset?.label ?? `Every ${value} minutes`;
	}

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

	{#if form?.action === 'setSchedule'}
		{#if form.error}
			<p class="mt-4 text-sm text-destructive">{form.error}</p>
		{:else}
			<p class="mt-4 text-sm text-muted-foreground">Schedule saved.</p>
		{/if}
	{/if}

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

				<div class="flex shrink-0 flex-wrap items-center gap-2">
					<form
						method="POST"
						action="?/setSchedule"
						use:enhance={() => {
							pending = 'setSchedule';
							return async ({ update }) => {
								await update();
								pending = null;
							};
						}}
						class="flex items-center gap-2"
					>
						<input type="hidden" name="task" value={job.task} />
						<Select.Root type="single" name="interval_minutes" bind:value={schedule[job.task]}>
							<Select.Trigger id={`schedule-${job.task}`} class="w-44" aria-label="Schedule">
								{scheduleLabel(job.task)}
							</Select.Trigger>
							<Select.Content>
								{#each intervals as interval (interval.value)}
									<Select.Item value={interval.value} label={interval.label}>
										{interval.label}
									</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
						<Button type="submit" variant="outline" size="sm" disabled={pending !== null}>
							Save
						</Button>
					</form>

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
			</div>
		{/each}
	</div>
</div>
