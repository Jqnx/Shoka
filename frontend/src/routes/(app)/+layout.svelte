<script lang="ts">
	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import SearchDialog from '$lib/components/SearchDialog.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar';

	let { data, children } = $props();
</script>

<Sidebar.Provider>
	<AppSidebar user={data.user} libraries={data.libraries} />
	<Sidebar.Inset class="bg-muted/40">
		<header
			class="sticky top-0 z-50 flex h-14 shrink-0 items-center gap-2 border-b border-border bg-background/95 px-4 backdrop-blur supports-backdrop-filter:bg-background/80 sm:px-6"
		>
			<Sidebar.Trigger class="-ms-1" />

			<div class="flex flex-1 justify-center">
				<SearchDialog libraries={data.libraries} />
			</div>

			<div class="w-6" aria-hidden="true"></div>
		</header>

		<main class="flex-1">
			{@render children()}
		</main>
	</Sidebar.Inset>
</Sidebar.Provider>
