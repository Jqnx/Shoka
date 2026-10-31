<script lang="ts">
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import * as TagsInput from '$lib/components/ui/tags-input/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Pencil } from '@lucide/svelte';
	import { enhance, applyAction } from '$app/forms';
	import type { ArtistDetail } from '$lib/types';

	let { artist }: { artist: ArtistDetail } = $props();

	let open = $state(false);
	let saving = $state(false);
	let error = $state<string | null>(null);

	// The detail response serializes empty aliases/urls as `null`, and
	// TagsInput requires an actual array.
	// svelte-ignore state_referenced_locally
	let name = $state(artist.name);
	// svelte-ignore state_referenced_locally
	let aliases = $state(artist.aliases ?? []);
	// svelte-ignore state_referenced_locally
	let urls = $state(artist.urls ?? []);
</script>

<Sheet.Root bind:open>
	<Sheet.Trigger>
		{#snippet child({ props })}
			<Button {...props} variant="outline" size="icon" aria-label="Edit">
				<Pencil />
			</Button>
		{/snippet}
	</Sheet.Trigger>
	<Sheet.Content class="w-full overflow-y-auto sm:max-w-md">
		<Sheet.Header>
			<Sheet.Title>Edit artist</Sheet.Title>
			<Sheet.Description>Aliases and links apply immediately; there is no undo.</Sheet.Description>
		</Sheet.Header>

		<form
			method="POST"
			action="?/update"
			class="flex flex-1 flex-col overflow-y-auto"
			use:enhance={() => {
				saving = true;
				error = null;
				return async ({ result, update }) => {
					saving = false;
					if (result.type === 'failure') {
						error = (result.data?.error as string | undefined) ?? 'Failed to update artist.';
						return;
					}
					if (result.type === 'success') {
						await update();
						open = false;
						return;
					}
					await applyAction(result);
				};
			}}
		>
			<Field.Group class="flex-1 gap-4 overflow-y-auto px-4">
				<Field.Field>
					<Field.Label for="name">Name</Field.Label>
					<Input id="name" name="name" bind:value={name} required />
				</Field.Field>

				<Field.Field>
					<Field.Label>Aliases</Field.Label>
					<TagsInput.Root value={aliases} onValueChange={(v) => (aliases = v)} addOnPaste>
						{#each aliases as alias (alias)}
							<TagsInput.Item value={alias} class="max-w-full min-w-0">
								<TagsInput.ItemText class="truncate" />
								<TagsInput.ItemDelete />
							</TagsInput.Item>
						{/each}
						<TagsInput.Input placeholder="Add alias…" />
					</TagsInput.Root>
					{#each aliases as alias (alias)}
						<input type="hidden" name="aliases" value={alias} />
					{/each}
				</Field.Field>

				<Field.Field>
					<Field.Label>Links</Field.Label>
					<TagsInput.Root value={urls} onValueChange={(v) => (urls = v)} addOnPaste>
						{#each urls as url (url)}
							<TagsInput.Item value={url} class="max-w-full min-w-0">
								<TagsInput.ItemText class="truncate" />
								<TagsInput.ItemDelete />
							</TagsInput.Item>
						{/each}
						<TagsInput.Input placeholder="Add URL…" />
					</TagsInput.Root>
					{#each urls as url (url)}
						<input type="hidden" name="urls" value={url} />
					{/each}
				</Field.Field>

				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}
			</Field.Group>

			<Sheet.Footer>
				<Button type="submit" disabled={saving}>
					{saving ? 'Saving…' : 'Save changes'}
				</Button>
			</Sheet.Footer>
		</form>
	</Sheet.Content>
</Sheet.Root>
