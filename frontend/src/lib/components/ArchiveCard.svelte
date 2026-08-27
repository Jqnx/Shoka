<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { Badge } from '$lib/components/ui/badge';
	import { Check } from '@lucide/svelte';
	import type { Archive } from '$lib/types';

	let {
		archive,
		selectionMode = false,
		selected = false,
		onLongPress,
		onToggle
	}: {
		archive: Archive;
		/** When true, a plain click toggles selection instead of navigating. */
		selectionMode?: boolean;
		selected?: boolean;
		/** Fired after a sustained press; the parent enters selection mode. */
		onLongPress?: (id: string) => void;
		onToggle?: (id: string) => void;
	} = $props();

	// Carries the current list page (sort/filters/pagination and all) through
	// to the archive detail page, so its "back" link can return here exactly
	// as left, instead of a bare library link.
	const from = $derived(page.url.pathname + page.url.search);
	const href = $derived(resolve(`/a/${archive.id}?from=${encodeURIComponent(from)}`));

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
		longPressFired = false;
		startX = e.clientX;
		startY = e.clientY;
		cancelPress();
		pressTimer = setTimeout(() => {
			longPressFired = true;
			onLongPress?.(archive.id);
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
			onToggle?.(archive.id);
		}
	}

	// Only swallow the context menu when it follows a touch/pen long-press
	// (where it would otherwise interrupt the hold-to-select gesture). A
	// mouse right-click keeps its native menu.
	function handleContextMenu(e: MouseEvent) {
		if (lastPointerType !== 'mouse') e.preventDefault();
	}

	// Keyboard parity: in selection mode the anchor shouldn't navigate, and
	// Space is the conventional toggle key once it acts like a checkbox.
	function handleKeyDown(e: KeyboardEvent) {
		if (!selectionMode) return;
		if (e.key === ' ' || e.key === 'Enter') {
			e.preventDefault();
			onToggle?.(archive.id);
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
	class="group relative flex cursor-pointer touch-pan-y flex-col overflow-hidden rounded-lg border bg-card transition-all select-none hover:-translate-y-0.5 hover:shadow-md {selected
		? 'border-primary ring-2 ring-primary'
		: 'border-border'}"
>
	<!-- Cover -->
	<div class="relative aspect-[2/3] w-full overflow-hidden bg-muted">
		<!-- Anchors and images are natively draggable: a mousedown plus the
		     slightest wobble starts a drag, which fires pointercancel and
		     kills an in-progress hold. -->
		<img
			src="/api/archives/{archive.id}/cover"
			alt={archive.title}
			draggable="false"
			class="absolute inset-0 size-full object-cover"
		/>
		<div
			class="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-black/80 via-black/30 to-transparent opacity-0 transition-opacity group-hover:opacity-100"
		></div>

		{#if archive.progress}
			<div class="absolute inset-x-0 top-0 h-1.5">
				<div
					class="h-full bg-sidebar-primary transition-all"
					style="width: {Math.round((archive.progress.current_page / archive.page_count) * 100)}%"
				></div>
			</div>
		{/if}

		<div class="absolute top-1.5 right-1.5">
			<span class="rounded bg-black/60 px-1.5 py-0.5 text-[11px] font-medium text-white/90">
				{archive.page_count}p
			</span>
		</div>

		{#if selectionMode}
			<div class="absolute inset-0 bg-primary/10"></div>
			<div class="absolute top-1.5 left-1.5">
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
	</div>

	<!-- Card body -->
	<div class="flex flex-col gap-1 p-2">
		<p
			class="line-clamp-2 text-sm leading-tight font-medium text-card-foreground group-hover:text-foreground"
		>
			{archive.title}
		</p>
		{#if archive.artists?.length}
			<p class="truncate text-xs text-muted-foreground">{archive.artists.join(', ')}</p>
		{/if}
		<div class="mt-0.5 flex flex-wrap gap-1">
			{#if archive.category}
				<Badge variant="secondary" class="h-4 px-1.5 text-[11px]">{archive.category}</Badge>
			{/if}
			{#if archive.language}
				<Badge variant="outline" class="h-4 px-1.5 text-[11px]">{archive.language}</Badge>
			{/if}
		</div>
	</div>
</a>
