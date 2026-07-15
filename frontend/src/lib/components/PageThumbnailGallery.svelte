<script lang="ts">
	import { SvelteSet } from 'svelte/reactivity';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { ImageOff } from '@lucide/svelte';

	let {
		archiveId,
		pageCount,
		thumbsReady
	}: {
		archiveId: string;
		pageCount: number;
		thumbsReady: boolean;
	} = $props();

	const readyIndexes = new SvelteSet<number>();
	let genError = $state<string | null>(null);

	// Triggers generation (idempotent - a no-op if already ready/running) and
	// streams progress via SSE, filling in thumbnails as they're generated
	// instead of waiting for the whole archive to finish.
	$effect(() => {
		readyIndexes.clear();
		genError = null;

		if (thumbsReady) {
			for (let i = 0; i < pageCount; i++) readyIndexes.add(i);
			return;
		}

		fetch(`/api/archives/${archiveId}/thumbnails`, { method: 'POST' }).catch(() => {});

		const source = new EventSource(`/api/archives/${archiveId}/thumbnails/events`);

		source.addEventListener('ready', (e) => {
			const payload = JSON.parse((e as MessageEvent).data) as { index: number };
			readyIndexes.add(payload.index);
		});

		source.addEventListener('done', (e) => {
			const payload = JSON.parse((e as MessageEvent).data) as { done: boolean; error?: string };
			if (payload.error) genError = payload.error;
			source.close();
		});

		source.onerror = () => {
			if (source.readyState === EventSource.CLOSED) {
				source.close();
			}
		};

		return () => source.close();
	});

	const pages = $derived(Array.from({ length: pageCount }, (_, i) => i));
</script>

<div>
	<h2 class="mb-3 text-sm font-semibold">Pages</h2>

	{#if genError}
		<p class="mb-3 flex items-center gap-1.5 text-sm text-destructive">
			<ImageOff class="size-4" />
			Failed to generate thumbnails: {genError}
		</p>
	{/if}

	<div class="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8">
		{#each pages as index (index)}
			<div class="aspect-[2/3] overflow-hidden rounded-md border border-border bg-muted">
				{#if readyIndexes.has(index)}
					<img
						src="/api/archives/{archiveId}/pages/{index}/thumbnail"
						alt="Page {index + 1}"
						loading="lazy"
						class="size-full object-cover"
					/>
				{:else}
					<Skeleton class="size-full" />
				{/if}
			</div>
		{/each}
	</div>
</div>
