<script lang="ts">
	import { Play } from '@lucide/svelte';
	import { tick } from 'svelte';
	import { audioTrackSrc, formatDuration } from '$lib/audio';
	import type { AudioTrack } from '$lib/types';

	let {
		audioId,
		updatedAt,
		tracks,
		defaultFolder
	}: {
		audioId: string;
		updatedAt: string;
		tracks: AudioTrack[];
		/** Top-level track folder to open; empty or missing falls back to the first. */
		defaultFolder: string;
	} = $props();

	// The element is only read from event handlers, so it needs no reactivity.
	let player: HTMLAudioElement | null = null;

	// Works can ship several versions as sibling folders (with/without sound
	// effects). Group in scan order; the API already returns tracks sorted by
	// path, so each folder's tracks stay contiguous.
	const groups = $derived.by(() => {
		const grouped: { folder: string; tracks: AudioTrack[] }[] = [];

		for (const track of tracks) {
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

	let selected = $state<number | null>(null);

	const current = $derived.by(() => {
		if (activeTracks.length === 0) return 0;

		return activeTracks.find((track) => track.index === selected)?.index ?? activeTracks[0].index;
	});
	const active = $derived(tracks.find((track) => track.index === current) ?? null);
	const src = $derived(active ? audioTrackSrc(audioId, active.index, updatedAt) : '');

	function folderLabel(folder: string) {
		return folder === '' ? 'Top level' : folder;
	}

	// Attachment instead of bind:this: the element is captured without making
	// it a reactive binding.
	function capturePlayer(node: HTMLAudioElement) {
		player = node;

		return () => {
			if (player === node) player = null;
		};
	}

	// Changing src resets the element to paused, so a new track only actually
	// plays if play() is issued after the DOM has the new source.
	async function playCurrent() {
		await tick();
		await player?.play().catch(() => {});
	}

	function selectFolder(folder: string) {
		if (folder === activeFolder) return;

		selectedFolder = folder;
		// Open the folder on its first track but do not start playback.
		selected = null;
	}

	function selectTrack(index: number) {
		if (index !== current) selected = index;
		void playCurrent();
	}

	// Stop at the last track of the open folder instead of wrapping around,
	// and never spill into the next version.
	async function handleEnded() {
		const index = activeTracks.findIndex((track) => track.index === current);
		if (index === -1 || index >= activeTracks.length - 1) return;

		selected = activeTracks[index + 1].index;
		await playCurrent();
	}
</script>

{#if tracks.length === 0}
	<p class="text-sm text-muted-foreground">No audio files.</p>
{:else}
	{#if groups.length > 1}
		<div class="mb-3 flex flex-wrap gap-1.5" role="group" aria-label="Version folder">
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
		<p class="mb-2 text-xs text-muted-foreground">{folderLabel(activeFolder)}</p>
	{/if}

	<audio
		{@attach capturePlayer}
		controls
		preload="metadata"
		{src}
		onended={handleEnded}
		class="w-full"
	></audio>

	<ol class="mt-2 divide-y divide-border rounded-md border border-border">
		{#each activeTracks as track (track.index)}
			<li>
				<button
					type="button"
					class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-muted/50 {track.index ===
					current
						? 'font-medium text-foreground'
						: 'text-muted-foreground'}"
					aria-current={track.index === current ? 'true' : undefined}
					onclick={() => selectTrack(track.index)}
				>
					<span
						class="flex size-6 shrink-0 items-center justify-center rounded-full border border-border"
					>
						{#if track.index === current}
							<Play class="size-3 fill-current" />
						{:else}
							<Play class="size-3" />
						{/if}
					</span>
					<span class="min-w-0 flex-1 truncate">{track.name}</span>
					<span class="shrink-0 text-xs tabular-nums">{formatDuration(track.duration_seconds)}</span
					>
				</button>
			</li>
		{/each}
	</ol>
{/if}
