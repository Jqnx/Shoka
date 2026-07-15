<script lang="ts">
	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Search } from '@lucide/svelte';
	import { page } from '$app/state';

	let { data, children } = $props();

	// The archives list route needs a library to scope to; default to the
	// first one until there's a "current library" concept to derive this from.
	const searchAction = $derived(data.libraries[0] ? `/${data.libraries[0].id}` : '/');
</script>

<Sidebar.Provider>
	<AppSidebar user={data.user} libraries={data.libraries} />
	<Sidebar.Inset class="bg-muted/40">
		<header
			class="sticky top-0 z-50 flex h-14 shrink-0 items-center gap-2 border-b border-border bg-background/95 px-4 backdrop-blur supports-backdrop-filter:bg-background/80 sm:px-6"
		>
			<Sidebar.Trigger class="-ms-1" />

			<div class="flex flex-1 justify-center">
				<form method="GET" action={searchAction} class="w-full max-w-lg">
					<Label for="search" class="sr-only">Search</Label>
					<div class="relative">
						<Search
							class="pointer-events-none absolute inset-s-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
						/>
						<Input
							id="search"
							name="q"
							type="search"
							placeholder="Search"
							value={page.url.searchParams.get('q') ?? ''}
							class="h-10 ps-9"
						/>
					</div>
				</form>
			</div>

			<div class="w-6" aria-hidden="true"></div>
		</header>

		<main class="flex-1">
			{@render children()}
		</main>
	</Sidebar.Inset>
</Sidebar.Provider>
