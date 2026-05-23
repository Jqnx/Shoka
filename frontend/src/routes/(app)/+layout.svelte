<script lang="ts">
	import { BookOpen, House, Library, Settings } from '@lucide/svelte';
	import * as Avatar from '$lib/components/ui/avatar';
	import { page } from '$app/state';
	import ModeToggle from '$lib/components/ModeToggle.svelte';

	let { children } = $props();

	const navLinks = [
		{ path: '/', label: 'Home', icon: House },
		{ path: '/a', label: 'Archives', icon: Library },
		{ path: '/admin', label: 'Admin', icon: Settings }
	];
</script>

<div class="flex min-h-screen flex-col bg-muted/40">
	<!-- Sticky top navbar -->
	<header
		class="sticky top-0 z-50 h-14 border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80"
	>
		<div class="mx-auto flex h-full max-w-screen-2xl items-center gap-6 px-4 sm:px-6">
			<!-- Logo -->
			<a href="/" class="flex shrink-0 items-center gap-2">
				<BookOpen class="size-5 text-foreground" />
				<span class="text-base font-bold tracking-tight">Shoka</span>
			</a>

			<div class="h-5 w-px bg-border"></div>

			<!-- Nav links -->
			<nav class="flex items-center gap-1">
				{#each navLinks as link (link.path)}
					{@const isActive = page.url.pathname === link.path}
					<a
						href={link.path}
						class="inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-sm font-medium transition-colors
							{isActive
							? 'bg-muted text-foreground'
							: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
					>
						{link.label}
					</a>
				{/each}
			</nav>

			<div class="flex-1"></div>

			<Avatar.Root class="size-8">
				<Avatar.Image src="" alt="User avatar" />
				<Avatar.Fallback>U</Avatar.Fallback>
			</Avatar.Root>
			<ModeToggle />
		</div>
	</header>

	<!-- Page content -->
	<main class="flex-1">
		{@render children()}
	</main>
</div>
