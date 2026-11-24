<script lang="ts">
	import { resolve } from '$app/paths';
	import { AudioLines, Check } from '@lucide/svelte';
	import { audioCoverSrc, type AudioListItem } from '$lib/types';
	import { formatDuration } from '$lib/audio';

	let {
		audio,
		selectionMode = false,
		selected = false,
		onLongPress,
		onToggle
	}: {
		audio: AudioListItem;
		/** When true, a plain click toggles selection instead of navigating. */
		selectionMode?: boolean;
		selected?: boolean;
		/** Fired after a sustained press; the parent enters selection mode. */
		onLongPress?: (id: string) => void;
		onToggle?: (id: string) => void;
	} = $props();

	const href = $derived(resolve(`/audio/${audio.id}`));

	const LONG_PRESS_MS = 800;
	// A press that drifts further than this is a scroll/drag, not a hold -
	// without it, swiping the grid on a touch screen would select cards.
	const MOVE_TOLERANCE_PX = 10;

	let pressTimer: ReturnType<typeof setTimeout> | undefined;
	let startX = 0;
	let startY = 0;
	// Which pointer type started the current interaction - so the context
	// menu is only suppressed for a touch/pen long-press, never a real
	// desktop right-click (which must keep "open in new tab" etc.).
	let lastPointerType = 'mouse';
	// Set when the timer fires so the click that follows the release can be
	// swallowed - otherwise releasing a long press would immediately toggle
	// off the selection it just created (or navigate).
	let longPressFired = false;

	function cancelPress() {
		clearTimeout(pressTimer);
		pressTimer = undefined;
	}

	function handlePointerDown(e: PointerEvent) {
		lastPointerType = e.pointerType;
		// Ignore secondary/middle buttons so right-click and open-in-new-tab
		// keep behaving normally.
		if (e.pointerType === 'mouse' && e.button !== 0) return;
		// Call sites that don't offer selection (favorites, voice-actor pages)
		// must not arm the timer: it would swallow the click that follows a
		// long hold, making the card look unresponsive.
		if (!onLongPress) return;
		longPressFired = false;
		startX = e.clientX;
		startY = e.clientY;
		cancelPress();
		pressTimer = setTimeout(() => {
			longPressFired = true;
			onLongPress?.(audio.id);
		}, LONG_PRESS_MS);
	}

	function handlePointerMove(e: PointerEvent) {
		if (!pressTimer) return;
		if (
			Math.abs(e.clientX - startX) > MOVE_TOLERANCE_PX ||
			Math.abs(e.clientY - startY) > MOVE_TOLERANCE_PX
		) {
			cancelPress();
		}
	}

	function handleClick(e: MouseEvent) {
		if (longPressFired) {
			// The hold already selected this card; consume its click.
			e.preventDefault();
			longPressFired = false;
			return;
		}
		if (selectionMode) {
			e.preventDefault();
			onToggle?.(audio.id);
		}
	}

	// Only swallow the context menu when it follows a touch/pen long-press
	// (where it would otherwise interrupt the hold-to-select gesture). A
	// mouse right-click keeps its native menu.
	function handleContextMenu(e: MouseEvent) {
		if (onLongPress && lastPointerType !== 'mouse') e.preventDefault();
	}

	// Keyboard parity: in selection mode the anchor shouldn't navigate, and
	// Space is the conventional toggle key once it acts like a checkbox.
	function handleKeyDown(e: KeyboardEvent) {
		if (!selectionMode) return;
		if (e.key === ' ' || e.key === 'Enter') {
			e.preventDefault();
			onToggle?.(audio.id);
		}
	}
</script>

<a
	href={selectionMode ? undefined : href}
	data-sveltekit-preload-data="tap"
	draggable="false"
	role={selectionMode ? 'checkbox' : undefined}
	aria-checked={selectionMode ? selected : undefined}
	tabindex={selectionMode ? 0 : undefined}
	onpointerdown={handlePointerDown}
	onpointermove={handlePointerMove}
	onpointerup={cancelPress}
	onpointercancel={cancelPress}
	onpointerleave={cancelPress}
	onclick={handleClick}
	onkeydown={handleKeyDown}
	oncontextmenu={handleContextMenu}
	class="group relative flex h-full cursor-pointer touch-pan-y flex-col overflow-hidden rounded-lg border bg-card transition-all select-none hover:-translate-y-0.5 hover:shadow-md {selected
		? 'border-primary ring-2 ring-primary'
		: 'border-border'}"
>
	<div class="relative aspect-[4/3] w-full overflow-hidden bg-muted">
		{#if audio.cover_ready}
			<img
				src={audioCoverSrc(audio)}
				alt={audio.title}
				loading="lazy"
				class="absolute inset-0 size-full object-cover"
			/>
		{:else}
			<div class="absolute inset-0 flex items-center justify-center">
				<AudioLines class="size-10 text-muted-foreground/40" />
			</div>
		{/if}

		{#if audio.is_played && !selectionMode}
			<div class="pointer-events-none absolute top-1.5 left-1.5">
				<span
					class="flex items-center gap-1 rounded bg-black/60 px-1.5 py-0.5 text-[11px] font-medium text-white/90"
				>
					<Check class="size-3" />
					Played
				</span>
			</div>
		{/if}

		{#if selectionMode}
			<div class="pointer-events-none absolute inset-0 bg-primary/10"></div>
			<div class="pointer-events-none absolute top-1.5 left-1.5">
				<div
					class="flex size-5 items-center justify-center rounded-full border-2 {selected
						? 'border-primary bg-primary text-primary-foreground'
						: 'border-white/80 bg-black/40'}"
				>
					{#if selected}
						<Check class="size-3" />
					{/if}
				</div>
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
