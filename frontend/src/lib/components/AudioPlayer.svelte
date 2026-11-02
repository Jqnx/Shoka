<script lang="ts">
	import { Play } from '@lucide/svelte';
	import { tick } from 'svelte';
	import { audioTrackSrc, formatDuration } from '$lib/audio';
	import type { AudioTrack } from '$lib/types';

	let {
		audioId,
		updatedAt,
		tracks
	}: {
		audioId: string;
		updatedAt: string;
		tracks: AudioTrack[];
	} = $props();

	let selected = $state(0);
	// The element is only read from event handlers, so it needs no reactivity.
	let player: HTMLAudioElement | null = null;

	// A reload can shrink the track list (a rescan removed files); clamping at
	// read time keeps the player from ever getting an empty src.
	const current = $derived(
		tracks.length === 0 ? 0 : Math.min(Math.max(selected, 0), tracks.length - 1)
	);
	const active = $derived(tracks[current] ?? null);
	const src = $derived(active ? audioTrackSrc(audioId, active.index, updatedAt) : '');

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

	function selectTrack(index: number) {
		if (index !== current) selected = index;
		void playCurrent();
	}

	// Stop at the last track instead of wrapping around: a work is meant to
	// be listened to front to back once.
	async function handleEnded() {
		if (current >= tracks.length - 1) return;

		selected = current + 1;
		await playCurrent();
	}
</script>

{#if tracks.length === 0}
	<p class="text-sm text-muted-foreground">No audio files.</p>
{:else}
	<audio
		{@attach capturePlayer}
		controls
		preload="metadata"
		{src}
		onended={handleEnded}
		class="w-full"
	></audio>

	<ol class="mt-2 divide-y divide-border rounded-md border border-border">
		{#each tracks as track (track.index)}
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
