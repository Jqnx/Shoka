<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';

	let {
		src,
		downloadName,
		mouseOnly = false,
		actions,
		trigger
	}: {
		/** Image URL used by the browser-item actions. */
		src: string;
		/** Optional filename for Save image as…; defaults to the URL's last segment. */
		downloadName?: string;
		/**
		 * Only open on a mouse right-click. Use where the containing element owns
		 * a touch hold gesture (e.g. card long-press selection).
		 */
		mouseOnly?: boolean;
		/** App-specific items rendered after the browser items. */
		actions?: Snippet;
		/** Renders the caller's existing element as the trigger (no wrapper DOM). */
		trigger: Snippet<[{ props: Record<string, unknown> }]>;
	} = $props();

	let open = $state(false);
	let canGoBack = $state(false);
	let canGoForward = $state(false);

	// Set by the trigger's pointer handlers; read once when the menu opens to
	// tell a touch long-press apart from a mouse right-click.
	let lastPointerType: string | null = null;
	// A touch release still synthesizes a click on the trigger after pointerup;
	// swallow that one so the underlying link/tap-zone doesn't navigate.
	let suppressOpeningClick = false;

	function isTouchLike(event: MouseEvent): boolean {
		const pointerType = (event as PointerEvent).pointerType;
		if (pointerType) return pointerType !== 'mouse';
		const capabilities = (
			event as MouseEvent & {
				sourceCapabilities?: { firesTouchEvents?: boolean } | null;
			}
		).sourceCapabilities;
		if (typeof capabilities?.firesTouchEvents === 'boolean') return capabilities.firesTouchEvents;
		// Firefox/Safari don't expose sourceCapabilities; the trigger's
		// pointerdown already recorded whether this interaction is touch.
		return lastPointerType !== null && lastPointerType !== 'mouse';
	}

	// A delegated trigger's props include bits-ui's own handlers. Record the
	// pointer type they saw, and in mouseOnly mode keep touch/pen events from
	// starting the primitive's long-press timer.
	function guardTriggerProps(props: Record<string, unknown>): Record<string, unknown> {
		const guarded = { ...props };

		for (const name of [
			'onpointerdown',
			'onpointermove',
			'onpointercancel',
			'onpointerup'
		] as const) {
			const handler = props[name];
			if (typeof handler !== 'function') continue;
			guarded[name] = (event: PointerEvent) => {
				if (event.pointerType) lastPointerType = event.pointerType;
				if (mouseOnly && event.pointerType !== 'mouse') return;
				(handler as (event: PointerEvent) => void)(event);
			};
		}

		const contextmenu = props.oncontextmenu;
		if (typeof contextmenu === 'function') {
			guarded.oncontextmenu = (event: MouseEvent) => {
				const touch = isTouchLike(event);
				lastPointerType = touch ? 'touch' : 'mouse';
				if (mouseOnly && touch) return;
				(contextmenu as (event: MouseEvent) => void)(event);
			};
		}

		const onclick = props.onclick;
		guarded.onclick = (event: MouseEvent) => {
			if (suppressOpeningClick) {
				suppressOpeningClick = false;
				event.preventDefault();
				event.stopPropagation();
				return;
			}
			if (typeof onclick === 'function') (onclick as (event: MouseEvent) => void)(event);
		};

		return guarded;
	}

	// Chromium dispatches contextmenu on the right-button press, so the menu is
	// already mounted under the cursor by the time the button comes back up.
	// Bits' item then treats that pointerup as a click and activates whichever
	// item is under the cursor (Back, on a page-sized image). Touch has the same
	// race after the trigger's long-press timer. Swallow the release that opened
	// the menu unless a fresh press landed inside it first.
	$effect(() => {
		if (!open) return;

		const openedByTouch = lastPointerType !== null && lastPointerType !== 'mouse';
		lastPointerType = null;
		suppressOpeningClick = openedByTouch;
		let menuPressed = false;
		let timer: ReturnType<typeof setTimeout>;

		const onPointerDown = (e: PointerEvent) => {
			if (e.target instanceof Element && e.target.closest('[data-slot="context-menu-content"]')) {
				menuPressed = true;
			}
		};
		const cleanup = () => {
			window.removeEventListener('pointerdown', onPointerDown, true);
			window.removeEventListener('pointerup', swallow, true);
			clearTimeout(timer);
		};
		function swallow(e: PointerEvent) {
			if (menuPressed) return; // a deliberate press inside the menu
			const touchRelease = e.pointerType !== 'mouse';
			if (e.button !== 2 && !(openedByTouch && touchRelease)) return;
			e.preventDefault();
			e.stopPropagation();
			cleanup();
		}
		timer = setTimeout(cleanup, 1000);
		window.addEventListener('pointerdown', onPointerDown, true);
		window.addEventListener('pointerup', swallow, true);

		return () => {
			suppressOpeningClick = false;
			cleanup();
		};
	});

	// 'sveltekit:history' is the index SvelteKit writes onto each history entry.
	// It is internal but read-only and defensive: a missing key falls back to 0.
	$effect(() => {
		if (!open) return;

		const index = (history.state as Record<string, unknown> | null)?.['sveltekit:history'];
		const position = typeof index === 'number' ? index : 0;
		canGoBack = position > 0;
		canGoForward = position < history.length - 1;
	});

	function goBack() {
		history.back();
	}

	function goForward() {
		history.forward();
	}

	function reload() {
		location.reload();
	}

	function openInNewTab() {
		window.open(src, '_blank', 'noopener');
	}

	// Chromium's clipboard only accepts PNG, while pages and thumbnails are
	// webp. The fetch + canvas conversion is what makes the copy possible.
	async function toPng(url: string): Promise<Blob> {
		const response = await fetch(url);
		if (!response.ok) throw new Error(`Failed to fetch image: ${response.status}`);

		const bitmap = await createImageBitmap(await response.blob());
		try {
			const canvas = document.createElement('canvas');
			canvas.width = bitmap.width;
			canvas.height = bitmap.height;
			const context = canvas.getContext('2d');
			if (!context) throw new Error('Canvas 2D context unavailable');
			context.drawImage(bitmap, 0, 0);

			return await new Promise<Blob>((resolve, reject) => {
				canvas.toBlob(
					(blob) => (blob ? resolve(blob) : reject(new Error('PNG encoding failed'))),
					'image/png'
				);
			});
		} finally {
			bitmap.close();
		}
	}

	async function copyImage() {
		try {
			// Constructed synchronously so the click's transient user activation
			// still applies when clipboard.write runs a task later.
			const item = new ClipboardItem({ 'image/png': toPng(src) });
			await navigator.clipboard.write([item]);
		} catch {
			// No toast system, and Chromium's own menu fails silently too.
		}
	}

	async function copyImageAddress() {
		try {
			await navigator.clipboard.writeText(src);
		} catch {
			// See copyImage.
		}
	}

	function saveImage() {
		const link = document.createElement('a');
		link.href = src;
		link.download = downloadName ?? src.split('/').pop() ?? 'image';
		link.click();
	}
