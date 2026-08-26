<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import * as TagsInput from '$lib/components/ui/tags-input/index.js';
	import { Combobox, MultiCombobox } from '$lib/components/ui/combobox/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Pencil } from '@lucide/svelte';
	import { enhance, applyAction } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import type { ArchiveLanguage, Artist, Character, Parody, Tag } from '$lib/types';

	let {
		ids,
		categories,
		languages,
		allArtists,
		allTags,
		allCharacters,
		allParodies,
		onDone
	}: {
		ids: string[];
		categories: string[];
		languages: ArchiveLanguage[];
		allArtists: Artist[];
		allTags: Tag[];
		allCharacters: Character[];
		allParodies: Parody[];
		onDone?: () => void;
	} = $props();

	const KEEP = '__keep__';

	const artistOptions = $derived(allArtists.map((a) => ({ value: a.name, label: a.name })));
	const tagOptions = $derived(allTags.map((t) => ({ value: t.name, label: t.name })));
	const characterOptions = $derived(allCharacters.map((c) => ({ value: c.name, label: c.name })));
	const parodyOptions = $derived(allParodies.map((p) => ({ value: p.name, label: p.name })));

	// A bulk edit has no single current value to prefill, so scalars get an
	// explicit "leave unchanged" option instead of a blank meaning two
	// different things.
	const categoryOptions = $derived([
		{ value: KEEP, label: 'Leave unchanged' },
		...categories.map((c) => ({ value: c, label: c }))
	]);
	const languageOptions = $derived([
		{ value: KEEP, label: 'Leave unchanged' },
		...languages.map((l) => ({ value: l.code, label: l.name }))
	]);

	let open = $state(false);
	let saving = $state(false);
	let error = $state<string | null>(null);

	let artists = $state<string[]>([]);
	let tags = $state<string[]>([]);
	let parodies = $state<string[]>([]);
	let characters = $state<string[]>([]);
	let circles = $state<string[]>([]);
	let category = $state(KEEP);
	let language = $state(KEEP);

	const hasChanges = $derived(
		artists.length > 0 ||
			tags.length > 0 ||
			parodies.length > 0 ||
			characters.length > 0 ||
			circles.length > 0 ||
			category !== KEEP ||
			language !== KEEP
	);

	function reset() {
		artists = [];
		tags = [];
		parodies = [];
		characters = [];
		circles = [];
		category = KEEP;
		language = KEEP;
		error = null;
	}
</script>

<Dialog.Root
	bind:open
	onOpenChange={(v) => {
		if (v) reset();
	}}
>
	<Dialog.Trigger>
		{#snippet child({ props })}
			<Button {...props} variant="outline" size="sm" class="gap-1.5">
				<Pencil class="size-3.5" />
				Edit
			</Button>
		{/snippet}
	</Dialog.Trigger>
	<Dialog.Content class="sm:max-w-lg">
		<Dialog.Header>
			<Dialog.Title>Edit {ids.length} {ids.length === 1 ? 'archive' : 'archives'}</Dialog.Title>
			<Dialog.Description>
				Artists, tags, parodies, circles and characters are added to what each archive already
				has. Category and language overwrite it.
			</Dialog.Description>
		</Dialog.Header>

		<form
			method="POST"
			action="?/bulkEdit"
			use:enhance={() => {
				saving = true;
				error = null;
				return async ({ result }) => {
					saving = false;
					if (result.type === 'failure') {
						error = (result.data?.error as string | undefined) ?? 'Failed to update archives.';
						return;
					}
					if (result.type === 'success') {
						open = false;
						// applyAction publishes the action's result to the page's
						// `form` prop, which is what renders the "N of M updated"
						// summary - without it the edit succeeds silently, unlike
						// the other bulk actions.
						await applyAction(result);
						await invalidateAll();
						onDone?.();
					}
				};
			}}
		>
			{#each ids as id (id)}
				<input type="hidden" name="ids" value={id} />
			{/each}
			{#each artists as v (v)}<input type="hidden" name="artists" value={v} />{/each}
			{#each tags as v (v)}<input type="hidden" name="tags" value={v} />{/each}
			{#each parodies as v (v)}<input type="hidden" name="parodies" value={v} />{/each}
			{#each characters as v (v)}<input type="hidden" name="characters" value={v} />{/each}
			{#each circles as v (v)}<input type="hidden" name="circles" value={v} />{/each}
			{#if category !== KEEP}<input type="hidden" name="category" value={category} />{/if}
			{#if language !== KEEP}<input type="hidden" name="language" value={language} />{/if}

			<div class="max-h-[60vh] overflow-y-auto pr-1">
				<Field.Group>
					<Field.Field>
						<Field.Label>Add artists</Field.Label>
						<MultiCombobox
							options={artistOptions}
							value={artists}
							onValueChange={(v) => (artists = v)}
							creatable
							placeholder="Add artist..."
							searchPlaceholder="Search artists..."
						/>
					</Field.Field>
					<Field.Field>
						<Field.Label>Add tags</Field.Label>
						<MultiCombobox
							options={tagOptions}
							value={tags}
							onValueChange={(v) => (tags = v)}
							creatable
							placeholder="Add tag..."
							searchPlaceholder="Search tags..."
						/>
					</Field.Field>
					<Field.Field>
						<Field.Label>Add parodies</Field.Label>
						<MultiCombobox
							options={parodyOptions}
							value={parodies}
							onValueChange={(v) => (parodies = v)}
							creatable
							placeholder="Add parody..."
							searchPlaceholder="Search parodies..."
						/>
					</Field.Field>
					<Field.Field>
						<Field.Label>Add characters</Field.Label>
						<MultiCombobox
							options={characterOptions}
							value={characters}
							onValueChange={(v) => (characters = v)}
							creatable
							placeholder="Add character..."
							searchPlaceholder="Search characters..."
						/>
					</Field.Field>
					<Field.Field>
						<Field.Label>Add circles</Field.Label>
						<TagsInput.Root value={circles} onValueChange={(v) => (circles = v)} addOnPaste>
							{#each circles as circle (circle)}
								<TagsInput.Item value={circle}>
									<TagsInput.ItemText />
									<TagsInput.ItemDelete />
								</TagsInput.Item>
							{/each}
							<TagsInput.Input placeholder="Add circle..." />
						</TagsInput.Root>
					</Field.Field>
					<Field.Field>
						<Field.Label>Set category</Field.Label>
						<Combobox
							options={categoryOptions}
							value={category}
							onValueChange={(v) => (category = v)}
							placeholder="Leave unchanged"
							searchPlaceholder="Search category..."
						/>
					</Field.Field>
					<Field.Field>
						<Field.Label>Set language</Field.Label>
						<Combobox
							options={languageOptions}
							value={language}
							onValueChange={(v) => (language = v)}
							placeholder="Leave unchanged"
							searchPlaceholder="Search language..."
						/>
					</Field.Field>
					{#if error}
						<p class="text-sm text-destructive">{error}</p>
					{/if}
				</Field.Group>
			</div>

			<Dialog.Footer showCloseButton class="mt-4">
				<Button type="submit" disabled={saving || !hasChanges}>
					{saving ? 'Saving…' : 'Apply to ' + ids.length}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
