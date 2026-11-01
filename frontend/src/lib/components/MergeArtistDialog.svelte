<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Combobox } from '$lib/components/ui/combobox/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { enhance, applyAction } from '$app/forms';
	import type { Artist, ArtistDetail } from '$lib/types';

	let {
		open = $bindable(false),
		artist,
		artists
	}: { open?: boolean; artist: ArtistDetail; artists: Artist[] } = $props();

	let targetID = $state('');
	let saving = $state(false);
	let error = $state<string | null>(null);

	const options = $derived(
		artists
			.filter((candidate) => candidate.id !== artist.id)
			.map((candidate) => ({ value: String(candidate.id), label: candidate.name }))
	);
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-lg">
		<Dialog.Header>
			<Dialog.Title>Merge "{artist.name}" into…</Dialog.Title>
			<Dialog.Description>
				Its archives move to the chosen artist, and its name and aliases become aliases there. This
				artist is then deleted.
			</Dialog.Description>
		</Dialog.Header>

		<form
			method="POST"
			action="?/mergeArtist"
			use:enhance={() => {
				saving = true;
				return async ({ result }) => {
					saving = false;
					if (result.type === 'failure') {
						error = (result.data?.error as string | undefined) ?? 'Failed to merge artists.';
						return;
					}
					// The post-merge redirect reuses this page component, so the
					// dialog would otherwise stay open on the target artist.
					open = false;
					await applyAction(result);
				};
			}}
		>
			<input type="hidden" name="target_id" value={targetID} />
			<Combobox
				{options}
				value={targetID}
				onValueChange={(v) => {
					targetID = v;
					error = null;
				}}
				placeholder="Search artists…"
				searchPlaceholder="Search artists…"
			/>

			{#if error}
				<p class="mt-3 text-sm text-destructive">{error}</p>
			{/if}

			<Dialog.Footer class="mt-4">
				<Button type="button" variant="outline" onclick={() => (open = false)}>Cancel</Button>
				<Button type="submit" variant="destructive" disabled={!targetID || saving}>
					{saving ? 'Merging…' : 'Merge'}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
