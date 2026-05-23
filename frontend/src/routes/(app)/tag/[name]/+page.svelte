<script lang="ts">
	import { resolve } from '$app/paths';
	import { ChevronLeft, Tag } from '@lucide/svelte';
	import ArchiveCard from '$lib/components/ArchiveCard.svelte';

	let { data } = $props();
</script>

<div class="mx-auto max-w-screen-2xl px-4 py-8 sm:px-6">
	<a
		href={resolve('/a')}
		class="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
	>
		<ChevronLeft class="size-4" />
		Archives
	</a>

	<div class="mb-6 flex items-center gap-2">
		<Tag class="size-5 text-muted-foreground" />
		<h1 class="text-2xl font-bold">{data.name}</h1>
		<span class="text-sm text-muted-foreground">({data.total})</span>
	</div>

	{#if data.archives.length === 0}
		<p class="text-muted-foreground">No archives found.</p>
	{:else}
		<div class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8">
			{#each data.archives as archive (archive.id)}
				<ArchiveCard {archive} />
			{/each}
		</div>
	{/if}
</div>
