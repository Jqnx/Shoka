<script lang="ts">
	import { resolve } from '$app/paths';
	import { AudioLines } from '@lucide/svelte';
	import AudioCard from '$lib/components/AudioCard.svelte';
	import type { AudioListItem } from '$lib/types';

	let {
		works,
		total,
		href,
		heading = 'Audio'
	}: {
		works: AudioListItem[];
		total: number;
		href: string;
		heading?: string;
	} = $props();
</script>

{#if total > 0}
	{@const viewAll = href as `/${string}`}
	<div class="mb-4 flex items-center gap-2">
		<AudioLines class="size-4 text-muted-foreground" />
		<h2 class="text-lg font-semibold">{heading}</h2>
		<span class="text-sm text-muted-foreground">({total})</span>
		{#if total > works.length}
			<a
				href={resolve(viewAll)}
				class="ml-auto text-xs text-muted-foreground transition-colors hover:text-foreground"
			>
				View all audio →
			</a>
		{/if}
	</div>

	<div
		class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8"
	>
		{#each works as work (work.id)}
			<AudioCard audio={work} />
		{/each}
	</div>
{/if}
