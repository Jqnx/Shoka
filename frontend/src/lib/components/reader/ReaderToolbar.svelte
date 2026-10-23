<script lang="ts">
	import type { ReaderSettings } from '$lib/types';
	import { resolve } from '$app/paths';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import {
		ChevronLeft,
		ChevronRight,
		ChevronsLeft,
		ChevronsRight,
		Heart,
		ImagePlus,
		Loader2,
		Settings,
		Undo2
	} from '@lucide/svelte';
	import ReaderSettingsPanel from './ReaderSettingsPanel.svelte';

	let {
		title,
		backHref,
		currentPage,
		pageCount,
		visible,
		settings,
		settingsOpen = $bindable(false),
		onSettingsChange,
		coverPage,
		settingCover,
		coverError,
		canSetCover,
		favorited,
		onSetCover,
		onToggleFavorite,
		onPinnedChange,
		onFirst,
		onPrev,
		onNext,
		onLast
	}: {
		title: string;
		backHref: string;
		currentPage: number;
		pageCount: number;
		visible: boolean;
		settings: ReaderSettings;
		settingsOpen?: boolean;
		onSettingsChange: (patch: Partial<ReaderSettings>) => void;
		/** 0-based page currently used as the archive's cover. */
		coverPage: number;
		settingCover: boolean;
		coverError: string | null;
		canSetCover: boolean;
		/** Whether the current page is in the user's page favorites. */
		favorited: boolean;
		onSetCover: () => void;
		onToggleFavorite: () => void;
		/** Reports pointer/focus entering or leaving the island (pauses auto-hide). */
		onPinnedChange?: (pinned: boolean) => void;
		onFirst: () => void;
		onPrev: () => void;
		onNext: () => void;
		onLast: () => void;
	} = $props();

	const isCover = $derived(currentPage === coverPage);
	const atStart = $derived(currentPage === 0);
	const atEnd = $derived(currentPage >= pageCount - 1);

	// Reading direction drives which side of the bar advances: in rtl the
	// pages themselves progress leftward, so the left-hand chevrons step
	// forward and the right-hand ones step back - same mapping the tap zones
	// and keyboard arrows use in Reader.svelte.
	const rtl = $derived(settings.reading_direction === 'rtl');

	// Every control is a full-height cell that fills its own column, so the
	// hover fill covers the entire bar height rather than a boxed button.
	// `w-9` below `sm` so `n of total` still fits beside them on a 360px phone -
	// unlike a real grid column, a `basis-*` on a shrinkable item can't widen an
	// intrinsically-sized flex line, so these have to be definite widths.
	const cell =
		'flex h-full w-8 shrink-0 items-center justify-center text-card-foreground transition-colors select-none hover:bg-accent hover:text-accent-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset focus-visible:outline-none disabled:pointer-events-none disabled:opacity-30 sm:w-12';

	// Pointer pinning only for an actual mouse: a touchscreen - including a
	// hybrid laptop's - still fires pointerenter/leave for taps, and pinning
	// for those would keep the bar up until the next tap lands elsewhere.
	// Checking the event's pointerType keeps this per-interaction, so a mouse
	// plugged in later works without re-evaluating anything at mount.
	let hoverPinned = false;
	let focusPinned = false;

	function reportPinned() {
		onPinnedChange?.(hoverPinned || focusPinned);
	}

	function handlePointerEnter(event: PointerEvent) {
		if (event.pointerType !== 'mouse') return;
		hoverPinned = true;
		reportPinned();
	}

	function handlePointerLeave(event: PointerEvent) {
		if (event.pointerType !== 'mouse') return;
		hoverPinned = false;
		reportPinned();
	}

	// Focus only counts when it's keyboard-driven. A mouse click focuses the
	// button it lands on too, and that focus never leaves by itself - pinning
	// for it would keep the bar on screen forever after one click.
	function handleFocusIn(event: FocusEvent) {
		focusPinned = !!((event.target as HTMLElement | null)?.matches?.(':focus-visible') ?? false);
		reportPinned();
	}

	function handleFocusOut() {
		focusPinned = false;
		reportPinned();
	}

	// The parent's flag only ever changes on a report from here, so a stale
	// `true` would stick forever if this component were torn down or re-created
	// while it was pinned. A fresh instance therefore asserts the unpinned
	// state on mount and releases the pin on teardown.
	$effect(() => {
		reportPinned();
		return () => onPinnedChange?.(false);
	});

	let islandEl = $state<HTMLDivElement | null>(null);

	// Hiding only translates the island, so a control a mouse click focused
	// (which never pins - `:focus-visible` is false for it) would keep focus
	// off-screen and still fire on Enter. Keyboard focus can't hit this: it
	// pins the bar, so the bar isn't hiding. Drop the stray focus on hide.
	$effect(() => {
		if (visible) return;
		const active = document.activeElement;
		if (active instanceof HTMLElement && islandEl?.contains(active)) active.blur();
	});
</script>

<!--
	The positioning shell is inert (pointer-events-none) so the gutters either
	side of the island don't swallow taps meant for the page behind it; only the
	island itself takes input back.
