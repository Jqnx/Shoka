<script lang="ts">
	import * as Carousel from '$lib/components/ui/carousel';
	import { resolve } from '$app/paths';
	import type { Component, Snippet } from 'svelte';

	let {
		title,
		href,
		icon: Icon,
		children
	}: {
		title: string;
		href: string;
		icon: Component;
		children: Snippet;
	} = $props();

	const carouselOpts = { align: 'start' as const, dragFree: true, loop: true };
</script>

<section>
	<div class="mb-3 flex items-center justify-between">
		<div class="flex items-center gap-2">
			<Icon class="size-4 text-muted-foreground" />
			<h2 class="text-sm font-semibold">{title}</h2>
		</div>
		<a
			href={resolve(href as `/${string}`)}
			class="text-xs text-muted-foreground transition-colors hover:text-foreground"
		>
			View all →
		</a>
	</div>
	<Carousel.Root opts={carouselOpts}>
		<Carousel.Content class="-ml-3">
			{@render children()}
		</Carousel.Content>
		<Carousel.Previous
			variant="secondary"
			size="icon-lg"
			class="-left-4 shadow-md disabled:hidden"
		/>
		<Carousel.Next variant="secondary" size="icon-lg" class="-right-4 shadow-md disabled:hidden" />
	</Carousel.Root>
</section>
