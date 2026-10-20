<script lang="ts">
	import type { ReaderSettings } from '$lib/types';
	import { onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PagedView from './PagedView.svelte';
	import ScrollView from './ScrollView.svelte';
	import ReaderToolbar from './ReaderToolbar.svelte';
	import { backgroundClass } from './reader-utils';

	let {
		archiveId,
		title,
		pageCount,
		backHref,
		initialSettings,
		initialPage,
		initialCoverPage,
		canSetCover
	}: {
		archiveId: string;
		title: string;
		pageCount: number;
		backHref: string;
		initialSettings: ReaderSettings;
		// 0-based - already validated/clamped server-side against the
		// archive's actual page count (see +page.server.ts).
		initialPage: number;
		// 0-based page currently used as the archive's cover (Archive.cover_page).
		initialCoverPage: number;
		// Whether the reader may set a page as the cover. Resolved in
		// +page.server.ts because this route resets past the (app) layout.
		canSetCover: boolean;
	} = $props();

	// Settings only ever need to be seeded once from the server-loaded prop -
	// after that, this component's own `settings` state is the source of
	// truth (updateSettings below keeps the backend in sync, not the other
	// way around).
	// svelte-ignore state_referenced_locally
	let settings = $state({ ...initialSettings });

	function clamp(n: number) {
		return Math.min(Math.max(n, 0), pageCount - 1);
	}

	// Deep-linkable/refresh-safe position - the server-loaded route param is
	// already validated/clamped against the page count, so it only needs
	// seeding once here; this component's own `currentPage` state is the
	// source of truth from then on (the effect below mirrors it back into
	// the URL, not the other way around).
	// svelte-ignore state_referenced_locally
	let currentPage = $state(initialPage);

	let toolbarVisible = $state(true);
	// True while the pointer is over the island or focus is inside it - the
	// auto-hide countdown is paused so the bar can't vanish out from under a
	// click, or from under a keyboard user mid-Tab through its buttons.
	let toolbarPinned = $state(false);
	let settingsOpen = $state(false);

	// Idle time before the toolbar slips away on its own. Short enough to get
	// the chrome out of the way while reading, long enough to reach for it.
	const TOOLBAR_HIDE_DELAY = 3000;

	// Restarts on every transition that should grant a fresh window: the bar
	// becoming visible (mount, tap, popover closed) or the popover opening or
	// closing. Page turns deliberately don't touch it - the countdown measures
	// time since you last asked for the controls, not since the last page.
	$effect(() => {
		if (!toolbarVisible || toolbarPinned || settingsOpen) return;

		const timeout = setTimeout(() => (toolbarVisible = false), TOOLBAR_HIDE_DELAY);
		return () => clearTimeout(timeout);
	});

	// This component's own source of truth from here on, same pattern as
	// `settings`/`currentPage` above - seeded once from the server load.
	// svelte-ignore state_referenced_locally
	let coverPage = $state(initialCoverPage);
	let settingCover = $state(false);
	let coverError = $state<string | null>(null);

	async function setCover(page: number) {
		settingCover = true;
		coverError = null;
		try {
			const res = await fetch(`/api/archives/${archiveId}/cover`, {
				method: 'PUT',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ page })
			});
			if (!res.ok) {
				coverError = 'Failed to set cover.';
				return;
			}
			coverPage = page;
		} catch {
			coverError = 'Failed to set cover.';
		} finally {
			settingCover = false;
		}
	}

	// A two-page spread advances two at a time, but only in paged view: the
	// scroll view lays pages out one under another, so a two-page step would
	// skip one. The stored `page_layout` survives switching to scroll.
	const step = $derived(
		settings.view_mode === 'paged' && settings.page_layout === 'double' ? 2 : 1
	);

	function goToNext() {
		currentPage = clamp(currentPage + step);
	}
	function goToPrev() {
		currentPage = clamp(currentPage - step);
	}
	function goToFirst() {
		currentPage = 0;
	}
	function goToLast() {
		currentPage = clamp(pageCount - 1);
	}

	// Tap/keyboard "right" and "left" flip which direction they navigate
	// depending on reading direction, matching the physical layout - in rtl
	// the next page is to the left, same as the pages themselves reading
	// right-to-left.
	function rightZone() {
		if (settings.reading_direction === 'ltr') goToNext();
		else goToPrev();
	}
	function leftZone() {
		if (settings.reading_direction === 'ltr') goToPrev();
		else goToNext();
	}

	// Tap-to-toggle; the auto-hide effect above re-arms whenever this flips
	// the bar back on.
	function toggleToolbar() {
		toolbarVisible = !toolbarVisible;
	}

	// Focus can still land in the island while it's off-screen (it stays
	// focusable), so treat that as a request to bring it back - typing into
	// invisible controls is worse than the bar reappearing - and keep it up
	// for as long as the focus stays inside.
	function handlePinnedChange(pinned: boolean) {
		toolbarPinned = pinned;
		if (pinned) toolbarVisible = true;
	}

	function handleScrollPageChange(index: number) {
		currentPage = index;
	}

	async function updateSettings(patch: Partial<ReaderSettings>) {
		settings = { ...settings, ...patch };
		try {
			await fetch('/api/reader-settings', {
				method: 'PATCH',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify(patch)
			});
		} catch {
			// Best-effort - the change still applies for this session even if it
			// fails to sync, it just won't carry over to another device.
		}
	}

	// PUT is idempotent and explicitly documented as safe to call on every
	// page turn, but it's still debounced here so flipping through several
	// pages quickly sends one request once things settle rather than one per
	// page. lastSavedPage skips redundant re-sends of the same value (e.g.
	// the debounce firing right as onDestroy's own flush already sent it).
	let lastSavedPage = -1;

	function flushProgress(page: number, keepalive = false) {
		if (page === lastSavedPage) return;
		lastSavedPage = page;
		fetch(`/api/archives/${archiveId}/progress`, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ page }),
			keepalive
		}).catch(() => {
			// Best-effort - a failed save just means this position doesn't
			// stick; nothing else in the reader depends on it succeeding.
		});
	}

	$effect(() => {
		const page = currentPage;
		const timeout = setTimeout(() => flushProgress(page), 600);
		return () => clearTimeout(timeout);
	});

	// Covers leaving via SvelteKit's own client-side router (e.g. clicking
	// the toolbar's back link) - that unmounts this component but never
	// unloads the document, so pagehide below wouldn't fire, and the pending
	// debounce above would get cancelled by its own cleanup before it had a
	// chance to run. onDestroy is also the one lifecycle hook that *does*
	// run during SSR (there's no real unmount there, so Svelte just fires it
	// immediately after the render) - guarded with `browser` since a
	// relative-URL fetch has no base to resolve against on the server and
	// would otherwise throw and crash the render entirely.
	onDestroy(() => {
		if (browser) flushProgress(currentPage);
	});

	// Covers the opposite case: a hard navigation or closing the tab (e.g.
	// the Escape handler below, which sets window.location.href directly
	// rather than going through the router) - the component here is never
	// unmounted by Svelte since the whole document goes away, so onDestroy
	// never runs. keepalive gives this request a chance to actually
	// complete during unload. $effect never runs during SSR at all, so this
	// one doesn't need the same browser guard as onDestroy above.
	$effect(() => {
		const handler = () => flushProgress(currentPage, true);
		window.addEventListener('pagehide', handler);
		return () => window.removeEventListener('pagehide', handler);
	});

	function handleKeydown(e: KeyboardEvent) {
		if (e.ctrlKey || e.metaKey || e.altKey) return;

		// The page context menu focuses its own content; let bits-ui's handlers
		// own Escape/arrows/Space there instead of turning pages or exiting.
		if (e.target instanceof Element && e.target.closest('[data-slot="context-menu-content"]')) {
			return;
		}

		// Own Escape-closes-the-popover ourselves rather than relying on the
		// popover's own handling - bits-ui's internal Escape listener sits on
		// `document` and fires during the bubble phase before a same-phase
		// `window` listener would, so by the time a bubble-phase handler here
		// re-read `settingsOpen` it'd already have been flipped to false,
		// letting Escape fall through and exit the whole reader instead of
		// just closing the panel. Listening in the capture phase (see
		// `onkeydowncapture` below) instead makes this the *first* handler to
		// see the event, before bits-ui's, so `settingsOpen` is still
		// accurate here.
		if (e.key === 'Escape' && settingsOpen) {
			e.preventDefault();
			settingsOpen = false;
			return;
		}
		// Otherwise leave the popover's own keyboard handling alone while open
		// (e.g. arrow keys moving focus between its buttons).
		if (settingsOpen) return;

		switch (e.key) {
			case 'ArrowRight':
				e.preventDefault();
				rightZone();
				break;
			case 'ArrowLeft':
				e.preventDefault();
				leftZone();
				break;
			case ' ':
				e.preventDefault();
				if (e.shiftKey) goToPrev();
				else goToNext();
				break;
			case 'Backspace':
				e.preventDefault();
				goToPrev();
				break;
			case 'Home':
				e.preventDefault();
				goToFirst();
				break;
			case 'End':
				e.preventDefault();
				goToLast();
				break;
			case 'Escape':
				e.preventDefault();
				window.location.href = backHref;
				break;
		}
	}

	// Keeps the /a/[id]/[page] path in sync without spamming browser history -
	// each turn replaces the current entry, so the back button leaves the
	// reader entirely rather than stepping through pages one at a time.
	// Wrapped in try/catch since this effect's very first run happens right
	// at mount, sometimes before SvelteKit's client router has finished
	// initializing - replaceState throws in that narrow window. Harmless to
	// skip: the URL already reflects the correct initial page either way.
	$effect(() => {
		try {
			replaceState(resolve(`/a/${archiveId}/${currentPage + 1}`), {});
		} catch {
			// see comment above
		}
	});
