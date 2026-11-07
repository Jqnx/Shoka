<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Activity, CopyCheck, Library, UsersRound, Wrench } from '@lucide/svelte';

	let { data } = $props();

	// The queue's own vocabulary: only statuses with work to report are named,
	// so an idle instance doesn't read "0 running · 0 pending".
	const jobSummary = $derived.by(() => {
		const counts = data.jobCounts;
		if (!counts) return '';

		const parts = [
			counts.running ? `${counts.running} running` : '',
			counts.pending ? `${counts.pending} pending` : '',
			counts.failed ? `${counts.failed} failed` : ''
		].filter(Boolean);

		return parts.length ? parts.join(' · ') : 'Queue is idle.';
	});
</script>

<svelte:head>
	<title>Admin | Shoka</title>
</svelte:head>

<div class="p-4 sm:p-6">
	<h1 class="text-2xl font-bold tracking-tight">Overview</h1>
	<p class="mt-1 text-sm text-muted-foreground">Manage your Shoka instance.</p>

	<div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		<Card.Root>
			<Card.Header>
				<Card.Title class="flex items-center gap-2">
					<Library class="size-4" />
					Libraries
				</Card.Title>
				<Card.Description>
					{data.libraries.length}
					{data.libraries.length === 1 ? 'library' : 'libraries'}
				</Card.Description>
			</Card.Header>
			<Card.Footer>
				<Button href="/admin/libraries" variant="outline" size="sm">Manage libraries</Button>
			</Card.Footer>
		</Card.Root>

		<Card.Root>
			<Card.Header>
				<Card.Title class="flex items-center gap-2">
					<CopyCheck class="size-4" />
					Duplicates
				</Card.Title>
				<Card.Description>
					Find archives that look alike across libraries by perceptual hash.
				</Card.Description>
			</Card.Header>
			<Card.Footer>
				<Button href="/admin/duplicates" variant="outline" size="sm">Review duplicates</Button>
			</Card.Footer>
		</Card.Root>

		{#if data.canReadJobs}
			<Card.Root>
				<Card.Header>
					<Card.Title class="flex items-center gap-2">
						<Activity class="size-4" />
						Jobs
					</Card.Title>
					<Card.Description>
						{#if data.jobCounts}
							{jobSummary}
						{:else}
							Queue status unavailable.
						{/if}
					</Card.Description>
				</Card.Header>
				<Card.Footer>
					<Button href="/admin/jobs" variant="outline" size="sm">View jobs</Button>
				</Card.Footer>
			</Card.Root>
		{/if}

		<Card.Root>
			<Card.Header>
				<Card.Title class="flex items-center gap-2">
					<Wrench class="size-4" />
					Maintenance
				</Card.Title>
				<Card.Description>Regenerate covers and other housekeeping tasks.</Card.Description>
			</Card.Header>
			<Card.Footer>
				<Button href="/admin/maintenance" variant="outline" size="sm">Open maintenance</Button>
			</Card.Footer>
		</Card.Root>

		{#if data.canListUsers}
			<Card.Root>
				<Card.Header>
					<Card.Title class="flex items-center gap-2">
						<UsersRound class="size-4" />
						Users
					</Card.Title>
					<Card.Description>
						{data.accounts}
						{data.accounts === 1 ? 'account' : 'accounts'}
					</Card.Description>
				</Card.Header>
				<Card.Footer>
					<Button href="/admin/users" variant="outline" size="sm">Manage users</Button>
				</Card.Footer>
			</Card.Root>
		{/if}
	</div>
</div>
