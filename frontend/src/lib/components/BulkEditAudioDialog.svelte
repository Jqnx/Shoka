<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import { Combobox, MultiCombobox } from '$lib/components/ui/combobox/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Pencil } from '@lucide/svelte';
	import { enhance, applyAction } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import type { ArchiveLanguage, Artist, Circle, Tag, VoiceActor } from '$lib/types';

	let {
		ids,
		languages,
		allArtists,
		allTags,
		allCircles,
		allVoiceActors,
		onDone
	}: {
		ids: string[];
		languages: ArchiveLanguage[];
		allArtists: Artist[];
		allTags: Tag[];
		allCircles: Circle[];
		allVoiceActors: VoiceActor[];
		onDone?: () => void;
	} = $props();

	const KEEP = '__keep__';

	const artistOptions = $derived(allArtists.map((a) => ({ value: a.name, label: a.name })));
	const tagOptions = $derived(allTags.map((t) => ({ value: t.name, label: t.name })));
	const voiceActorOptions = $derived(allVoiceActors.map((v) => ({ value: v.name, label: v.name })));

	// A bulk edit has no single current value to prefill, so scalars get an
	// explicit "leave unchanged" option; "None" is how you clear one. The
	// language table can contain an empty code, which would duplicate the None
	// key, so drop blanks from the value list.
	const languageOptions = $derived([
		{ value: KEEP, label: 'Leave unchanged' },
		{ value: '', label: 'None' },
		...languages.filter((l) => l.code !== '').map((l) => ({ value: l.code, label: l.name }))
	]);
	const circleOptions = $derived([
		{ value: KEEP, label: 'Leave unchanged' },
		{ value: '', label: 'None' },
		...allCircles.filter((c) => c.name !== '').map((c) => ({ value: c.name, label: c.name }))
	]);

	let open = $state(false);
	let saving = $state(false);
	let error = $state<string | null>(null);

	let artists = $state<string[]>([]);
	let tags = $state<string[]>([]);
	let voiceActors = $state<string[]>([]);
	let language = $state(KEEP);
	let circle = $state(KEEP);

	const hasChanges = $derived(
		artists.length > 0 ||
			tags.length > 0 ||
			voiceActors.length > 0 ||
			language !== KEEP ||
			circle !== KEEP
	);

	function reset() {
		artists = [];
		tags = [];
		voiceActors = [];
		language = KEEP;
		circle = KEEP;
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
			<Dialog.Title>Edit {ids.length} {ids.length === 1 ? 'work' : 'works'}</Dialog.Title>
			<Dialog.Description>
				Artists, tags and voice actors are added to what each work already has. Language and circle
				overwrite it.
			</Dialog.Description>
		</Dialog.Header>

		<form
			method="POST"
			action="?/bulkAudioEdit"
			use:enhance={() => {
				saving = true;
				error = null;
				return async ({ result }) => {
					saving = false;
					if (result.type === 'failure') {
						error = (result.data?.error as string | undefined) ?? 'Failed to update audio works.';
						return;
					}
					if (result.type === 'success') {
						open = false;
						// applyAction publishes the action's result to the page's
						// `form` prop, which is what renders the "N of M updated"
						// summary.
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
			{#each voiceActors as v (v)}<input type="hidden" name="voice_actors" value={v} />{/each}
			{#if language !== KEEP}<input type="hidden" name="language" value={language} />{/if}
			{#if circle !== KEEP}<input type="hidden" name="circle" value={circle} />{/if}

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
						<Field.Label>Add voice actors</Field.Label>
						<MultiCombobox
							options={voiceActorOptions}
							value={voiceActors}
							onValueChange={(v) => (voiceActors = v)}
							creatable
							placeholder="Add voice actor..."
							searchPlaceholder="Search voice actors..."
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
					<Field.Field>
						<Field.Label>Set circle</Field.Label>
						<Combobox
							options={circleOptions}
							value={circle}
							onValueChange={(v) => (circle = v)}
							placeholder="Leave unchanged"
							searchPlaceholder="Search circles..."
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
