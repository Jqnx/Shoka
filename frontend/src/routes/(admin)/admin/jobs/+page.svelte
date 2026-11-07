<script lang="ts">
	import type { JobSnapshot, JobStatus, JobView } from '$lib/types';
	import { Badge } from '$lib/components/ui/badge';
	import { cn } from '$lib/utils.js';
	import { CircleAlert, Radio } from '@lucide/svelte';

	let { data } = $props();

	// The stream replaces the snapshot wholesale, so there is no delta state
	// to reconcile. Derived over the loaded data rather than seeding $state
	// from it, so a fresh server load is picked up on re-navigation instead
	// of the component sticking with a stale initial value.
	let streamed = $state<JobSnapshot | null>(null);
	const snapshot = $derived(streamed ?? data.snapshot);
	let live = $state(false);

	// Elapsed times have to advance on their own: an idle queue produces no
	// new snapshots, so without this a long-running job's timer would freeze.
	let now = $state(Date.now());

	$effect(() => {
		const source = new EventSource('/api/admin/jobs/events');

		source.addEventListener('snapshot', (e) => {
			streamed = JSON.parse((e as MessageEvent).data) as JobSnapshot;
			live = true;
		});

		source.onerror = () => {
			live = false;
			// EventSource retries on its own unless it has given up entirely.
			if (source.readyState === EventSource.CLOSED) source.close();
		};

		return () => source.close();
	});

	$effect(() => {
		const id = setInterval(() => (now = Date.now()), 1000);
		return () => clearInterval(id);
	});

	const tiles: { status: JobStatus; label: string; class: string }[] = [
		{ status: 'running', label: 'Running', class: 'text-blue-500' },
		{ status: 'pending', label: 'Pending', class: 'text-amber-500' },
		{ status: 'done', label: 'Done', class: 'text-emerald-500' },
		{ status: 'failed', label: 'Failed', class: 'text-destructive' }
	];

	function badgeVariant(status: JobStatus) {
		if (status === 'failed') return 'destructive' as const;
		if (status === 'running') return 'default' as const;
		return 'secondary' as const;
	}

	function elapsed(iso: string) {
		const seconds = Math.max(0, Math.round((now - new Date(iso).getTime()) / 1000));
		if (seconds < 60) return `${seconds}s`;

		const minutes = Math.floor(seconds / 60);
		if (minutes < 60) return `${minutes}m ${seconds % 60}s`;

		// Failures are kept until purged, so these ages run to days - without
		// the rollover a week-old failure reads as "174h 6m".
		const hours = Math.floor(minutes / 60);
		if (hours < 24) return `${hours}h ${minutes % 60}m`;

		return `${Math.floor(hours / 24)}d ${hours % 24}h`;
	}

	// Attempt 1 of 3 is the normal first run, not a retry - only worth
	// surfacing once a job has actually been retried.
	function isRetry(job: JobView) {
		return job.attempts > 1;
	}
</script>

<svelte:head>
	<title>Jobs | Shoka Admin</title>
</svelte:head>

<div class="p-4 sm:p-6">
	<div class="flex flex-wrap items-start justify-between gap-3">
		<div>
			<h1 class="text-2xl font-bold tracking-tight">Jobs</h1>
			<p class="mt-1 text-sm text-muted-foreground">
				Background work queued for this Shoka instance, updated live.
			</p>
		</div>

		<span
			class={cn(
				'inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs',
				live ? 'border-emerald-500/30 text-emerald-500' : 'border-border text-muted-foreground'
			)}
		>
			<Radio class="size-3.5" />
			{live ? 'Live' : 'Reconnecting…'}
		</span>
	</div>

	{#if data.error}
		<p class="mt-4 flex items-center gap-1.5 text-sm text-destructive">
			<CircleAlert class="size-4" />
			{data.error}
		</p>
	{/if}

	<div class="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
		{#each tiles as tile (tile.status)}
			<div class="rounded-xl border border-border p-4">
				<p class="text-xs text-muted-foreground">{tile.label}</p>
				<p class={cn('mt-1 text-2xl font-bold tabular-nums', tile.class)}>
					{snapshot.counts[tile.status] ?? 0}
				</p>
			</div>
		{/each}
	</div>

	<h2 class="mt-8 text-sm font-semibold">Active</h2>
	{#if snapshot.active.length === 0}
		<p
			class="mt-2 rounded-xl border border-dashed border-border p-6 text-center text-sm text-muted-foreground"
		>
			Nothing queued or running.
		</p>
	{:else}
		<div class="mt-2 divide-y divide-border overflow-hidden rounded-xl border border-border">
			{#each snapshot.active as job (job.id)}
				<div class="flex flex-wrap items-center gap-x-3 gap-y-1 p-3 text-sm">
					<Badge variant={badgeVariant(job.status)} class="shrink-0">{job.status}</Badge>
					<span class="font-medium">{job.type}</span>
					{#if job.target}
						<span class="truncate text-muted-foreground">{job.target}</span>
					{/if}
					{#if isRetry(job)}
						<span class="text-xs text-amber-500">
							attempt {job.attempts}/{job.max_attempts}
						</span>
					{/if}
					<span class="ms-auto shrink-0 text-muted-foreground tabular-nums">
						{elapsed(job.status === 'running' ? job.updated_at : job.created_at)}
					</span>
				</div>
			{/each}
		</div>
	{/if}

	{#if snapshot.recent_failures.length > 0}
		<h2 class="mt-8 text-sm font-semibold">Recent failures</h2>
		<div class="mt-2 divide-y divide-border overflow-hidden rounded-xl border border-border">
			{#each snapshot.recent_failures as job (job.id)}
				<div class="p-3 text-sm">
					<div class="flex flex-wrap items-center gap-x-3 gap-y-1">
						<span class="font-medium">{job.type}</span>
						{#if job.target}
							<span class="truncate text-muted-foreground">{job.target}</span>
						{/if}
						<span class="text-xs text-muted-foreground">
							after {job.attempts}/{job.max_attempts} attempts
						</span>
						<span class="ms-auto shrink-0 text-muted-foreground tabular-nums">
							{elapsed(job.updated_at)} ago
						</span>
					</div>
					{#if job.error}
						<p class="mt-1 font-mono text-xs break-words text-destructive">{job.error}</p>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