</script>

<svelte:window onkeydowncapture={handleKeydown} />

<svelte:head>
	<title>{title} | Shoka</title>
</svelte:head>

<div class="relative {backgroundClass(settings.background)}">
	<ReaderToolbar
		{title}
		{backHref}
		{currentPage}
		{pageCount}
		visible={toolbarVisible}
		{settings}
		bind:settingsOpen
		onSettingsChange={updateSettings}
		{coverPage}
		{settingCover}
		{coverError}
		{canSetCover}
		onSetCover={() => setCover(currentPage)}
		onPinnedChange={handlePinnedChange}
		onFirst={goToFirst}
		onPrev={goToPrev}
		onNext={goToNext}
		onLast={goToLast}
	/>

	{#if settings.view_mode === 'scroll'}
		<ScrollView
			{archiveId}
			{pageCount}
			{settings}
			initialPage={currentPage}
			{coverPage}
			{canSetCover}
			onSetCover={(i) => setCover(i)}
			onCurrentPageChange={handleScrollPageChange}
			onToggleToolbar={toggleToolbar}
		/>
	{:else}
		<PagedView
			{archiveId}
			{pageCount}
			{currentPage}
			{settings}
			{coverPage}
			{canSetCover}
			onSetCover={(i) => setCover(i)}
			onLeftZone={leftZone}
			onRightZone={rightZone}
			onToggleToolbar={toggleToolbar}
		/>
	{/if}
</div>
