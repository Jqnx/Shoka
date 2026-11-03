<script lang="ts">
	import { fly } from 'svelte/transition';
	import {
		AudioLines,
		ChevronUp,
		LoaderCircle,
		Pause,
		Play,
		RotateCcw,
		RotateCw,
		SkipBack,
		SkipForward,
		Square,
		Volume1,
		Volume2,
		VolumeX
	} from '@lucide/svelte';
	import AudioSeekSlider from '$lib/components/AudioSeekSlider.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Slider } from '$lib/components/ui/slider/index.js';
	import { formatDuration } from '$lib/audio';
	import { player } from '$lib/player.svelte';
</script>

{#if player.source}
	{@const source = player.source}
	{@const VolumeIcon =
		player.muted || player.volume === 0 ? VolumeX : player.volume < 0.5 ? Volume1 : Volume2}

	{#if !player.expanded && player.barVisible}
		<div
			in:fly={{ y: 72, duration: 220 }}
			out:fly={{ y: 72, duration: 180 }}
			class="fixed right-0 bottom-0 left-0 z-40 flex h-[4.5rem] items-center gap-2 border-t border-border bg-background/95 px-3 text-foreground backdrop-blur supports-backdrop-filter:bg-background/85 sm:px-4 md:left-(--sidebar-width) md:peer-data-[state=collapsed]:left-0"
			title={player.error ?? undefined}
		>
			<button
				type="button"
				class="flex min-w-0 flex-1 items-center gap-2 text-left"
				aria-label={`Open player: ${player.track?.name ?? source.title}`}
				onclick={() => (player.expanded = true)}
			>
				{#if source.coverUrl}
					<img src={source.coverUrl} alt="" class="size-12 shrink-0 rounded-md object-cover" />
				{:else}
					<span
						class="flex size-12 shrink-0 items-center justify-center rounded-md bg-muted"
						aria-hidden="true"
					>
						<AudioLines class="size-5 text-muted-foreground" />
					</span>
				{/if}
				<span class="flex min-w-0 flex-col">
					<span class="truncate text-sm font-medium">{player.track?.name ?? source.title}</span>
					<span class="truncate text-xs text-muted-foreground">{source.title}</span>
				</span>
			</button>

			<div class="flex min-w-0 flex-none flex-col items-center justify-center gap-0.5 sm:flex-[2]">
				<div class="flex shrink-0 items-center gap-1">
					<Button
						variant="ghost"
						size="icon"
						aria-label="Previous track"
						onclick={() => player.previous()}
					>
						<SkipBack class="size-4" />
					</Button>
					<Button
						size="icon-lg"
						aria-label={player.playing ? 'Pause' : 'Play'}
						onclick={() => player.toggle()}
					>
						{#if player.loading && !player.playing}
							<LoaderCircle class="size-4 animate-spin" />
						{:else if player.playing}
							<Pause class="size-4 fill-current" />
						{:else}
							<Play class="size-4 fill-current" />
						{/if}
					</Button>
					<Button variant="ghost" size="icon" aria-label="Next track" onclick={() => player.next()}>
						<SkipForward class="size-4" />
					</Button>
					<Button variant="ghost" size="icon" aria-label="Stop" onclick={() => player.stop()}>
						<Square class="size-3.5 fill-current" />
					</Button>
				</div>

				<div class="hidden w-full max-w-xl items-center gap-2 sm:flex">
					<span class="w-10 shrink-0 text-right text-xs tabular-nums">
						{formatDuration(player.currentTime)}
					</span>
					<AudioSeekSlider class="flex-1" />
					<span class="w-10 shrink-0 text-xs tabular-nums">
						{formatDuration(player.duration)}
					</span>
				</div>
			</div>

			<div class="flex min-w-0 shrink-0 items-center justify-end gap-1 sm:flex-1">
				<Button
					variant="ghost"
					size="icon"
					aria-label={player.muted ? 'Unmute' : 'Mute'}
					onclick={() => player.toggleMute()}
				>
					<VolumeIcon class="size-4" />
				</Button>
				<div class="hidden w-24 lg:block">
					<Slider
						type="single"
						min={0}
						max={1}
						step={0.01}
						value={player.volume}
						ariaLabel="Volume"
						onValueChange={(value: number) => player.setVolume(value)}
					/>
				</div>
				<Button
					variant="ghost"
					size="icon"
					aria-label="Expand player"
					onclick={() => (player.expanded = true)}
				>
					<ChevronUp class="size-4" />
				</Button>
			</div>

			<AudioSeekSlider class="absolute inset-x-0 top-0 sm:hidden" />
		</div>
	{/if}

	{#if player.expanded && player.barVisible}
		<div
			in:fly={{ y: 24, duration: 220 }}
			out:fly={{ y: 16, duration: 160 }}
			role="region"
			aria-label="Now playing"
			class="fixed right-2 bottom-2 z-50 flex max-h-[calc(100svh-1.5rem)] w-72 flex-col gap-3 overflow-y-auto rounded-lg border border-border bg-popover p-2.5 text-popover-foreground shadow-xl sm:right-3 sm:bottom-3"
		>
			<div class="flex h-6 shrink-0 items-center justify-center">
				<button
					type="button"
					class="flex h-4 w-28 items-center justify-center rounded-full transition-colors hover:bg-muted focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-hidden"
					aria-label="Collapse player"
					onclick={() => (player.expanded = false)}
				>
					<span class="h-1.5 w-24 rounded-full bg-muted-foreground/40"></span>
				</button>
			</div>

			{#if source.coverUrl}
				<img src={source.coverUrl} alt={source.title} class="block w-full" />
			{:else}
				<div
					class="flex aspect-square w-full items-center justify-center rounded-md border border-border bg-muted"
				>
					<AudioLines class="size-8 text-muted-foreground" />
				</div>
			{/if}

			<div class="flex items-center gap-2">
				<span class="w-10 shrink-0 text-[11px] text-muted-foreground tabular-nums">
					{formatDuration(player.currentTime)}
				</span>
				<AudioSeekSlider class="flex-1" />
				<span class="w-10 shrink-0 text-right text-[11px] text-muted-foreground tabular-nums">
					{formatDuration(player.duration)}
				</span>
			</div>

			<div class="min-w-0">
				<p class="truncate text-sm font-medium">{player.track?.name ?? source.title}</p>
				<p class="truncate text-center text-xs text-muted-foreground">{source.title}</p>
			</div>

			<div class="flex items-center justify-center gap-1">
				<Button
					variant="ghost"
					size="icon"
					aria-label="Previous track"
					onclick={() => player.previous()}
				>
					<SkipBack class="size-4" />
				</Button>
				<Button
					variant="ghost"
					size="icon"
					aria-label="Seek back 10 seconds"
					onclick={() => player.seek(player.currentTime - 10)}
				>
					<RotateCcw class="size-4" />
				</Button>
				<Button
					size="icon-lg"
					aria-label={player.playing ? 'Pause' : 'Play'}
					onclick={() => player.toggle()}
				>
					{#if player.loading && !player.playing}
						<LoaderCircle class="size-4 animate-spin" />
					{:else if player.playing}
						<Pause class="size-4 fill-current" />
					{:else}
						<Play class="size-4 fill-current" />
					{/if}
				</Button>
				<Button
					variant="ghost"
					size="icon"
					aria-label="Seek forward 10 seconds"
					onclick={() => player.seek(player.currentTime + 10)}
				>
					<RotateCw class="size-4" />
				</Button>
				<Button variant="ghost" size="icon" aria-label="Next track" onclick={() => player.next()}>
					<SkipForward class="size-4" />
				</Button>
				<Button variant="ghost" size="icon" aria-label="Stop" onclick={() => player.stop()}>
					<Square class="size-4 fill-current" />
				</Button>
			</div>

			<div class="flex items-center gap-2">
				<Button
					variant="ghost"
					size="icon"
					aria-label={player.muted ? 'Unmute' : 'Mute'}
					onclick={() => player.toggleMute()}
				>
					<VolumeIcon class="size-4" />
				</Button>
				<Slider
					type="single"
					min={0}
					max={1}
					step={0.01}
					value={player.volume}
					ariaLabel="Volume"
					class="flex-1"
					onValueChange={(value: number) => player.setVolume(value)}
				/>
				<Volume2 class="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
			</div>

			{#if player.error}
				<p class="text-xs text-destructive">{player.error}</p>
			{/if}
		</div>
	{/if}
{/if}
