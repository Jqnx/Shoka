<script lang="ts">
	import { resolve } from '$app/paths';
	import { goto, afterNavigate } from '$app/navigation';

	afterNavigate((nav) => {
		if (nav.type === 'goto') window.scrollTo({ top: 0, behavior: 'smooth' });
	});
	import * as Pagination from '$lib/components/ui/pagination';
	import { ChevronLeft, Sword } from '@lucide/svelte';
	import ArchiveCard from '$lib/components/ArchiveCard.svelte';

	let { data } = $props();

	// The tag/character/parody-by-name endpoints aren't library-scoped, so
	// there's no library context here; default to the first one, same as
	// the rest of the app until there's a "current library" concept.
	const archivesHref = $derived(data.libraries[0] ? `/${data.libraries[0].id}` : '/');
</script>

<svelte:head>
	<title>{data.name} | Shoka</title>
</svelte:head>

<div class="mx-auto max-w-screen-2xl px-4 py-8 sm:px-6">
	<a
		href={archivesHref}
		class="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
	>
		<ChevronLeft class="size-4" />
		Archives
	</a>

	<div class="mb-6 flex items-center gap-2">
		<Sword class="size-5 text-muted-foreground" />
		<h1 class="text-2xl font-bold">{data.name}</h1>
		<span class="text-sm text-muted-foreground">({data.total})</span>
	</div>

	{#if data.archives.length === 0}
		<p class="text-muted-foreground">No archives found.</p>
	{:else}
		<div
			class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8"
		>
			{#each data.archives as archive (archive.id)}
				<ArchiveCard {archive} />
			{/each}
		</div>
	{/if}

	{#if data.total > data.limit}
		<div class="mt-8">
			<Pagination.Root
				count={data.total}
				perPage={data.limit}
				page={data.page}
				onPageChange={(p) => goto(resolve(`/character/${encodeURIComponent(data.name)}?page=${p}`), { noScroll: true })}
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