-->
<div class="pointer-events-none fixed inset-x-0 bottom-4 z-50 flex justify-center px-3">
	<!--
	The island's pointer/focus handlers only pause the auto-hide countdown; the
	wrapper itself is not a control, so there is nothing for AT to activate here.
-->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		bind:this={islandEl}
		class="pointer-events-auto relative w-fit max-w-full rounded-2xl border border-border bg-card/95 text-card-foreground shadow-lg backdrop-blur transition-transform duration-200 {visible
			? 'translate-y-0'
			: 'translate-y-[calc(100%+1rem)]'}"
		onpointerenter={handlePointerEnter}
		onpointerleave={handlePointerLeave}
		onfocusin={handleFocusIn}
		onfocusout={handleFocusOut}
	>
		{#if coverError}
			<!--
				The pill sits above the island, so the hide transform alone would
				leave it stranded on screen after the bar slides away - it has to be
				hidden explicitly. `pointer-events-none` because it is a passive
				message: its strip must keep passing taps to the page behind it.
			-->
			<p
				class="pointer-events-none absolute inset-x-0 bottom-full mb-2 rounded-lg border border-border bg-card/95 px-3 py-2 text-center text-xs text-destructive {visible
					? ''
					: 'invisible'}"
			>
				{coverError}
			</p>
		{/if}

		<!--
			`overflow-hidden` + `rounded-[inherit]` so a hovered corner cell follows
			the island's rounding instead of squaring it off; kept on the nav rather
			than the shell above, which has to let the error strip escape it.
		-->
		<nav
			class="flex h-12 items-stretch overflow-hidden rounded-[inherit]"
			aria-label="Reader controls"
		>
			<a
				href={resolve(backHref as `/${string}`)}
				class={cell}
				aria-label="Back to {title}"
				title="Back to {title}"
			>
				<Undo2 class="size-4.5" />
			</a>

			<button
				type="button"
				class={cell}
				disabled={rtl ? atEnd : atStart}
				aria-label={rtl ? 'Go to last page' : 'Go to first page'}
				title={rtl ? 'Go to last page' : 'Go to first page'}
				onclick={rtl ? onLast : onFirst}
			>
				<ChevronsLeft class="size-4.5" />
			</button>

			<button
				type="button"
				class={cell}
				disabled={rtl ? atEnd : atStart}
				aria-label={rtl ? 'Next page' : 'Previous page'}
				title={rtl ? 'Next page' : 'Previous page'}
				onclick={rtl ? onNext : onPrev}
			>
				<ChevronLeft class="size-4.5" />
			</button>

			<!--
				`shrink-0` so a narrow bar can never eat into the one readout it has:
				the cells either side keep their definite width and the island's own
				`max-w-full` caps the whole thing instead.
			-->
			<p
				class="flex h-full shrink-0 items-center justify-center px-2 text-xs font-medium whitespace-nowrap tabular-nums sm:text-sm"
			>
				{currentPage + 1} of {pageCount}
			</p>

			<button
				type="button"
				class={cell}
				disabled={rtl ? atStart : atEnd}
				aria-label={rtl ? 'Previous page' : 'Next page'}
				title={rtl ? 'Previous page' : 'Next page'}
				onclick={rtl ? onPrev : onNext}
			>
				<ChevronRight class="size-4.5" />
			</button>

			<button
				type="button"
				class={cell}
				disabled={rtl ? atStart : atEnd}
				aria-label={rtl ? 'Go to first page' : 'Go to last page'}
				title={rtl ? 'Go to first page' : 'Go to last page'}
				onclick={rtl ? onFirst : onLast}
			>
				<ChevronsRight class="size-4.5" />
			</button>

			{#if canSetCover}
				<button
					type="button"
					class={cell}
					disabled={isCover || settingCover}
					aria-label={isCover ? 'Current page is already the cover' : 'Set current page as cover'}
					title={isCover ? 'Current page is already the cover' : 'Set current page as cover'}
					onclick={onSetCover}
				>
					{#if settingCover}
						<Loader2 class="size-4.5 animate-spin" />
					{:else}
						<ImagePlus class={isCover ? 'size-4.5 fill-current' : 'size-4.5'} />
					{/if}
				</button>
			{/if}

			<button
				type="button"
				class={cell}
				aria-label={favorited ? 'Remove current page from favorites' : 'Favorite current page'}
				title={favorited ? 'Remove current page from favorites' : 'Favorite current page'}
				aria-pressed={favorited}
				onclick={onToggleFavorite}
			>
				<Heart class={favorited ? 'size-4.5 fill-red-500 stroke-red-500' : 'size-4.5'} />
			</button>

			<Popover.Root bind:open={settingsOpen}>
				<Popover.Trigger>
					{#snippet child({ props })}
						<button {...props} class={cell} aria-label="Reader settings" title="Reader settings">
							<Settings class="size-4.5" />
						</button>
					{/snippet}
				</Popover.Trigger>
				<Popover.Content side="top" align="end" class="w-64">
					<ReaderSettingsPanel {settings} onChange={onSettingsChange} />
				</Popover.Content>
			</Popover.Root>
		</nav>
	</div>
</div>
