<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import { cn } from '$lib/utils.js';
	import { Heart, Star, X } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	let { favorites, rating }: { favorites: boolean; rating: number } = $props();

	const hasActiveFilters = $derived(favorites || rating > 0);

	const ratingLabel = $derived(
		rating > 0 ? `${rating} star${rating === 1 ? '' : 's'}` : 'Any rating'
	);
	// Hover previews a star count without changing the filter.
	let hoveredStar = $state<number | null>(null);

	function setParam(name: string, value: string) {
		const params = new URLSearchParams(page.url.searchParams);
		if (value) params.set(name, value);
		else params.delete(name);
		params.delete('page');
		goto(resolve(`${page.url.pathname}?${params}`), { noScroll: true, keepFocus: true });
	}

	function clearFilters() {
		const params = new URLSearchParams(page.url.searchParams);
		for (const key of ['favorites', 'rating', 'page']) {
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
