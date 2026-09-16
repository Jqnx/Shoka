<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import UserMenu from '$lib/components/user-menu.svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import {
		Activity,
		BookOpen,
		CopyCheck,
		House,
		LayoutDashboard,
		Library,
		UsersRound,
		Wrench
	} from '@lucide/svelte';
	import type { ComponentProps } from 'svelte';
	import { can, type Permissions } from '$lib/auth/permissions';

	let {
		ref = $bindable(null),
		user,
		...restProps
	}: ComponentProps<typeof Sidebar.Root> & { user: NonNullable<App.Locals['user']> } = $props();

	const generalLinks = [{ path: '/', label: 'Home', icon: House }];

	type NavLink = { path: string; label: string; icon: typeof House; needs?: Permissions };

	const allManagementLinks: NavLink[] = [
		{ path: '/admin', label: 'Overview', icon: LayoutDashboard },
		{ path: '/admin/libraries', label: 'Libraries', icon: Library, needs: { library: ['read'] } },
		{
			path: '/admin/duplicates',
			label: 'Duplicates',
			icon: CopyCheck,
			needs: { maintenance: ['read'] }
		},
		{ path: '/admin/jobs', label: 'Jobs', icon: Activity, needs: { job: ['read'] } },
		{
			path: '/admin/maintenance',
			label: 'Maintenance',
			icon: Wrench,
			needs: { maintenance: ['read'] }
		},
		{ path: '/admin/users', label: 'Users', icon: UsersRound, needs: { user: ['list'] } }
	];

	const managementLinks = $derived(
		allManagementLinks.filter((link) => !link.needs || can(user, link.needs))
	);
</script>

{#snippet navMenu(links: typeof generalLinks)}
	<Sidebar.Menu>
		{#each links as link (link.path)}
			{@const isActive = page.url.pathname === link.path}
			<Sidebar.MenuItem>
				<Sidebar.MenuButton {isActive}>
					{#snippet child({ props })}
						<a href={resolve(link.path as `/${string}`)} {...props}>
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
						<a href={resolve('/')} {...props}>
							<div
								class="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground"
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
