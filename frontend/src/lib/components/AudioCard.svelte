<script lang="ts">
	import { resolve } from '$app/paths';
	import { AudioLines } from '@lucide/svelte';
	import { audioCoverSrc, type AudioListItem } from '$lib/types';
	import { formatDuration } from '$lib/audio';

	let { audio }: { audio: AudioListItem } = $props();
</script>

<a
	href={resolve(`/audio/${audio.id}`)}
	data-sveltekit-preload-data="tap"
	class="group relative flex h-full cursor-pointer flex-col overflow-hidden rounded-lg border border-border bg-card transition-all select-none hover:-translate-y-0.5 hover:shadow-md"
>
	<div class="relative w-full overflow-hidden bg-muted">
		{#if audio.cover_ready}
			<img
				src={audioCoverSrc(audio)}
				alt={audio.title}
				loading="lazy"
				width={audio.cover_width || undefined}
				height={audio.cover_height || undefined}
				class="block h-auto w-full"
			/>
		{:else}
			<div class="flex aspect-square items-center justify-center">
				<AudioLines class="size-10 text-muted-foreground/40" />
			</div>
		{/if}

		<div class="pointer-events-none absolute right-1.5 bottom-1.5">
			<span class="rounded bg-black/60 px-1.5 py-0.5 text-[11px] font-medium text-white/90">
				{formatDuration(audio.duration_seconds)}
			</span>
		</div>
	</div>

	<div class="flex flex-col gap-1 p-2">
		<p
			class="line-clamp-2 text-xs leading-tight font-medium text-card-foreground group-hover:text-foreground"
		>
			{audio.title}
		</p>
		{#if audio.circle}
			<p class="truncate text-xs text-muted-foreground">{audio.circle}</p>
		{/if}
	</div>
</a>
