<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { buttonVariants } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';

	let { children } = $props();

	const links = [
		{ href: '/settings', label: 'Account' },
		{ href: '/settings/appearance', label: 'Appearance' }
	] as const;
</script>

<div class="mx-auto max-w-2xl p-4 sm:p-6">
	<nav class="mb-6 inline-flex gap-1 rounded-lg bg-muted p-1">
		{#each links as link (link.href)}
			{@const active = page.url.pathname === link.href}
			<a
				href={resolve(link.href)}
				aria-current={active ? 'page' : undefined}
				class={cn(
					buttonVariants({ variant: 'ghost', size: 'sm' }),
					active && 'bg-background shadow-sm'
				)}
			>
				{link.label}
			</a>
		{/each}
	</nav>

	{@render children()}
</div>
