<script lang="ts">
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import * as TagsInput from '$lib/components/ui/tags-input/index.js';
	import { Combobox, MultiCombobox } from '$lib/components/ui/combobox/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import MetadataFetchDialog from '$lib/components/MetadataFetchDialog.svelte';
	import { Pencil } from '@lucide/svelte';
	import { enhance, applyAction } from '$app/forms';
	import type { Archive, ArchiveLanguage, Artist, Character, Parody, Tag } from '$lib/types';

	let {
		archive,
		libraryId,
		categories,
		languages,
		allArtists,
		allTags,
		allCharacters,
		allParodies
	}: {
		archive: Archive;
		libraryId: string;
		categories: string[];
		languages: ArchiveLanguage[];
		allArtists: Artist[];
		allTags: Tag[];
		allCharacters: Character[];
		allParodies: Parody[];
	} = $props();

	// Artists/tags/characters/parodies have a backend-provided "all" list to
	// search against (like ArchivesSidebar's filters), so they get the same
	// searchable MultiCombobox, just in `creatable` mode so a genuinely new
	// value can still be added. Circles has no such endpoint, so it stays
	// free-text TagsInput below.
	const artistOptions = $derived(allArtists.map((a) => ({ value: a.name, label: a.name })));
	const tagOptions = $derived(allTags.map((t) => ({ value: t.name, label: t.name })));
	const characterOptions = $derived(allCharacters.map((c) => ({ value: c.name, label: c.name })));
	const parodyOptions = $derived(allParodies.map((p) => ({ value: p.name, label: p.name })));

	// The API only ever gives us language back as a converted display name
	// ("English"), never the raw ISO code it's actually stored/written as. The
	// /api/archives/languages endpoint pairs up every code currently in use
	// with the same display name conversion, so - as long as the archive has
	// a language set at all - it's guaranteed to contain a matching entry we
	// can resolve back to the real code, rather than guessing.
	function resolveLanguageCode(name: string | null) {
		return languages.find((l) => l.name === name)?.code ?? '';
	}

	let open = $state(false);
	let saving = $state(false);
	let error = $state<string | null>(null);

	// svelte-ignore state_referenced_locally
	let title = $state(archive.title);
	// svelte-ignore state_referenced_locally
	let summary = $state(archive.summary ?? '');
	// svelte-ignore state_referenced_locally
	let category = $state(archive.category ?? '');
	// svelte-ignore state_referenced_locally
	let releaseDate = $state(archive.release_date?.slice(0, 10) ?? '');
	// svelte-ignore state_referenced_locally
	let languageValue = $state(resolveLanguageCode(archive.language));

	// archive.artists/tags/parodies/circles/characters are `null` (not `[]`)
	// whenever the archive has zero of that relation - the backend appends
	// onto a nil Go slice, so JSON-marshaling an empty one produces `null`.
	// TagsInput/MultiCombobox both require an actual array, so this can't be
	// skipped.
	// svelte-ignore state_referenced_locally
	let artists = $state(archive.artists ?? []);
	// svelte-ignore state_referenced_locally
	let tags = $state(archive.tags ?? []);
	// svelte-ignore state_referenced_locally
	let parodies = $state(archive.parodies ?? []);
	// svelte-ignore state_referenced_locally
	let circles = $state(archive.circles ?? []);
	// svelte-ignore state_referenced_locally
	let characters = $state(archive.characters ?? []);
	// archive.urls is ArchiveSourceLink[] | null on the detail response;
	// the edit form only works in raw URL strings. `?? []` for the same
	// null-not-[] reason as the relations above.
	// svelte-ignore state_referenced_locally
	let urls = $state(archive.urls?.map((u) => u.url) ?? []);

	// A metadata fetch can hand back a category/language not yet reflected
	// in `categories`/`languages` (those only cover values already saved
	// somewhere), so the currently-selected value is added in if it's
	// missing - otherwise the Combobox falls back to showing its placeholder
	// instead of the actual (still unsaved) value.
	const CATEGORY_OPTIONS = $derived.by(() => {
		const opts = [{ value: '', label: 'None' }, ...categories.map((c) => ({ value: c, label: c }))];
		if (category && !opts.some((o) => o.value === category)) {
			opts.push({ value: category, label: category });
		}
		return opts;
	});
	const LANGUAGE_OPTIONS = $derived.by(() => {
		const opts = [
			{ value: '', label: 'None' },
			...languages.map((l) => ({ value: l.code, label: l.name }))
		];
		if (languageValue && !opts.some((o) => o.value === languageValue)) {
			opts.push({ value: languageValue, label: languageValue });
		}
		return opts;
	});

	function applyFetchedMetadata(fields: {
		title?: string;
		summary?: string;
		category?: string;
		language?: string;
		releaseDate?: string;
		artists?: string[];
		tags?: string[];
		parodies?: string[];
		circles?: string[];
		characters?: string[];
		urls?: string[];
	}) {
		if (fields.title !== undefined) title = fields.title;
		if (fields.summary !== undefined) summary = fields.summary;
		if (fields.category !== undefined) category = fields.category;
		if (fields.language !== undefined) languageValue = fields.language;
		if (fields.releaseDate !== undefined) releaseDate = fields.releaseDate;
		if (fields.artists !== undefined) artists = fields.artists;
		if (fields.tags !== undefined) tags = fields.tags;
		if (fields.parodies !== undefined) parodies = fields.parodies;
		if (fields.circles !== undefined) circles = fields.circles;
		if (fields.characters !== undefined) characters = fields.characters;
		if (fields.urls !== undefined) urls = fields.urls;
	}
