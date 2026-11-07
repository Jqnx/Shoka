<script lang="ts">
	import { Pause, Play } from '@lucide/svelte';
	import { formatDuration } from '$lib/audio';
	import { cn } from '$lib/utils.js';
	import type { AudioTrack } from '$lib/types';

	let {
		tracks,
		activeIndex,
		playing,
		onSelect,
		class: className = ''
	}: {
		tracks: AudioTrack[];
		activeIndex: number | null;
		playing: boolean;
		onSelect: (index: number) => void;
		class?: string;
	} = $props();
</script>

<ol class={cn('divide-y divide-border rounded-md border border-border', className)}>
	{#each tracks as track (track.index)}
		{@const isActive = track.index === activeIndex}
		<li>
			<button
				type="button"
				data-track-index={track.index}
				aria-current={isActive ? 'true' : undefined}
				class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-muted/50 {isActive
					? 'font-medium text-foreground'
					: 'text-muted-foreground'}"
				onclick={() => onSelect(track.index)}
			>
				<span
					class="flex size-6 shrink-0 items-center justify-center rounded-full border border-border"
				>
					{#if isActive && playing}
						<Pause class="size-3 fill-current" />
					{:else}
						<Play class="size-3" />
					{/if}
				</span>
				<span class="min-w-0 flex-1 truncate">{track.name}</span>
				<span class="shrink-0 text-xs tabular-nums">
					{formatDuration(track.duration_seconds)}
				</span>
			</button>
		</li>
	{/each}
</ol>
