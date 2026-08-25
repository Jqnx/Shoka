<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { CopyCheck, Library, Wrench } from '@lucide/svelte';

	let { data } = $props();
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
	</div>
</div>
