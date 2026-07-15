<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { enhance } from '$app/forms';
	import { ImageUp } from '@lucide/svelte';

	let { form } = $props();

	let regenerating = $state(false);
</script>

<svelte:head>
	<title>Maintenance | Shoka Admin</title>
</svelte:head>

<div class="p-4 sm:p-6">
	<h1 class="text-2xl font-bold tracking-tight">Maintenance</h1>
	<p class="mt-1 text-sm text-muted-foreground">Housekeeping tasks for your Shoka instance.</p>

	<div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		<Card.Root>
			<Card.Header>
				<Card.Title class="flex items-center gap-2">
					<ImageUp class="size-4" />
					Regenerate covers
				</Card.Title>
				<Card.Description>
					Re-enqueue cover thumbnail generation for every archive in every library.
				</Card.Description>
			</Card.Header>
			<Card.Footer class="flex-col items-start gap-2">
				<form
					method="POST"
					action="?/regenerateCovers"
					use:enhance={() => {
						regenerating = true;
						return async ({ update }) => {
							await update();
							regenerating = false;
						};
					}}
				>
					<Button type="submit" variant="outline" size="sm" disabled={regenerating}>
						{regenerating ? 'Enqueuing…' : 'Regenerate covers'}
					</Button>
				</form>
				{#if form?.success}
					<p class="text-sm text-muted-foreground">Cover generation enqueued for all archives.</p>
				{:else if form?.error}
					<p class="text-sm text-destructive">{form.error}</p>
				{/if}
			</Card.Footer>
		</Card.Root>
	</div>
</div>
