<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import UserMenu from '$lib/components/user-menu.svelte';
	import { page } from '$app/state';
	import { BookOpen, CopyCheck, House, LayoutDashboard, Library, Wrench } from '@lucide/svelte';
	import type { User } from 'better-auth';
	import type { ComponentProps } from 'svelte';

	let {
		ref = $bindable(null),
		user,
		...restProps
	}: ComponentProps<typeof Sidebar.Root> & { user: User } = $props();

	const generalLinks = [{ path: '/', label: 'Home', icon: House }];

	const managementLinks = [
		{ path: '/admin', label: 'Overview', icon: LayoutDashboard },
		{ path: '/admin/libraries', label: 'Libraries', icon: Library },
		{ path: '/admin/duplicates', label: 'Duplicates', icon: CopyCheck },
		{ path: '/admin/maintenance', label: 'Maintenance', icon: Wrench }
	];
</script>

{#snippet navMenu(links: typeof generalLinks)}
	<Sidebar.Menu>
		{#each links as link (link.path)}
			{@const isActive = page.url.pathname === link.path}
			<Sidebar.MenuItem>
				<Sidebar.MenuButton {isActive}>
					{#snippet child({ props })}
						<a href={link.path} {...props}>
							<link.icon />
							<span>{link.label}</span>
						</a>
					{/snippet}
				</Sidebar.MenuButton>
			</Sidebar.MenuItem>
		{/each}
	</Sidebar.Menu>
{/snippet}

<Sidebar.Root bind:ref {...restProps}>
	<Sidebar.Header>
		<Sidebar.Menu>
			<Sidebar.MenuItem>
				<Sidebar.MenuButton size="lg">
					{#snippet child({ props })}
						<a href="/" {...props}>
							<div
								class="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg"
							>
								<BookOpen class="size-4" />
							</div>
							<span class="text-base font-bold tracking-tight">Shoka</span>
						</a>
					{/snippet}
				</Sidebar.MenuButton>
			</Sidebar.MenuItem>
		</Sidebar.Menu>
	</Sidebar.Header>

	<Sidebar.Content>
		<Sidebar.Group>
			{@render navMenu(generalLinks)}
		</Sidebar.Group>

		<Sidebar.Group>
			<Sidebar.GroupLabel>Administration</Sidebar.GroupLabel>
			{@render navMenu(managementLinks)}
		</Sidebar.Group>
	</Sidebar.Content>
	<Sidebar.Footer>
		<UserMenu {user} />
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
