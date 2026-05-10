<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';

	type Props = {
		id: number;
		title: string;
		artist: string;
		category: string;
		language: string;
		pages: number;
		coverHue: number;
	};

	let { id, title, artist, category, language, pages, coverHue }: Props = $props();
</script>

<a
	href={`/archive/${id}`}
	class="group relative flex cursor-pointer flex-col overflow-hidden rounded-lg border border-border bg-card transition-all hover:-translate-y-0.5 hover:shadow-md"
>
	<!-- Cover (2:3 portrait ratio placeholder) -->
	<div
		class="relative aspect-[2/3] w-full overflow-hidden"
		style="background: oklch(0.55 0.12 {coverHue})"
	>
		<div
			class="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-black/80 via-black/30 to-transparent opacity-0 transition-opacity group-hover:opacity-100"
		></div>

		<div class="absolute top-1.5 right-1.5">
			<span class="rounded bg-black/60 px-1.5 py-0.5 text-[10px] font-medium text-white/90">
				{pages}p
			</span>
		</div>

		<div
			class="absolute inset-x-0 bottom-0 translate-y-1 p-2 opacity-0 transition-all group-hover:translate-y-0 group-hover:opacity-100"
		>
			<p class="line-clamp-2 text-xs font-semibold leading-tight text-white">{title}</p>
		</div>
	</div>

	<!-- Card body -->
	<div class="flex flex-col gap-1 p-2">
		<p class="line-clamp-2 text-xs font-medium leading-tight text-card-foreground group-hover:text-foreground">
			{title}
		</p>
		<p class="truncate text-[11px] text-muted-foreground">{artist}</p>
		<div class="mt-0.5 flex flex-wrap gap-1">
			<Badge variant="secondary" class="h-4 px-1.5 text-[10px]">{category}</Badge>
			<Badge variant="outline" class="h-4 px-1.5 text-[10px]">{language}</Badge>
		</div>
	</div>
</a>
