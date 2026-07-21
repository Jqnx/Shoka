<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import * as Pagination from '$lib/components/ui/pagination';
	import { Paintbrush } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	let { data } = $props();
</script>

<svelte:head>
	<title>Artists | Shoka</title>
</svelte:head>

<div class="mx-auto max-w-screen-2xl px-4 py-8 sm:px-6">
	<div class="mb-6 flex items-center gap-2">
		<Paintbrush class="size-5 text-muted-foreground" />
		<h1 class="text-2xl font-bold">Artists</h1>
		<span class="text-sm text-muted-foreground">({data.total})</span>
	</div>

	{#if data.artists.length === 0}
		<p class="text-muted-foreground">No artists found.</p>
	{:else}
		<div class="flex flex-wrap gap-2">
			{#each data.artists as artist (artist.id)}
				<Badge href={resolve(`/artist/${artist.id}`)} variant="outline" class="text-sm">
					{artist.name}
					<span class="text-muted-foreground">{artist.count}</span>
				</Badge>
			{/each}
		</div>
	{/if}

	{#if data.total > data.limit}
		<div class="mt-8">
			<Pagination.Root
				count={data.total}
				perPage={data.limit}
				page={data.page}
				onPageChange={(p) => goto(resolve(`/artist?page=${p}`), { noScroll: true })}
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
