<script lang="ts">
	import { SvelteSet } from 'svelte/reactivity';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { audioImagePreviewSrc, audioImageThumbSrc } from '$lib/audio';
	import type { AudioImage } from '$lib/types';

	let { audioId, updatedAt, images }: { audioId: string; updatedAt: string; images: AudioImage[] } =
		$props();

	const readyIndexes = new SvelteSet<number>();
	let previewIndex = $state<number | null>(null);
	let genError = $state<string | null>(null);

	const previewImage = $derived(
		previewIndex === null ? null : (images.find((image) => image.index === previewIndex) ?? null)
	);

	// Triggers generation (idempotent - a no-op if already ready/running) and
	// streams progress via SSE, filling thumbnails in as they are generated.
	// This trigger is ungated, so there is no permission check here.
	$effect(() => {
		const id = audioId;

		readyIndexes.clear();
		genError = null;
		previewIndex = null;

		fetch(`/api/audio/${id}/thumbnails`, { method: 'POST' }).catch(() => {});

		const source = new EventSource(`/api/audio/${id}/thumbnails/events`);

		source.addEventListener('ready', (e) => {
			const payload = JSON.parse((e as MessageEvent).data) as { index: number };
			readyIndexes.add(payload.index);
		});

		source.addEventListener('done', (e) => {
			const payload = JSON.parse((e as MessageEvent).data) as { done: boolean; error?: string };
			if (payload.error) {
				genError = payload.error;
			} else {
				// The broadcaster drops events beyond its buffer, so a
				// successful done means every thumbnail is on disk.
				for (const image of images) readyIndexes.add(image.index);
			}
			source.close();
		});

		source.onerror = () => {
			if (source.readyState === EventSource.CLOSED) {
				source.close();
			}
		};

		return () => source.close();
	});
</script>

<div>
	{#if genError}
		<p class="mb-3 text-sm text-destructive">Failed to generate thumbnails: {genError}</p>
	{/if}

	<div
		class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8"
	>
		{#each images as image (image.index)}
			<button
				type="button"
				class="aspect-square overflow-hidden rounded-lg border border-border bg-muted"
				onclick={() => (previewIndex = image.index)}
				aria-label="View {image.name}"
			>
				{#if readyIndexes.has(image.index)}
					<img
						src={audioImageThumbSrc(audioId, image.index, updatedAt)}
						alt={image.name}
						loading="lazy"
						decoding="async"
						class="size-full object-cover transition-transform hover:scale-105"
					/>
				{:else}
					<Skeleton class="size-full" />
				{/if}
			</button>
		{/each}
	</div>
</div>

<Dialog.Root
	open={previewIndex !== null}
	onOpenChange={(isOpen) => {
		if (!isOpen) previewIndex = null;
	}}
>
	<Dialog.Content class="sm:max-w-[min(90vw,72rem)]">
		<Dialog.Header>
			<Dialog.Title class="sr-only">{previewImage?.name ?? 'Image'}</Dialog.Title>
		</Dialog.Header>
		{#if previewImage}
			<img
				src={audioImagePreviewSrc(audioId, previewImage.index, updatedAt)}
				alt={previewImage.name}
				class="max-h-[80vh] w-full object-contain"
			/>
		{/if}
	</Dialog.Content>
</Dialog.Root>
