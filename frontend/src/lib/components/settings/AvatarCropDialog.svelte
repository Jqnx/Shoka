<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Slider } from '$lib/components/ui/slider/index.js';
	import Cropper from 'svelte-easy-crop';
	import type { CropArea } from 'svelte-easy-crop';

	const OUTPUT_SIZE = 512;

	let {
		open = $bindable(false),
		image,
		oncropped
	}: {
		open?: boolean;
		/** Object URL for the source file - caller owns revocation. */
		image: string;
		oncropped: (blob: Blob) => void;
	} = $props();

	// The caller wraps this component in `{#key}` on a nonce bumped per file
	// pick (see AvatarCard.svelte), so a fresh instance - and fresh initial
	// state here - is exactly what re-picking a file gets. No reset effect
	// needed.
	let crop = $state({ x: 0, y: 0 });
	let zoom = $state(1);
	let pixels = $state<CropArea | null>(null);
	let applying = $state(false);
	let error = $state<string | null>(null);

	function loadImage(src: string): Promise<HTMLImageElement> {
		return new Promise((resolve, reject) => {
			const img = new Image();
			img.onload = () => resolve(img);
			img.onerror = () => reject(new Error('Could not load image.'));
			img.src = src;
		});
	}

	async function applyCrop() {
		if (!pixels) return;
		applying = true;
		error = null;

		try {
			const img = await loadImage(image);
			const canvas = document.createElement('canvas');
			canvas.width = OUTPUT_SIZE;
			canvas.height = OUTPUT_SIZE;
			const ctx = canvas.getContext('2d');
			if (!ctx) throw new Error('Canvas unavailable.');
			ctx.imageSmoothingQuality = 'high';

			// svelte-easy-crop reports fractional pixels in source-image space;
			// clamp so a rounding edge case can't hand drawImage a negative or
			// out-of-bounds source rect.
			const sx = Math.max(0, Math.round(pixels.x));
			const sy = Math.max(0, Math.round(pixels.y));
			const sw = Math.min(Math.round(pixels.width), img.naturalWidth - sx);
			const sh = Math.min(Math.round(pixels.height), img.naturalHeight - sy);

			ctx.drawImage(img, sx, sy, sw, sh, 0, 0, OUTPUT_SIZE, OUTPUT_SIZE);

			const blob = await new Promise<Blob | null>((resolve) =>
				canvas.toBlob(resolve, 'image/webp', 0.9)
			);
			if (!blob) throw new Error('Could not generate image.');

			oncropped(blob);
			open = false;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not crop image.';
		} finally {
			applying = false;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Crop avatar</Dialog.Title>
			<Dialog.Description>Drag to reposition, use the slider to zoom.</Dialog.Description>
		</Dialog.Header>

		<div class="relative h-72 w-full overflow-hidden rounded-md bg-muted">
			<Cropper
				{image}
				bind:crop
				bind:zoom
				aspect={1}
				cropShape="round"
				minZoom={1}
				maxZoom={3}
				oncropcomplete={(e) => (pixels = e.pixels)}
			/>
		</div>

		<div class="flex items-center gap-3">
			<span class="text-xs text-muted-foreground">Zoom</span>
			<Slider type="single" bind:value={zoom} min={1} max={3} step={0.01} class="flex-1" />
		</div>

		{#if error}
			<p class="text-sm text-destructive">{error}</p>
		{/if}

		<Dialog.Footer showCloseButton>
			<Button type="button" disabled={applying || !pixels} onclick={applyCrop}>
				{applying ? 'Applying…' : 'Apply'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
