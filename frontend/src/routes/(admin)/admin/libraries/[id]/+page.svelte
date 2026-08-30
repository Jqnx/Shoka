<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import LibrarySourceCard from '$lib/components/library-source-card.svelte';
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import { ArrowLeft } from '@lucide/svelte';

	let { data, form } = $props();

	let saving = $state(false);

	// The backend rejects a non-zero interval below 15 minutes, so the
	// presets start there.
	const scanIntervals = [
		{ value: '0', label: 'Off' },
		{ value: '15', label: 'Every 15 minutes' },
		{ value: '60', label: 'Every hour' },
		{ value: '360', label: 'Every 6 hours' },
		{ value: '720', label: 'Every 12 hours' },
		{ value: '1440', label: 'Every 24 hours' }
	];

	// Uncontrolled-after-mount: seeded from the loaded library, then edited
	// locally until the form is submitted (same pattern as
	// library-source-card.svelte).
	// svelte-ignore state_referenced_locally
	let scanInterval = $state(String(data.library.scan_interval_minutes));
	// svelte-ignore state_referenced_locally
	let watchEnabled = $state(data.library.watch_enabled);

	const typeLabel = $derived(
		data.library.type.charAt(0).toUpperCase() + data.library.type.slice(1)
	);

	const scanIntervalLabel = $derived(
		scanIntervals.find((i) => i.value === scanInterval)?.label ?? 'Off'
	);

	const lastScanned = $derived(
		data.library.last_scanned_at
			? new Date(data.library.last_scanned_at).toLocaleString(undefined, {
					dateStyle: 'medium',
					timeStyle: 'short'
				})
			: null
	);
</script>

<svelte:head>
	<title>{data.library.name} | Shoka Admin</title>
</svelte:head>

<div class="mx-auto max-w-3xl p-4 sm:p-6">
	<a
		href={resolve('/admin/libraries')}
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
			action="?/updateGeneral"
			use:enhance={() => {
				saving = true;
				return async ({ update }) => {
					saving = false;
					await update();
				};
			}}
		>
			<input type="hidden" name="watch_enabled" value={String(watchEnabled)} />

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

					<Field.Field>
						<Field.Label for="scan-interval">Automatic scanning</Field.Label>
						<Select.Root type="single" name="scan_interval_minutes" bind:value={scanInterval}>
							<Select.Trigger id="scan-interval" class="w-full">
								{scanIntervalLabel}
							</Select.Trigger>
							<Select.Content>
								{#each scanIntervals as interval (interval.value)}
									<Select.Item value={interval.value} label={interval.label}>
										{interval.label}
									</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
						<Field.Description>
							Rescans this library on a timer, measured from the last scan.
							{#if lastScanned}
								Last scanned {lastScanned}.
							{:else}
								Never scanned yet.
							{/if}
						</Field.Description>
					</Field.Field>

					<Field.Field orientation="horizontal">
						<Field.Content>
							<Field.Label for="watch-enabled">Watch for file changes</Field.Label>
							<Field.Description>
								Picks up new files as soon as they appear. Filesystem events aren't delivered over
								network shares or some container mounts — turn this off and rely on automatic
								scanning if it isn't working.
							</Field.Description>
						</Field.Content>
						<Switch id="watch-enabled" bind:checked={watchEnabled} />
					</Field.Field>

					{#if form?.action === 'updateGeneral' && form.error}
						<p class="text-sm text-destructive">{form.error}</p>
					{/if}
				</Field.Group>
			</Card.Content>
			<Card.Footer>
				<Button type="submit" size="sm" disabled={saving}>
					{saving ? 'Saving…' : 'Save'}
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
