<script lang="ts">
	import AdminSidebar from '$lib/components/admin-sidebar.svelte';
	import AudioPlayerBar from '$lib/components/AudioPlayerBar.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { player } from '$lib/player.svelte';

	let { data, children } = $props();
</script>

<Sidebar.Provider>
	<AdminSidebar user={data.user} />
	<Sidebar.Inset
		class="bg-muted/40 transition-[padding-bottom] duration-200 ease-out {player.source &&
		player.barVisible &&
		!player.expanded
			? 'pb-[4.5rem]'
			: ''}"
	>
		<header
			class="sticky top-0 z-50 flex h-14 shrink-0 items-center gap-2 border-b border-border bg-background/95 px-4 backdrop-blur supports-backdrop-filter:bg-background/80 sm:px-6"
		>
			<Sidebar.Trigger class="-ms-1" />
		</header>

		<main class="flex-1">
			{@render children()}
		</main>
	</Sidebar.Inset>
	<AudioPlayerBar />
</Sidebar.Provider>
