<script lang="ts">
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import { Combobox, MultiCombobox } from '$lib/components/ui/combobox/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Pencil } from '@lucide/svelte';
	import { enhance, applyAction } from '$app/forms';
	import type { ArchiveLanguage, Artist, Audio, Circle, Tag, VoiceActor } from '$lib/types';

	let {
		audio,
		languages,
		allArtists,
		allTags,
		allCircles,
		allVoiceActors
	}: {
		audio: Audio;
		languages: ArchiveLanguage[];
		allArtists: Artist[];
		allTags: Tag[];
		allCircles: Circle[];
		allVoiceActors: VoiceActor[];
	} = $props();

	let open = $state(false);
	let saving = $state(false);
	let error = $state<string | null>(null);

	// svelte-ignore state_referenced_locally
	let title = $state(audio.title);
	// release_date arrives as an RFC3339 timestamp; an <input type="date">
	// only accepts the date part.
	// svelte-ignore state_referenced_locally
	let releaseDate = $state(audio.release_date?.slice(0, 10) ?? '');
	// svelte-ignore state_referenced_locally
	let language = $state(audio.language ?? '');
	// svelte-ignore state_referenced_locally
	let circle = $state(audio.circle ?? '');
	// svelte-ignore state_referenced_locally
	let artists = $state(audio.artists ?? []);
	// svelte-ignore state_referenced_locally
	let tags = $state(audio.tags ?? []);
	// svelte-ignore state_referenced_locally
	let voiceActors = $state(audio.voice_actors ?? []);

	const languageOptions = $derived.by(() => {
		const options = [
			{ value: '', label: 'None' },
			...languages.map((l) => ({ value: l.code, label: l.name }))
		];
		if (language && !options.some((o) => o.value === language)) {
			options.push({ value: language, label: language });
		}
		return options;
	});

	const circleOptions = $derived.by(() => {
		const options = [
			{ value: '', label: 'None' },
			...allCircles.map((c) => ({ value: c.name, label: c.name }))
		];
		if (circle && !options.some((o) => o.value === circle)) {
			options.push({ value: circle, label: circle });
		}
		return options;
	});

	const artistOptions = $derived(allArtists.map((a) => ({ value: a.name, label: a.name })));
	const tagOptions = $derived(allTags.map((t) => ({ value: t.name, label: t.name })));
	const voiceActorOptions = $derived(allVoiceActors.map((v) => ({ value: v.name, label: v.name })));
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
			<Sheet.Title>Edit audio</Sheet.Title>
			<Sheet.Description
				>Saving replaces every field and relation; there is no undo.</Sheet.Description
			>
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
						error = (result.data?.error as string | undefined) ?? 'Failed to update audio.';
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
					<Field.Label for="release_date">Release date</Field.Label>
					<Input id="release_date" name="release_date" type="date" bind:value={releaseDate} />
				</Field.Field>

				<Field.Field>
					<Field.Label>Language</Field.Label>
					<Combobox
						options={languageOptions}
						value={language}
						onValueChange={(v) => (language = v)}
						placeholder="None"
					/>
				</Field.Field>

				<Field.Field>
					<Field.Label>Circle</Field.Label>
					<Combobox
						options={circleOptions}
						value={circle}
						onValueChange={(v) => (circle = v)}
						placeholder="None"
					/>
				</Field.Field>

				<Field.Field>
					<Field.Label>Artists</Field.Label>
					<MultiCombobox
						options={artistOptions}
						value={artists}
						onValueChange={(v) => (artists = v)}
						creatable
						placeholder="Add artist…"
					/>
					{#each artists as artist (artist)}
						<input type="hidden" name="artists" value={artist} />
					{/each}
				</Field.Field>

				<Field.Field>
					<Field.Label>Tags</Field.Label>
					<MultiCombobox
						options={tagOptions}
						value={tags}
						onValueChange={(v) => (tags = v)}
						creatable
						placeholder="Add tag…"
					/>
					{#each tags as tag (tag)}
						<input type="hidden" name="tags" value={tag} />
					{/each}
				</Field.Field>

				<Field.Field>
					<Field.Label>Voice actors</Field.Label>
					<MultiCombobox
						options={voiceActorOptions}
						value={voiceActors}
						onValueChange={(v) => (voiceActors = v)}
						creatable
						placeholder="Add voice actor…"
					/>
					{#each voiceActors as voiceActor (voiceActor)}
						<input type="hidden" name="voice_actors" value={voiceActor} />
					{/each}
				</Field.Field>

				{#if error}
					<p class="text-sm text-destructive">{error}</p>
				{/if}
			</Field.Group>

			<input type="hidden" name="language" value={language} />
			<input type="hidden" name="circle" value={circle} />

			<Sheet.Footer>
				<Button type="submit" disabled={saving}>
					{saving ? 'Saving…' : 'Save changes'}
				</Button>
			</Sheet.Footer>
		</form>
	</Sheet.Content>
</Sheet.Root>