</script>

{#snippet delegatedTrigger({ props }: { props: Record<string, unknown> })}
	{@render trigger({ props: guardTriggerProps(props) })}
{/snippet}

<ContextMenu.Root bind:open>
	<ContextMenu.Trigger child={delegatedTrigger} />
	<ContextMenu.Content class="min-w-56">
		<ContextMenu.Item disabled={!canGoBack} onSelect={goBack}>
			Back
			<ContextMenu.Shortcut>Alt+←</ContextMenu.Shortcut>
		</ContextMenu.Item>
		<ContextMenu.Item disabled={!canGoForward} onSelect={goForward}>
			Forward
			<ContextMenu.Shortcut>Alt+→</ContextMenu.Shortcut>
		</ContextMenu.Item>
		<ContextMenu.Item onSelect={reload}>
			Reload
			<ContextMenu.Shortcut>Ctrl+R</ContextMenu.Shortcut>
		</ContextMenu.Item>
		<ContextMenu.Separator />
		<ContextMenu.Item onSelect={openInNewTab}>Open image in new tab</ContextMenu.Item>
		<ContextMenu.Item onSelect={copyImage}>Copy image</ContextMenu.Item>
		<ContextMenu.Item onSelect={copyImageAddress}>Copy image address</ContextMenu.Item>
		<ContextMenu.Item onSelect={saveImage}>Save image as…</ContextMenu.Item>
		{#if actions}
			<ContextMenu.Separator />
			{@render actions()}
		{/if}
	</ContextMenu.Content>
</ContextMenu.Root>
