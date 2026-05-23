<script lang="ts">
	import * as Carousel from '$lib/components/ui/carousel';
	import ArchiveCard from '$lib/components/ArchiveCard.svelte';
	import { Clock, BookPlus, CalendarDays } from '@lucide/svelte';

	let { data } = $props();

	const recentlyRead = $derived(
		data.archives
			.filter((a) => a.progress !== null)
			.sort((a, b) => {
				const da = a.progress ? new Date(a.progress.last_read).getTime() : 0;
				const db = b.progress ? new Date(b.progress.last_read).getTime() : 0;
				return db - da;
			})
			.slice(0, 10)
	);

	const recentlyAdded = $derived(
		[...data.archives]
			.sort((a, b) => b.created_at.localeCompare(a.created_at))
			.slice(0, 10)
	);

	const recentlyReleased = $derived(
		data.archives
			.filter((a) => a.release_date !== null)
			.sort((a, b) => {
				const da = new Date(a.release_date!).getTime();
				const db = new Date(b.release_date!).getTime();
				return db - da;
			})
			.slice(0, 10)
	);

	const carouselOpts = { align: 'start' as const, dragFree: true, loop: true };
</script>

<div class="mx-auto max-w-screen-2xl space-y-10 px-4 py-8 sm:px-6">
	<!-- Recently Read -->
	{#if recentlyRead.length > 0}
		<section>
			<div class="mb-3 flex items-center justify-between">
				<div class="flex items-center gap-2">
					<Clock class="size-4 text-muted-foreground" />
					<h2 class="text-sm font-semibold">Recently Read</h2>
				</div>
				<a href="/a" class="text-xs text-muted-foreground transition-colors hover:text-foreground">
					View all →
				</a>
			</div>
			<Carousel.Root opts={carouselOpts}>
				<Carousel.Content class="-ml-3">
					{#each recentlyRead as archive (archive.id)}
						<Carousel.Item class="basis-[220px] pl-3">
							<ArchiveCard {archive} />
						</Carousel.Item>
					{/each}
				</Carousel.Content>
				<Carousel.Previous class="-left-4" />
				<Carousel.Next class="-right-4" />
			</Carousel.Root>
		</section>
	{/if}

	<!-- Recently Added -->
	{#if recentlyAdded.length > 0}
		<section>
			<div class="mb-3 flex items-center justify-between">
				<div class="flex items-center gap-2">
					<BookPlus class="size-4 text-muted-foreground" />
					<h2 class="text-sm font-semibold">Recently Added</h2>
				</div>
				<a href="/a" class="text-xs text-muted-foreground transition-colors hover:text-foreground">
					View all →
				</a>
			</div>
			<Carousel.Root opts={carouselOpts}>
				<Carousel.Content class="-ml-3">
					{#each recentlyAdded as archive (archive.id)}
						<Carousel.Item class="basis-[220px] pl-3">
							<ArchiveCard {archive} />
						</Carousel.Item>
					{/each}
				</Carousel.Content>
				<Carousel.Previous class="-left-4" />
				<Carousel.Next class="-right-4" />
			</Carousel.Root>
		</section>
	{/if}

	<!-- Recently Released -->
	{#if recentlyReleased.length > 0}
		<section>
			<div class="mb-3 flex items-center justify-between">
				<div class="flex items-center gap-2">
					<CalendarDays class="size-4 text-muted-foreground" />
					<h2 class="text-sm font-semibold">Recently Released</h2>
				</div>
				<a href="/a" class="text-xs text-muted-foreground transition-colors hover:text-foreground">
					View all →
				</a>
			</div>
			<Carousel.Root opts={carouselOpts}>
				<Carousel.Content class="-ml-3">
					{#each recentlyReleased as archive (archive.id)}
						<Carousel.Item class="basis-[220px] pl-3">
							<ArchiveCard {archive} />
						</Carousel.Item>
					{/each}
				</Carousel.Content>
				<Carousel.Previous class="-left-4" />
				<Carousel.Next class="-right-4" />
			</Carousel.Root>
		</section>
	{/if}

	{#if recentlyRead.length === 0 && recentlyAdded.length === 0 && recentlyReleased.length === 0}
		<div class="flex flex-col items-center justify-center py-24 text-center">
			<p class="text-base font-medium text-foreground">Your library is empty</p>
			<p class="mt-1 text-sm text-muted-foreground">Add some archives to get started</p>
		</div>
	{/if}
</div>
