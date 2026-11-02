<script lang="ts">
	import { resolve } from '$app/paths';
	import * as Carousel from '$lib/components/ui/carousel';
	import ArchiveCard from '$lib/components/ArchiveCard.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { can } from '$lib/auth/permissions';
	import { firstArchiveLibrary } from '$lib/types';
	import { Clock, BookPlus, CalendarDays } from '@lucide/svelte';

	let { data } = $props();

	const recentlyRead = $derived(data.recentlyRead);
	const recentlyAdded = $derived(data.recentlyAdded);
	const recentlyReleased = $derived(data.recentlyReleased);

	// Same "default to the first doujinshi library" fallback as the carousels'
	// own data.
	const archiveLibrary = $derived(firstArchiveLibrary(data.libraries));
	const archivesHref = $derived((archiveLibrary ? `/${archiveLibrary.id}` : '/') as `/${string}`);

	const carouselOpts = { align: 'start' as const, dragFree: true, loop: true };
</script>

<svelte:head>
	<title>Shoka</title>
</svelte:head>

<div class="mx-auto max-w-screen-2xl space-y-10 px-4 py-8 sm:px-6">
	<!-- Recently Read -->
	{#if recentlyRead.length > 0}
		<section>
			<div class="mb-3 flex items-center justify-between">
				<div class="flex items-center gap-2">
					<Clock class="size-4 text-muted-foreground" />
					<h2 class="text-sm font-semibold">Recently Read</h2>
				</div>
				<a
					href={resolve(archivesHref)}
					class="text-xs text-muted-foreground transition-colors hover:text-foreground"
				>
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
				<Carousel.Previous
					variant="secondary"
					size="icon-lg"
					class="-left-4 shadow-md disabled:hidden"
				/>
				<Carousel.Next
					variant="secondary"
					size="icon-lg"
					class="-right-4 shadow-md disabled:hidden"
				/>
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
				<a
					href={resolve(`${archivesHref}?sort=created_at_desc`)}
					class="text-xs text-muted-foreground transition-colors hover:text-foreground"
				>
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
				<Carousel.Previous
					variant="secondary"
					size="icon-lg"
					class="-left-4 shadow-md disabled:hidden"
				/>
				<Carousel.Next
					variant="secondary"
					size="icon-lg"
					class="-right-4 shadow-md disabled:hidden"
				/>
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
				<a
					href={resolve(`${archivesHref}?sort=release_date_desc`)}
					class="text-xs text-muted-foreground transition-colors hover:text-foreground"
				>
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
				<Carousel.Previous
					variant="secondary"
					size="icon-lg"
					class="-left-4 shadow-md disabled:hidden"
				/>
				<Carousel.Next
					variant="secondary"
					size="icon-lg"
					class="-right-4 shadow-md disabled:hidden"
				/>
			</Carousel.Root>
		</section>
	{/if}

	{#if recentlyRead.length === 0 && recentlyAdded.length === 0 && recentlyReleased.length === 0}
		<div class="flex flex-col items-center justify-center py-24 text-center">
			{#if !archiveLibrary && data.libraries.length > 0}
				<p class="text-base font-medium text-foreground">No archives yet</p>
				<p class="mt-1 text-sm text-muted-foreground">
					Your audio libraries are available in the Audio section.
				</p>
				<Button class="mt-4" href={resolve('/audio')}>Open audio</Button>
			{:else}
				<p class="text-base font-medium text-foreground">Your library is empty</p>
				<p class="mt-1 text-sm text-muted-foreground">Add some archives to get started</p>
				{#if data.libraries.length === 0 && can(data.user, { library: ['create'] })}
					<Button class="mt-4" href={resolve('/setup')}>Add your first library</Button>
				{/if}
			{/if}
		</div>
	{/if}
</div>
