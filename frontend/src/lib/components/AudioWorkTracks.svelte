<script lang="ts">
	import { Pause, Play } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import AudioTrackList from '$lib/components/AudioTrackList.svelte';
	import { player, sourceFromAudio } from '$lib/player.svelte';
	import type { Audio, AudioTrack } from '$lib/types';

	let { audio, defaultFolder }: { audio: Audio; defaultFolder: string } = $props();

	// Works can ship several versions as sibling folders (with/without sound
	// effects). Group in scan order; the API already returns tracks sorted by
	// path, so each folder's tracks stay contiguous.
	const groups = $derived.by(() => {
		const grouped: { folder: string; tracks: AudioTrack[] }[] = [];

		for (const track of audio.tracks) {
			const group = grouped.find((candidate) => candidate.folder === track.folder);
			if (group) group.tracks.push(track);
			else grouped.push({ folder: track.folder, tracks: [track] });
		}

		return grouped;
	});

	// svelte-ignore state_referenced_locally
	let selectedFolder = $state<string | null>(defaultFolder === '' ? null : defaultFolder);

	// Clamp at read time: the configured folder may not exist in every work,
	// and a reload can remove the folder or track currently selected. A null
	// selection means "no preference" so it cannot be confused with the ''
	// group that holds root-level tracks.
	const activeFolder = $derived(
		selectedFolder !== null && groups.some((group) => group.folder === selectedFolder)
			? selectedFolder
			: (groups[0]?.folder ?? '')
	);
	const activeTracks = $derived(
		groups.find((group) => group.folder === activeFolder)?.tracks ?? []
	);

	const sourceIsCurrent = $derived(player.isCurrent(audio.id, audio.updated_at, activeFolder));
	// Stop hides the player but keeps the source; the work page must not show
	// that retained track as active until the bar is visible again.
	const showsActive = $derived(player.barVisible && sourceIsCurrent);
	const activeIndex = $derived(showsActive ? player.trackIndex : null);
	const playingHere = $derived(showsActive && player.playing);

	function folderLabel(folder: string) {
		return folder === '' ? 'Top level' : folder;
	}

	function selectFolder(folder: string) {
		if (folder === activeFolder) return;

		selectedFolder = folder;
	}

	function toggleFolder() {
		if (activeTracks.length === 0) return;

		if (sourceIsCurrent) player.toggle();
		else player.playSource(sourceFromAudio(audio, activeFolder), activeTracks[0].index);
	}

	function selectTrack(index: number) {
		if (player.isCurrent(audio.id, audio.updated_at, activeFolder, index)) player.toggle();
		else player.playSource(sourceFromAudio(audio, activeFolder), index);
	}
</script>

{#if audio.tracks.length === 0}
	<p class="text-sm text-muted-foreground">No audio files.</p>
{:else}
	<div class="flex flex-wrap items-center justify-between gap-2">
		<div>
			{#if groups.length > 1}
				<div class="flex flex-wrap gap-1.5" role="group" aria-label="Version folder">
					{#each groups as group (group.folder)}
						{@const isActive = group.folder === activeFolder}
						<button
							type="button"
							class="rounded-md border px-2.5 py-1 text-xs font-medium transition-colors {isActive
								? 'border-primary bg-primary text-primary-foreground'
								: 'border-border text-muted-foreground hover:text-foreground'}"
							aria-label="{folderLabel(group.folder)} ({group.tracks.length} tracks)"
							aria-pressed={isActive}
							onclick={() => selectFolder(group.folder)}
						>
							{folderLabel(group.folder)}
							<span class="ms-1 opacity-70">{group.tracks.length}</span>
						</button>
					{/each}
				</div>
			{:else if activeFolder}
				<p class="text-xs text-muted-foreground">{folderLabel(activeFolder)}</p>
			{/if}
		</div>

		<Button size="sm" onclick={toggleFolder} disabled={activeTracks.length === 0}>
			{#if playingHere}
				<Pause class="size-3.5" />
				Pause
			{:else}
				<Play class="size-3.5" />
				Play
			{/if}
		</Button>
	</div>

	<AudioTrackList
		tracks={activeTracks}
		{activeIndex}
		playing={player.playing}
		onSelect={selectTrack}
		class="mt-3"
	/>
{/if}
