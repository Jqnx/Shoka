<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { CircleAlert, House, Lock, SearchX, TriangleAlert } from '@lucide/svelte';

	// Rendered by every route that doesn't have a closer +error.svelte, so it
	// stands on its own - no sidebar, no page chrome. Only the statuses the app
	// actually throws get their own wording.
	const INFO: Record<number, { title: string; icon: typeof House }> = {
		403: { title: 'Access denied', icon: Lock },
		404: { title: 'Page not found', icon: SearchX },
		500: { title: 'Something went wrong', icon: TriangleAlert }
	};

	const info = $derived(INFO[page.status] ?? { title: `Error ${page.status}`, icon: CircleAlert });
</script>

<svelte:head>
	<title>{info.title} | Shoka</title>
</svelte:head>

<div class="flex h-screen w-full items-center justify-center px-4">
	<Card.Root class="mx-auto w-full max-w-sm">
		<Card.Header>
			<Card.Title class="flex items-center gap-2 text-2xl">
				<info.icon class="size-5" />
				{info.title}
			</Card.Title>
			<Card.Description>{page.error?.message}</Card.Description>
		</Card.Header>
		<Card.Footer class="justify-between gap-3">
			<span class="text-xs text-muted-foreground tabular-nums">Error {page.status}</span>
			<Button href="/" variant="outline" size="sm">Go home</Button>
		</Card.Footer>
	</Card.Root>
</div>