</script>

<Sheet.Root bind:open>
	<Sheet.Trigger>
		{#snippet child({ props })}
			<Button {...props} variant="outline" size="sm">
				<Pencil />
				Edit
			</Button>
		{/snippet}
	</Sheet.Trigger>
	<Sheet.Content class="w-full overflow-y-auto sm:max-w-md">
		<Sheet.Header>
			<Sheet.Title>Edit archive</Sheet.Title>
			<Sheet.Description
				>Manual edits aren't protected from a later metadata fetch overwriting them.</Sheet.Description
			>
		</Sheet.Header>

		<div class="px-4">
			<MetadataFetchDialog
				archiveId={archive.id}
				{libraryId}
				current={{
					title,
					summary,
					category,
					language: languageValue,
					releaseDate,
					artists,
					tags,
					parodies,
					circles,
					characters,
					urls
				}}
				{languages}
				onApply={applyFetchedMetadata}
			/>
		</div>

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
						error = (result.data?.error as string | undefined) ?? 'Failed to update archive.';
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
					<Field.Label for="title">Title</Field.Label>
					<Input id="title" name="title" bind:value={title} required />
				</Field.Field>

				<Field.Field>
					<Field.Label for="summary">Summary</Field.Label>
					<Textarea id="summary" name="summary" bind:value={summary} class="min-h-24" />
				</Field.Field>

				<Field.Field>
					<Field.Label for="category">Category</Field.Label>
					<Combobox
						options={CATEGORY_OPTIONS}
						value={category}
						onValueChange={(v) => (category = v)}
						placeholder="None"
						searchPlaceholder="Search category..."
					/>
					<input type="hidden" name="category" value={category} />
				</Field.Field>

				<Field.Field>
					<Field.Label for="language">Language</Field.Label>
					<Combobox
						options={LANGUAGE_OPTIONS}
						value={languageValue}
						onValueChange={(v) => (languageValue = v)}
						placeholder="None"
						searchPlaceholder="Search language..."
					/>
					<input type="hidden" name="language" value={languageValue} />
				</Field.Field>

				<Field.Field>
					<Field.Label for="release_date">Release date</Field.Label>
					<Input id="release_date" name="release_date" type="date" bind:value={releaseDate} />
				</Field.Field>

				<Field.Field>
					<Field.Label>Artists</Field.Label>
					<MultiCombobox
						options={artistOptions}
						value={artists}
						onValueChange={(v) => (artists = v)}
						placeholder="Add artist..."
						searchPlaceholder="Search artists..."
						emptyText="No artists found."
						creatable
					/>
					{#each artists as artist (artist)}
						<input type="hidden" name="artists" value={artist} />
					{/each}
				</Field.Field>

				<Field.Field>
					<Field.Label>Circles</Field.Label>
					<TagsInput.Root value={circles} onValueChange={(v) => (circles = v)} addOnPaste>
						{#each circles as circle (circle)}
							<TagsInput.Item value={circle}>
								<TagsInput.ItemText />
								<TagsInput.ItemDelete />
							</TagsInput.Item>
						{/each}
						<TagsInput.Input placeholder="Add circle..." />
					</TagsInput.Root>
					{#each circles as circle (circle)}
						<input type="hidden" name="circles" value={circle} />
					{/each}
				</Field.Field>

				<Field.Field>
					<Field.Label>Parodies</Field.Label>
					<MultiCombobox
						options={parodyOptions}
						value={parodies}
						onValueChange={(v) => (parodies = v)}
						placeholder="Add parody..."
						searchPlaceholder="Search parodies..."
						emptyText="No parodies found."
						creatable
					/>
					{#each parodies as parody (parody)}
						<input type="hidden" name="parodies" value={parody} />
					{/each}
				</Field.Field>

				<Field.Field>
					<Field.Label>Characters</Field.Label>
					<MultiCombobox
						options={characterOptions}
						value={characters}
						onValueChange={(v) => (characters = v)}
						placeholder="Add character..."
						searchPlaceholder="Search characters..."
						emptyText="No characters found."
						creatable
					/>
					{#each characters as character (character)}
						<input type="hidden" name="characters" value={character} />
					{/each}
				</Field.Field>

				<Field.Field>
					<Field.Label>Tags</Field.Label>
					<MultiCombobox
						options={tagOptions}
						value={tags}
						onValueChange={(v) => (tags = v)}
						placeholder="Add tag..."
						searchPlaceholder="Search tags..."
						emptyText="No tags found."
						creatable
					/>
					{#each tags as tag (tag)}
						<input type="hidden" name="tags" value={tag} />
					{/each}
				</Field.Field>

				<Field.Field>
					<Field.Label>Source links</Field.Label>
					<TagsInput.Root value={urls} onValueChange={(v) => (urls = v)} addOnPaste>
						{#each urls as url (url)}
							<TagsInput.Item value={url} class="max-w-full min-w-0">
								<TagsInput.ItemText class="truncate" />
								<TagsInput.ItemDelete />
							</TagsInput.Item>
						{/each}
						<TagsInput.Input placeholder="Add source URL..." />
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
