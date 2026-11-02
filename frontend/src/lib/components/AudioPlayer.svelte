<script lang="ts">
	import { Play } from '@lucide/svelte';
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

	let current = $state(0);

	const active = $derived(tracks[current] ?? null);
	const src = $derived(active ? audioTrackSrc(audioId, active.index, updatedAt) : '');

	// Stop at the last track instead of wrapping around: a work is meant to
	// be listened to front to back once.
	function handleEnded() {
		if (current < tracks.length - 1) current += 1;
	}
</script>

{#if tracks.length === 0}
	<p class="text-sm text-muted-foreground">No audio files.</p>
{:else}
	<audio controls preload="metadata" {src} onended={handleEnded} class="w-full"></audio>

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
					onclick={() => (current = track.index)}
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
