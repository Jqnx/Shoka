<script lang="ts">
	import * as Pagination from '$lib/components/ui/pagination';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import AudioCard from '$lib/components/AudioCard.svelte';
	import AudioSidebar from '$lib/components/AudioSidebar.svelte';
	import FiltersTrigger from '$lib/components/FiltersTrigger.svelte';
	import { AudioLines } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	let { data } = $props();
</script>

<svelte:head>
	<title>{data.library.name} | Shoka</title>
</svelte:head>

<Sidebar.Provider open={false}>
	<Sidebar.Inset class="bg-transparent">
		<div class="flex flex-wrap items-center justify-between gap-2 px-4 py-3 sm:px-6">
			<div>
				<h1 class="text-base font-semibold">{data.library.name}</h1>
				<p class="text-sm text-muted-foreground">
					{data.total}
					{data.total === 1 ? 'work' : 'works'} found
				</p>
			</div>

			<div class="flex items-center gap-2">
				<FiltersTrigger />
			</div>
		</div>

		<div class="px-4 py-4 sm:px-6">
			{#if data.audio.length === 0}
				<div class="flex flex-col items-center justify-center py-24 text-center">
					<AudioLines class="mb-3 size-10 text-muted-foreground/40" />
					<p class="text-base font-medium text-foreground">No audio found</p>
				</div>
			{:else}
				<div
					class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8"
				>
					{#each data.audio as work (work.id)}
						<AudioCard audio={work} />
					{/each}
				</div>
			{/if}

			{#if data.total > data.limit}
				<div class="mt-8">
					<Pagination.Root
						count={data.total}
						perPage={data.limit}
						page={data.page}
						onPageChange={(p) => {
							const params = new URLSearchParams(page.url.searchParams);
							params.set('page', String(p));
							goto(resolve(`${page.url.pathname}?${params}`), { noScroll: true });
						}}
					>
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
											<Pagination.Link page={p} isActive={data.page === p.value} />
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
		</div>
	</Sidebar.Inset>

	<AudioSidebar
		filters={data.filters}
		artists={data.artists}
		tags={data.tags}
		circles={data.allCircles}
		voiceActors={data.voiceActors}
	/>
</Sidebar.Provider>
