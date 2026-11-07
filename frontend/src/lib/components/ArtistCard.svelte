<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { enhance } from '$app/forms';
	import ArchiveCover from '$lib/components/ArchiveCover.svelte';
	import { cn } from '$lib/utils.js';
	import { BookOpen, Heart, Paintbrush } from '@lucide/svelte';
	import type { ArtistListItem } from '$lib/types';

	let { artist }: { artist: ArtistListItem } = $props();

	// Same as ArchiveCard: carry the current list page through to the detail
	// page so its "back" link returns here with the filters intact.
	const from = $derived(page.url.pathname + page.url.search);
	const href = $derived(resolve(`/artist/${artist.id}?from=${encodeURIComponent(from)}`));
</script>

<div
	class="group relative flex h-full flex-col overflow-hidden rounded-lg border border-border bg-card transition-all hover:-translate-y-0.5 hover:shadow-md"
>
	<a {href} data-sveltekit-preload-data="tap" class="flex flex-1 flex-col">
		<div class="relative aspect-[2/3] w-full overflow-hidden bg-muted">
			{#if artist.cover}
				<ArchiveCover
					archive={{
						id: artist.cover.archive_id,
						cover_page: artist.cover.page,
						thumbs_ready: artist.cover.thumbs_ready
					}}
					alt={artist.name}
					class="absolute inset-0 size-full object-cover"
				/>
			{:else}
				<div class="absolute inset-0 flex items-center justify-center">
					<Paintbrush class="size-8 text-muted-foreground/40" />
				</div>
			{/if}
		</div>
		<div class="flex flex-col gap-1 p-2">
			<p
				class="line-clamp-2 text-xs leading-tight font-medium text-card-foreground group-hover:text-foreground"
			>
				{artist.name}
			</p>
			<p class="flex items-center gap-1 text-xs text-muted-foreground">
				<BookOpen class="size-3" />
				{artist.count}
			</p>
		</div>
	</a>

	<!-- Sibling of the link, not a child: a button nested in an anchor is
	     invalid HTML and clicks on it would also navigate. -->
	<form method="POST" action="?/toggleFavorite" use:enhance class="absolute top-1.5 right-1.5">
		<input type="hidden" name="id" value={artist.id} />
		<input type="hidden" name="favorited" value={artist.is_favorited} />
		<button
			type="submit"
			aria-label={artist.is_favorited ? 'Remove from favorites' : 'Add to favorites'}
			aria-pressed={artist.is_favorited}
			class="flex size-7 items-center justify-center rounded-full bg-black/50 text-white/90 transition-colors hover:bg-black/70"
		>
			<Heart class={cn('size-4', artist.is_favorited && 'fill-red-500 stroke-red-500')} />
		</button>
	</form>
</div>
