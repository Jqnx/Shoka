<script lang="ts">
	import { resolve } from '$app/paths';
	import { afterNavigate, replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import * as Carousel from '$lib/components/ui/carousel';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import ArchiveCard from '$lib/components/ArchiveCard.svelte';
	import AudioCard from '$lib/components/AudioCard.svelte';
	import HomeCarousel from '$lib/components/HomeCarousel.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { can } from '$lib/auth/permissions';
	import { firstArchiveLibrary } from '$lib/types';
	import { BookPlus, CalendarDays, CirclePlay, Clock } from '@lucide/svelte';

	let { data } = $props();

	const recentlyRead = $derived(data.recentlyRead);
	const recentlyAdded = $derived(data.recentlyAdded);
	const recentlyReleased = $derived(data.recentlyReleased);
	const audioRecentlyPlayed = $derived(data.audioRecentlyPlayed);
	const audioRecentlyAdded = $derived(data.audioRecentlyAdded);
	const audioRecentlyReleased = $derived(data.audioRecentlyReleased);

	// Same "default to the first doujinshi library" fallback as the carousels'
	// own data.
	const archiveLibrary = $derived(firstArchiveLibrary(data.libraries));
	const archivesHref = $derived((archiveLibrary ? `/${archiveLibrary.id}` : '/') as `/${string}`);
	const audioLibrary = $derived(data.libraries.find((library) => library.type === 'audio'));
	const audioHref = $derived((audioLibrary ? `/${audioLibrary.id}` : '/') as `/${string}`);

	const archivesEmpty = $derived(
		recentlyRead.length === 0 && recentlyAdded.length === 0 && recentlyReleased.length === 0
	);
	const audioEmpty = $derived(
		audioRecentlyPlayed.length === 0 &&
			audioRecentlyAdded.length === 0 &&
			audioRecentlyReleased.length === 0
	);

	type TabValue = 'archives' | 'audio';

	function tabFromUrl(): TabValue {
		return page.url.searchParams.get('tab') === 'audio' ? 'audio' : 'archives';
	}

	// replaceState updates page.state, not page.url, so a tab clicked in this
	// session must be read back from state - the URL still describes the
	// page that was rendered.
	function currentTab(): TabValue {
		const stored = page.state?.tab;

		return stored === 'audio' || stored === 'archives' ? stored : tabFromUrl();
	}

	// The tab lives in the URL so reloads, bookmarks, and returning from a
	// work keep the address bar naming what is visible.
	let tab = $state<TabValue>(currentTab());

	// afterNavigate fires for real navigations (including Back/Forward) but not
	// for the replaceState used by tab clicks.
	afterNavigate((nav) => {
		if (nav.type === 'goto') window.scrollTo({ top: 0, behavior: 'smooth' });
		tab = currentTab();
	});

	function setTab(next: string) {
		tab = next === 'audio' ? 'audio' : 'archives';

		const params = new URLSearchParams(window.location.search);
		if (tab === 'archives') params.delete('tab');
		else params.set('tab', tab);

		const query = params.toString();
		replaceState(resolve(`/?${query}`), { tab });
	}
</script>

<svelte:head>
	<title>Shoka</title>
</svelte:head>

{#snippet archivesContent()}
	<div class="space-y-10">
		{#if recentlyRead.length > 0}
			<HomeCarousel title="Recently Read" href={archivesHref} icon={Clock}>
				{#each recentlyRead as archive (archive.id)}
					<Carousel.Item class="basis-[220px] pl-3">
						<ArchiveCard {archive} />
					</Carousel.Item>
				{/each}
			</HomeCarousel>
		{/if}

		{#if recentlyAdded.length > 0}
			<HomeCarousel
				title="Recently Added"
				href={`${archivesHref}?sort=created_at_desc`}
				icon={BookPlus}
			>
				{#each recentlyAdded as archive (archive.id)}
					<Carousel.Item class="basis-[220px] pl-3">
						<ArchiveCard {archive} />
					</Carousel.Item>
				{/each}
			</HomeCarousel>
		{/if}

		{#if recentlyReleased.length > 0}
			<HomeCarousel
				title="Recently Released"
				href={`${archivesHref}?sort=release_date_desc`}
				icon={CalendarDays}
			>
				{#each recentlyReleased as archive (archive.id)}
					<Carousel.Item class="basis-[220px] pl-3">
						<ArchiveCard {archive} />
					</Carousel.Item>
				{/each}
			</HomeCarousel>
		{/if}
	</div>
{/snippet}

{#snippet audioContent()}
	<div class="space-y-10">
		{#if audioRecentlyPlayed.length > 0}
			<HomeCarousel title="Recently Played" href={audioHref} icon={CirclePlay}>
				{#each audioRecentlyPlayed as work (work.id)}
					<Carousel.Item class="basis-[calc(100vw_-_3rem)] pl-3 sm:basis-[372px]">
						<AudioCard audio={work} />
					</Carousel.Item>
				{/each}
			</HomeCarousel>
		{/if}

		{#if audioRecentlyAdded.length > 0}
			<HomeCarousel title="Recently Added" href={audioHref} icon={BookPlus}>
				{#each audioRecentlyAdded as work (work.id)}
					<Carousel.Item class="basis-[calc(100vw_-_3rem)] pl-3 sm:basis-[372px]">
						<AudioCard audio={work} />
					</Carousel.Item>
				{/each}
			</HomeCarousel>
		{/if}

		{#if audioRecentlyReleased.length > 0}
			<HomeCarousel title="Recently Released" href={audioHref} icon={CalendarDays}>
				{#each audioRecentlyReleased as work (work.id)}
					<Carousel.Item class="basis-[calc(100vw_-_3rem)] pl-3 sm:basis-[372px]">
						<AudioCard audio={work} />
					</Carousel.Item>
				{/each}
			</HomeCarousel>
		{/if}
	</div>
{/snippet}

<div class="mx-auto max-w-screen-2xl px-4 py-8 sm:px-6">
	{#if !archivesEmpty && !audioEmpty}
		<Tabs.Root bind:value={tab} onValueChange={setTab} class="gap-8">
			<Tabs.List class="self-center">
				<Tabs.Trigger value="archives">Archives</Tabs.Trigger>
				<Tabs.Trigger value="audio">Voice Works</Tabs.Trigger>
			</Tabs.List>

			<Tabs.Content value="archives">
				{@render archivesContent()}
			</Tabs.Content>

			<Tabs.Content value="audio">
				{@render audioContent()}
			</Tabs.Content>
		</Tabs.Root>
	{:else if !archivesEmpty}
		{@render archivesContent()}
	{:else if !audioEmpty}
		{@render audioContent()}
	{:else}
		<div class="flex flex-col items-center justify-center py-24 text-center">
			<p class="text-base font-medium text-foreground">Your library is empty</p>
			<p class="mt-1 text-sm text-muted-foreground">Add a library and scan it to get started</p>
			{#if data.libraries.length === 0 && can(data.user, { library: ['create'] })}
				<Button class="mt-4" href={resolve('/setup')}>Add your first library</Button>
			{/if}
		</div>
	{/if}
</div>
