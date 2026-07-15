<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { enhance, applyAction } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import { Plus } from '@lucide/svelte';
	import type { Snippet } from 'svelte';
	import type { LibraryType } from '$lib/types';

	let {
		types,
		trigger
	}: {
		/** Every library type the backend's schema knows about, supported or not (from GET /api/libraries/types). */
		types: LibraryType[];
		/** Custom trigger content; receives the props to spread onto it. Defaults to a plain button. */
		trigger?: Snippet<[{ props: Record<string, unknown> }]>;
	} = $props();

	function typeLabel(type: string) {
		return type.charAt(0).toUpperCase() + type.slice(1);
	}

	const DEFAULT_TYPE = 'doujinshi';

	function defaultType() {
		return types.find((t) => t.type === DEFAULT_TYPE)?.type ?? (types[0]?.type ?? DEFAULT_TYPE);
	}

	let open = $state(false);
	let creating = $state(false);
	let error = $state<string | null>(null);
	let libraryType = $state(defaultType());
	const libraryTypeLabel = $derived(typeLabel(libraryType));
</script>

<Dialog.Root bind:open>
	<Dialog.Trigger>
		{#snippet child({ props })}
			{#if trigger}
				{@render trigger({ props })}
			{:else}
				<Button {...props} size="sm">
					<Plus />
					Add library
				</Button>
			{/if}
		{/snippet}
	</Dialog.Trigger>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Add library</Dialog.Title>
			<Dialog.Description>
				Shoka will scan this folder for archives and keep watching it for changes.
			</Dialog.Description>
		</Dialog.Header>

		<form
			method="POST"
			action="/admin/libraries?/create"
			use:enhance={() => {
				creating = true;
				error = null;
				return async ({ result }) => {
					creating = false;
					if (result.type === 'failure') {
						error = (result.data?.error as string | undefined) ?? 'Failed to create library.';
						return;
					}
					if (result.type === 'success') {
						open = false;
						libraryType = defaultType();
						await invalidateAll();
						return;
					}
					await applyAction(result);
				};
			}}
		>
			<Field.Group>
				<Field.Field>
					<Field.Label for="add-library-name">Name</Field.Label>
					<Input id="add-library-name" name="name" placeholder="My Library" required />
				</Field.Field>
				<Field.Field>
					<Field.Label for="add-library-path">Path</Field.Label>
					<Input id="add-library-path" name="path" placeholder="/data/library" required />
				</Field.Field>
				<Field.Field>
					<Field.Label for="add-library-type">Content type</Field.Label>
					<Select.Root type="single" name="type" bind:value={libraryType}>
						<Select.Trigger id="add-library-type" class="w-full">
							{libraryTypeLabel}
						</Select.Trigger>
						<Select.Content>
							{#each types as type (type.type)}
								<Select.Item
									value={type.type}
									label={typeLabel(type.type)}
									disabled={!type.supported}
								>
									{typeLabel(type.type)}{!type.supported ? ' (coming soon)' : ''}
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</Field.Field>
				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}
			</Field.Group>

			<Dialog.Footer showCloseButton class="mt-4">
				<Button type="submit" disabled={creating}>
					{creating ? 'Adding…' : 'Add library'}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
