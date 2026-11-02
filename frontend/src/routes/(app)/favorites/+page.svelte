<script lang="ts">
	import * as Pagination from '$lib/components/ui/pagination';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import ArchiveCard from '$lib/components/ArchiveCard.svelte';
	import AudioCard from '$lib/components/AudioCard.svelte';
	import FavoritePagesGrid from '$lib/components/favorites/FavoritePagesGrid.svelte';
	import { AudioLines, Search } from '@lucide/svelte';
	import { goto, replaceState, afterNavigate } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	let { data } = $props();

	type TabValue = 'archives' | 'audio' | 'pages';

	function tabFromUrl(): TabValue {
		const value = page.url.searchParams.get('tab');

		return value === 'audio' || value === 'pages' ? value : 'archives';
	}

	// The tab lives in the URL so reloads, bookmarks, and returning from a
	// work keep the list the address bar names.
	let tab = $state<TabValue>(tabFromUrl());

	// afterNavigate fires for real navigations (including Back/Forward) but not
	// for the replaceState used by tab clicks, so re-reading here keeps the
	// visible tab aligned with the URL without undoing a tab switch.
	afterNavigate((nav) => {
		if (nav.type === 'goto') window.scrollTo({ top: 0, behavior: 'smooth' });
		tab = tabFromUrl();
	});

	function setTab(next: string) {
		tab = next === 'audio' || next === 'pages' ? next : 'archives';

		const params = new URLSearchParams(page.url.searchParams);
		if (tab === 'archives') params.delete('tab');
		else params.set('tab', tab);

		const query = params.toString();
		replaceState(resolve(`/favorites?${query}`), {});
	}

	// Archives and audio paginate independently: they share the page URL, so
	// each tab carries its own page parameter and remembers itself there.
	function goToPage(param: 'page' | 'audio_page', value: number) {
		const params = new URLSearchParams(page.url.searchParams);
		params.set(param, String(value));
		params.set('tab', param === 'audio_page' ? 'audio' : 'archives');
		goto(resolve(`${page.url.pathname}?${params}`), { noScroll: true });
	}
</script>

<svelte:head>
	<title>Favorites | Shoka</title>
</svelte:head>

{#snippet pagination(count: number, perPage: number, current: number, param: 'page' | 'audio_page')}
	{#if count > perPage}
		<div class="mt-8">
			<Pagination.Root {count} {perPage} page={current} onPageChange={(p) => goToPage(param, p)}>
				{#snippet children({ pages })}
					<Pagination.Content>
						<Pagination.Item>
							<Pagination.Previous />
						</Pagination.Item>
						{#each pages as p (p.key)}
							{#if p.type === 'ellipsis'}
								<Pagination.Item>
									<Pagination.Ellipsis />
								</Pagination.Item>
							{:else}
								<Pagination.Item>
									<Pagination.Link page={p} isActive={current === p.value} />
								</Pagination.Item>
							{/if}
						{/each}
						<Pagination.Item>
							<Pagination.Next />
						</Pagination.Item>
					</Pagination.Content>
				{/snippet}
			</Pagination.Root>
		</div>
	{/if}
{/snippet}

<Sidebar.Provider open={false}>
	<Sidebar.Inset class="bg-transparent">
		<div class="flex flex-wrap items-center justify-between gap-2 px-4 py-3 sm:px-6">
			<h1 class="text-base font-semibold">Favorites</h1>
		</div>

		<div class="px-4 py-4 sm:px-6">
			<Tabs.Root bind:value={tab} onValueChange={setTab} class="gap-6">
				<Tabs.List class="self-center">
					<Tabs.Trigger value="archives">Archives</Tabs.Trigger>
					<Tabs.Trigger value="audio">Audio</Tabs.Trigger>
					<Tabs.Trigger value="pages">Pages</Tabs.Trigger>
				</Tabs.List>

				<Tabs.Content value="archives" class="text-sm">
					{#if data.archives.length === 0}
						<div class="flex flex-col items-center justify-center py-24 text-center">
							<Search class="mb-3 size-10 text-muted-foreground/40" />
							<p class="text-base font-medium text-foreground">No favorites yet</p>
						</div>
					{:else}
						<div
							class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8"
						>
							{#each data.archives as archive (archive.id)}
								<ArchiveCard {archive} />
							{/each}
						</div>
					{/if}

					{@render pagination(data.total, data.limit, data.page, 'page')}
				</Tabs.Content>

				<Tabs.Content value="audio" class="text-sm">
					{#if data.audioTotal === 0}
						<div class="flex flex-col items-center justify-center py-24 text-center">
							<AudioLines class="mb-3 size-10 text-muted-foreground/40" />
							<p class="text-base font-medium text-foreground">No favorite audio yet</p>
						</div>
					{:else}
						<div
							class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7"
						>
							{#each data.audio as work (work.id)}
								<AudioCard audio={work} />
							{/each}
						</div>
					{/if}

					{@render pagination(data.audioTotal, data.audioLimit, data.audioPage, 'audio_page')}
				</Tabs.Content>

				<Tabs.Content value="pages" class="text-sm">
					<!-- bits-ui keeps every tab panel mounted and only hides the
					     inactive ones, so the grid is gated explicitly to stay
					     lazy: no fetch on page load, a fresh one per activation. -->
					{#if tab === 'pages'}
						<FavoritePagesGrid />
					{/if}
				</Tabs.Content>
			</Tabs.Root>
		</div>
	</Sidebar.Inset>
</Sidebar.Provider>
