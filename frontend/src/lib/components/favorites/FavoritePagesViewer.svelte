<script lang="ts">
	import type { FitMode, FavoritePageItem, ReaderSettings } from '$lib/types';
	import { goto, replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Slider } from '$lib/components/ui/slider/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import {
		BookOpen,
		ChevronLeft,
		ChevronRight,
		Heart,
		Pause,
		Play,
		Settings,
		Shuffle,
		Undo2
	} from '@lucide/svelte';
	import { backgroundClass, pageImageStyle, pageUrl } from '$lib/components/reader/reader-utils';

	let {
		items: initialItems,
		initialAt,
		initialSettings
	}: {
		items: FavoritePageItem[];
		initialAt: number;
		initialSettings: ReaderSettings;
	} = $props();

	// The list is mutated locally as items are unfavorited, so keep our own
	// copy rather than writing through the (frozen) prop.
	// svelte-ignore state_referenced_locally
	let items = $state([...initialItems]);
	// svelte-ignore state_referenced_locally
	let settings = $state({ ...initialSettings });
	// svelte-ignore state_referenced_locally
	let at = $state(initialAt);
	let playing = $state(false);
	// Session-only on purpose: a shuffled playback order across devices would
	// be surprising, unlike the persisted slideshow timing.
	let shuffle = $state(false);
	// Maps display position -> items index.
	let order = $state<number[]>(items.map((_, i) => i));
	let settingsOpen = $state(false);
	let scroller = $state<HTMLDivElement | null>(null);
	let progressBar = $state<HTMLDivElement | null>(null);

	const current = $derived(items.length > 0 ? items[order[at]] : undefined);
	const favorited = $derived(current !== undefined);

	// Reset to the top-left whenever the displayed item changes - including
	// after an unfavorite at the same `at`, which slides the next item in.
	$effect(() => {
		void current;
		scroller?.scrollTo({ top: 0, left: 0 });
	});

	const nextItem = $derived.by(() => {
		if (items.length === 0) return undefined;
		if (at + 1 < order.length) return items[order[at + 1]];
		if (settings.slideshow_loop) return items[order[0]];
		return undefined;
	});

	function fisherYates(values: number[]): number[] {
		const shuffled = [...values];
		for (let i = shuffled.length - 1; i > 0; i--) {
			const j = Math.floor(Math.random() * (i + 1));
			[shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
		}
		return shuffled;
	}

	function toggleShuffle() {
		if (!shuffle) {
			// Keep the current item first so starting playback doesn't jump.
			const currentIndex = order[at];
			order = [currentIndex, ...fisherYates(order.filter((i) => i !== currentIndex))];
			at = 0;
			shuffle = true;
		} else {
			shuffle = false;
			order = items.map((_, i) => i);
		}
	}

	function goToPrev() {
		if (at > 0) at--;
	}

	function goToNext() {
		if (at + 1 < items.length) at++;
	}

	function advance() {
		if (at + 1 < items.length) at++;
		else if (settings.slideshow_loop) at = 0;
		else playing = false;
	}

	function togglePlay() {
		playing = !playing;
	}

	// Re-arms on every manual navigation (the `at` read below) and whenever
	// the interval changes, so the next page gets the full new delay.
	$effect(() => {
		if (!playing) return;

		const interval = settings.slideshow_interval_seconds;
		const position = at;
		const timer = setTimeout(() => {
			if (position === at) advance();
		}, interval * 1000);

		return () => clearTimeout(timer);
	});

	// Driven directly rather than through $state so 60fps updates don't
	// re-render the viewer. Restarts whenever the slide or interval changes,
	// mirroring the timer above.
	$effect(() => {
		const bar = progressBar;
		if (!bar) return;

		if (!playing) {
			bar.style.transform = 'scaleX(0)';
			return;
		}

		void at;

		const duration = settings.slideshow_interval_seconds * 1000;
		const start = performance.now();
		let frame = 0;

		const tick = (now: number) => {
			const progress = Math.min((now - start) / duration, 1);
			bar.style.transform = `scaleX(${progress})`;
			if (progress < 1) frame = requestAnimationFrame(tick);
		};

		frame = requestAnimationFrame(tick);

		return () => cancelAnimationFrame(frame);
	});

	function applyLocal(patch: Partial<ReaderSettings>) {
		settings = { ...settings, ...patch };
	}

	async function updateSettings(patch: Partial<ReaderSettings>) {
		applyLocal(patch);
		try {
			await fetch('/api/reader-settings', {
				method: 'PATCH',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify(patch)
			});
		} catch {
			// Best-effort - the change still applies this session even if it
			// fails to sync to another device.
		}
	}

	async function unfavoriteCurrent() {
		const item = current;
		if (!item) return;

		try {
			const res = await fetch(
				`/api/archives/${item.archive_id}/pages/${item.page_index}/favorite`,
				{ method: 'DELETE' }
			);
			if (!res.ok) return;
		} catch {
			return;
		}

		const removedIndex = order[at];
		items.splice(removedIndex, 1);
		order = order.filter((i) => i !== removedIndex).map((i) => (i > removedIndex ? i - 1 : i));

		if (items.length === 0) {
			goto(resolve('/favorites'));
			return;
		}

		at = Math.min(at, items.length - 1);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.ctrlKey || e.metaKey || e.altKey) return;

		if (e.target instanceof Element && e.target.closest('[data-slot="context-menu-content"]')) {
			return;
		}

		if (e.key === 'Escape' && settingsOpen) {
			e.preventDefault();
			settingsOpen = false;
			return;
		}
		if (settingsOpen) return;

		switch (e.key) {
			case 'ArrowLeft':
				e.preventDefault();
				goToPrev();
				break;
			case 'ArrowRight':
				e.preventDefault();
				goToNext();
				break;
			case ' ':
				e.preventDefault();
				togglePlay();
				break;
			case 'Home':
				e.preventDefault();
				at = 0;
				break;
			case 'End':
				e.preventDefault();
				at = items.length - 1;
				break;
			case 'Escape':
				e.preventDefault();
				goto(resolve('/favorites'));
				break;
		}
	}

	// Keep the position in the URL without adding history entries, so a
	// refresh/deep link lands on the same slide.
	$effect(() => {
		try {
			replaceState(resolve(`/favorites/viewer?at=${at}`), {});
		} catch {
			// See Reader.svelte: the very first effect run can precede the
			// router's own initialisation.
		}
	});

	const cell =
		'flex h-full w-9 shrink-0 items-center justify-center text-card-foreground transition-colors select-none hover:bg-accent hover:text-accent-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset focus-visible:outline-none disabled:pointer-events-none disabled:opacity-30 sm:w-11';
