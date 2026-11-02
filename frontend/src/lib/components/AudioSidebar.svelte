<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Separator } from '$lib/components/ui/separator';
	import { MultiCombobox } from '$lib/components/ui/combobox';
	import { cn } from '$lib/utils.js';
	import { Heart, Star, X } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import type { Artist, Circle, Tag, VoiceActor } from '$lib/types';

	type Filters = {
		q: string;
		favorites: string;
		rating: string;
		tag: string[];
		artist: string[];
		circle: string[];
		voice_actor: string[];
	};

	let {
		filters,
		artists,
		tags,
		circles,
		voiceActors
	}: {
		filters: Filters;
		artists: Artist[];
		tags: Tag[];
		circles: Circle[];
		voiceActors: VoiceActor[];
	} = $props();

	let search = $derived(filters.q);

	const favorites = $derived(filters.favorites === 'true');
	const rating = $derived(Number(filters.rating) || 0);
	const hasActiveFilters = $derived(
		filters.q !== '' ||
			filters.favorites !== '' ||
			filters.rating !== '' ||
			filters.tag.length > 0 ||
			filters.artist.length > 0 ||
			filters.circle.length > 0 ||
			filters.voice_actor.length > 0
	);

	const ratingLabel = $derived(
		rating > 0 ? `${rating} star${rating === 1 ? '' : 's'}` : 'Any rating'
	);
	let hoveredStar = $state<number | null>(null);

	const artistOptions = $derived(
		artists.map((artist) => ({ value: artist.name, label: artist.name }))
	);
	const tagOptions = $derived(tags.map((tag) => ({ value: tag.name, label: tag.name })));
	const circleOptions = $derived(
		circles.map((circle) => ({ value: circle.name, label: circle.name }))
	);
	const voiceActorOptions = $derived(
		voiceActors.map((voiceActor) => ({ value: voiceActor.name, label: voiceActor.name }))
	);

	function setParam(name: string, value: string) {
		const params = new URLSearchParams(page.url.searchParams);
		if (value) params.set(name, value);
		else params.delete(name);
		params.delete('page');
		goto(resolve(`${page.url.pathname}?${params}`), { noScroll: true, keepFocus: true });
	}

	function setMultiParam(name: string, values: string[]) {
		const params = new URLSearchParams(page.url.searchParams);
		params.delete(name);
		for (const value of values) {
			if (value) params.append(name, value);
		}
		params.delete('page');
		goto(resolve(`${page.url.pathname}?${params}`), { noScroll: true, keepFocus: true });
	}

	function submitSearch(event: SubmitEvent) {
		event.preventDefault();
		setParam('q', search.trim());
	}

	function clearFilters() {
		const params = new URLSearchParams(page.url.searchParams);
		for (const key of [
			'q',
			'favorites',
			'rating',
			'tag',
			'artist',
			'circle',
			'voice_actor',
			'page'
		]) {
			params.delete(key);
		}
		goto(resolve(`${page.url.pathname}?${params}`), { noScroll: true });
	}
</script>

<Sidebar.Root side="right" collapsible="offcanvas" class="top-14 h-[calc(100svh-3.5rem)]">
	{#if hasActiveFilters}
		<Sidebar.Header class="p-3">
			<Button
				variant="ghost"
				size="sm"
				onclick={clearFilters}
				class="justify-start gap-1.5 text-muted-foreground"
			>
				<X class="size-3.5" />
				Clear filters
			</Button>
		</Sidebar.Header>
	{/if}

	<Sidebar.Content class="gap-4 px-3 py-2">
		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Search</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<form onsubmit={submitSearch}>
					<Input bind:value={search} placeholder="Search audio…" aria-label="Search audio" />
				</form>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Separator />

		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Artists</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<MultiCombobox
					options={artistOptions}
					value={filters.artist}
					onValueChange={(v) => setMultiParam('artist', v)}
					placeholder="Add artist..."
					searchPlaceholder="Search artists..."
					emptyText="No artists found."
				/>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Separator />

		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Tags</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<MultiCombobox
					options={tagOptions}
					value={filters.tag}
					onValueChange={(v) => setMultiParam('tag', v)}
					placeholder="Add tag..."
					searchPlaceholder="Search tags..."
					emptyText="No tags found."
				/>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Separator />

		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Circles</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<MultiCombobox
					options={circleOptions}
					value={filters.circle}
					onValueChange={(v) => setMultiParam('circle', v)}
					placeholder="Add circle..."
					searchPlaceholder="Search circles..."
					emptyText="No circles found."
				/>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Separator />

		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Voice actors</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<MultiCombobox
					options={voiceActorOptions}
					value={filters.voice_actor}
					onValueChange={(v) => setMultiParam('voice_actor', v)}
					placeholder="Add voice actor..."
					searchPlaceholder="Search voice actors..."
					emptyText="No voice actors found."
				/>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Separator />

		<Sidebar.Group class="p-0">
			<Sidebar.GroupLabel class="px-0">Favorites</Sidebar.GroupLabel>
			<Sidebar.GroupContent class="mt-1">
				<Button
					variant={favorites ? 'default' : 'outline'}
					size="sm"
					class="w-full justify-start gap-1.5"
					aria-pressed={favorites}
					onclick={() => setParam('favorites', favorites ? '' : 'true')}
				>
					<Heart class={cn('size-3.5', favorites && 'fill-current')} />
					Favorites only
				</Button>
			</Sidebar.GroupContent>
		</Sidebar.Group>

		<Separator />

		<Sidebar.Group class="p-0">
			<div class="flex items-center justify-between">
				<Sidebar.GroupLabel class="px-0">Rating</Sidebar.GroupLabel>
				<span class="text-xs text-muted-foreground">{ratingLabel}</span>
			</div>
			<Sidebar.GroupContent class="mt-1">
				<div
					role="group"
					aria-label="Rating filter"
					class="flex items-center justify-center gap-1"
					onmouseleave={() => (hoveredStar = null)}
				>
					{#each [1, 2, 3, 4, 5] as n (n)}
						<button
							type="button"
							class="p-0.5 text-muted-foreground transition-colors hover:text-primary"
							aria-label={rating === n
								? `Clear ${n}-star rating filter`
								: `Filter by ${n} star${n === 1 ? '' : 's'}`}
							aria-pressed={rating === n}
							onmouseenter={() => (hoveredStar = n)}
							onclick={() => setParam('rating', rating === n ? '' : String(n))}
						>
							<Star
								class={cn(
									'size-5',
									(hoveredStar ?? rating ?? 0) >= n ? 'fill-primary stroke-primary' : 'fill-none'
								)}
							/>
						</button>
					{/each}
				</div>
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>
</Sidebar.Root>
