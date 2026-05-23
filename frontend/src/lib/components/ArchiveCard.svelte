<script lang="ts">
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import type { Archive } from '$lib/types';

	let { archive }: { archive: Archive } = $props();
</script>

<a
	href={resolve(`/a/${archive.id}`)}
	class="group relative flex cursor-pointer flex-col overflow-hidden rounded-lg border border-border bg-card transition-all hover:-translate-y-0.5 hover:shadow-md"
>
	<!-- Cover -->
	<div class="relative aspect-[2/3] w-full overflow-hidden bg-muted">
		<img
			src="http://localhost:8080/api/archives/{archive.id}/cover"
			alt={archive.title}
			class="absolute inset-0 size-full object-cover"
		/>
		<div
			class="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-black/80 via-black/30 to-transparent opacity-0 transition-opacity group-hover:opacity-100"
		></div>

		<div class="absolute top-1.5 right-1.5">
			<span class="rounded bg-black/60 px-1.5 py-0.5 text-[10px] font-medium text-white/90">
				{archive.page_count}p
			</span>
		</div>

		<div
			class="absolute inset-x-0 bottom-0 translate-y-1 p-2 opacity-0 transition-all group-hover:translate-y-0 group-hover:opacity-100"
		>
			<p class="line-clamp-2 text-xs leading-tight font-semibold text-white">{archive.title}</p>
		</div>
	</div>

	<!-- Card body -->
	<div class="flex flex-col gap-1 p-2">
		<p
			class="line-clamp-2 text-xs leading-tight font-medium text-card-foreground group-hover:text-foreground"
		>
			{archive.title}
		</p>
		{#if archive.artists?.length}
			<p class="truncate text-[11px] text-muted-foreground">{archive.artists.join(', ')}</p>
		{/if}
		<div class="mt-0.5 flex flex-wrap gap-1">
			{#if archive.category}
				<Badge variant="secondary" class="h-4 px-1.5 text-[10px]">{archive.category}</Badge>
			{/if}
			{#if archive.language}
				<Badge variant="outline" class="h-4 px-1.5 text-[10px]">{archive.language}</Badge>
			{/if}
		</div>
	</div>
</a>
