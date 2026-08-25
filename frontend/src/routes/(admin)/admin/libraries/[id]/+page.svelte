<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import LibrarySourceCard from '$lib/components/library-source-card.svelte';
	import { enhance } from '$app/forms';
	import { ArrowLeft } from '@lucide/svelte';

	let { data, form } = $props();

	let renaming = $state(false);

	const typeLabel = $derived(
		data.library.type.charAt(0).toUpperCase() + data.library.type.slice(1)
	);
</script>

<svelte:head>
	<title>{data.library.name} | Shoka Admin</title>
</svelte:head>

<div class="mx-auto max-w-3xl p-4 sm:p-6">
	<a
		href="/admin/libraries"
		class="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
	>
		<ArrowLeft class="size-3.5" />
		Libraries
	</a>

	<div class="mt-3 flex items-center gap-2">
		<h1 class="text-2xl font-bold tracking-tight">{data.library.name}</h1>
		<Badge variant="outline">{typeLabel}</Badge>
	</div>
	<p class="mt-1 truncate font-mono text-sm text-muted-foreground">{data.library.path}</p>

	<Card.Root class="mt-6">
		<form
			class="contents"
			method="POST"
			action="?/rename"
			use:enhance={() => {
				renaming = true;
				return async ({ update }) => {
					renaming = false;
					await update();
				};
			}}
		>
			<Card.Header>
				<Card.Title>General</Card.Title>
				<Card.Description>
					The path and content type are fixed after a library is created.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<Field.Group>
					<Field.Field>
						<Field.Label for="name">Name</Field.Label>
						<Input id="name" name="name" value={data.library.name} required />
					</Field.Field>
					{#if form?.action === 'rename' && form.error}
						<p class="text-sm text-destructive">{form.error}</p>
					{/if}
				</Field.Group>
			</Card.Content>
			<Card.Footer>
				<Button type="submit" size="sm" disabled={renaming}>
					{renaming ? 'Saving…' : 'Save'}
				</Button>
			</Card.Footer>
		</form>
	</Card.Root>

	<div class="mt-8">
		<h2 class="text-lg font-semibold">Metadata sources</h2>
		<p class="mt-1 text-sm text-muted-foreground">
			Shoka tries these in priority order when fetching metadata for an archive in this library.
		</p>

		<div class="mt-4 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
			{#each data.sources as source (source.source)}
				<LibrarySourceCard
					{source}
					error={form?.action === 'updateSource' && form.source === source.source
						? form.error
						: undefined}
				/>
			{/each}
		</div>
	</div>
</div>