</script>

<svelte:window onkeydowncapture={handleKeydown} />

<svelte:head>
	<title>{current ? `${current.title} | Favorite Pages | Shoka` : 'Favorite Pages | Shoka'}</title>
</svelte:head>

<div
	bind:this={scroller}
	class="relative h-dvh overflow-auto {backgroundClass(settings.background)}"
>
	<div class="flex h-full min-w-full">
		{#if current}
			<img
				src={pageUrl(current.archive_id, current.page_index)}
				alt="{current.title} page {current.page_index + 1}"
				style={pageImageStyle(settings.fit_mode)}
				class="m-auto select-none"
				draggable="false"
			/>
		{/if}
	</div>

	{#if nextItem}
		<img
			src={pageUrl(nextItem.archive_id, nextItem.page_index)}
			alt=""
			class="hidden"
			aria-hidden="true"
		/>
	{/if}

	<div class="pointer-events-none fixed inset-x-0 bottom-4 z-50 flex justify-center px-3">
		<div
			class="pointer-events-auto relative w-fit max-w-full rounded-2xl border border-border bg-card/95 text-card-foreground shadow-lg backdrop-blur"
		>
			<nav
				class="flex h-12 max-w-full items-stretch overflow-hidden rounded-[inherit]"
				aria-label="Favorite pages viewer"
			>
				<a
					href={resolve('/favorites')}
					class={cell}
					aria-label="Back to favorite pages"
					title="Favorite pages"
				>
					<Undo2 class="size-4.5" />
				</a>

				<button
					type="button"
					class={cell}
					disabled={at === 0}
					aria-label="Previous page"
					title="Previous page"
					onclick={goToPrev}
				>
					<ChevronLeft class="size-4.5" />
				</button>

				<p
					class="flex h-full shrink-0 items-center justify-center px-2 text-xs font-medium whitespace-nowrap tabular-nums sm:text-sm"
				>
					{at + 1} / {items.length}
				</p>

				<button
					type="button"
					class={cell}
					disabled={at >= items.length - 1}
					aria-label="Next page"
					title="Next page"
					onclick={goToNext}
				>
					<ChevronRight class="size-4.5" />
				</button>

				{#if current}
					<a
						href={resolve(`/a/${current.archive_id}`)}
						class="flex h-full max-w-32 min-w-0 items-center gap-1.5 px-2 text-xs font-medium hover:bg-accent hover:text-accent-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none focus-visible:ring-inset sm:max-w-56"
						title="Open {current.title} in the library"
					>
						<BookOpen class="size-4 shrink-0" />
						<span class="truncate">{current.title}</span>
						<span class="shrink-0 text-muted-foreground">p. {current.page_index + 1}</span>
					</a>

					<button
						type="button"
						class={cell}
						aria-label="Remove this page from favorites"
						title="Remove from favorites"
						aria-pressed={favorited}
						onclick={unfavoriteCurrent}
					>
						<Heart class="size-4.5 fill-red-500 stroke-red-500" />
					</button>
				{/if}

				<button
					type="button"
					class={cell}
					aria-label={playing ? 'Pause slideshow' : 'Play slideshow'}
					title={playing ? 'Pause slideshow' : 'Play slideshow'}
					onclick={togglePlay}
				>
					{#if playing}
						<Pause class="size-4.5" />
					{:else}
						<Play class="size-4.5" />
					{/if}
				</button>

				<Popover.Root bind:open={settingsOpen}>
					<Popover.Trigger>
						{#snippet child({ props })}
							<button
								{...props}
								class={cell}
								aria-label="Slideshow settings"
								title="Slideshow settings"
							>
								<Settings class="size-4.5" />
							</button>
						{/snippet}
					</Popover.Trigger>
					<Popover.Content side="top" align="end" class="w-64">
						<div class="flex flex-col gap-3">
							<div>
								<div
									class="mb-1.5 flex items-center justify-between text-xs font-medium text-muted-foreground"
								>
									<span>Interval</span>
									<span class="tabular-nums">{settings.slideshow_interval_seconds}s</span>
								</div>
								<Slider
									type="multiple"
									min={1}
									max={60}
									step={1}
									value={[settings.slideshow_interval_seconds]}
									onValueChange={(value: number[]) =>
										applyLocal({
											slideshow_interval_seconds: value[0]
										})}
									onValueCommit={(value: number[]) =>
										updateSettings({
											slideshow_interval_seconds: value[0]
										})}
								/>
							</div>

							<label class="flex items-center justify-between gap-3 text-sm">
								<span>Loop</span>
								<Switch
									checked={settings.slideshow_loop}
									onCheckedChange={(checked) => updateSettings({ slideshow_loop: checked })}
								/>
							</label>

							<label class="flex items-center justify-between gap-3 text-sm">
								<span class="flex items-center gap-1.5">
									<Shuffle class="size-4" />
									Shuffle
									<span class="text-xs text-muted-foreground">(this session only)</span>
								</span>
								<Switch checked={shuffle} onCheckedChange={toggleShuffle} />
							</label>

							<div>
								<p class="mb-1.5 text-xs font-medium text-muted-foreground">Fit</p>
								<div class="flex gap-1">
									{#each ['width', 'height', 'original'] as fit (fit)}
										<Button
											type="button"
											size="sm"
											variant={settings.fit_mode === fit ? 'secondary' : 'ghost'}
											class="flex-1"
											onclick={() => updateSettings({ fit_mode: fit as FitMode })}
										>
											{fit === 'width' ? 'Width' : fit === 'height' ? 'Height' : 'Original'}
										</Button>
									{/each}
								</div>
							</div>
						</div>
					</Popover.Content>
				</Popover.Root>
			</nav>

			{#if playing}
				<div
					class="pointer-events-none absolute inset-x-0 bottom-0 h-0.5 overflow-hidden rounded-b-2xl bg-muted"
					aria-hidden="true"
				>
					<div
						bind:this={progressBar}
						class="h-full origin-left bg-primary"
						style="transform: scaleX(0)"
					></div>
				</div>
			{/if}
		</div>
	</div>
</div>
