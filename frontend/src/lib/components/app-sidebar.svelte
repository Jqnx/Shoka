<script lang="ts">
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { cn } from '$lib/utils.js';
	import { toggleMode, mode } from 'mode-watcher';
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import {
		BookOpen,
		ChevronsUpDown,
		Drama,
		House,
		Library,
		LogOut,
		Moon,
		Settings,
		Sun,
		Tags,
		Users
	} from '@lucide/svelte';
	import type { User } from 'better-auth';
	import type { ComponentProps } from 'svelte';

	let {
		ref = $bindable(null),
		user,
		...restProps
	}: ComponentProps<typeof Sidebar.Root> & { user: User } = $props();

	const sidebar = Sidebar.useSidebar();

	const generalLinks = [{ path: '/', label: 'Home', icon: House }];

	const adminLinks = [{ path: '/admin', label: 'Admin', icon: Settings }];

	const libraryLinks = [{ path: '/a', label: 'Archives', icon: Library }];

	const metadataLinks = [
		{ path: '/tag', label: 'Tags', icon: Tags },
		{ path: '/character', label: 'Characters', icon: Users },
		{ path: '/parody', label: 'Parodies', icon: Drama }
	];

	const initials = $derived(
		user.name
			.split(' ')
			.map((part: string) => part[0])
			.slice(0, 2)
			.join('')
			.toUpperCase()
	);
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
			<Sidebar.GroupLabel>Libraries</Sidebar.GroupLabel>
			{@render navMenu(libraryLinks)}
		</Sidebar.Group>

		<Sidebar.Group>
			<Sidebar.GroupLabel>Metadata</Sidebar.GroupLabel>
			{@render navMenu(metadataLinks)}
		</Sidebar.Group>
	</Sidebar.Content>
	<Sidebar.Footer>
		{@render navMenu(adminLinks)}
		<Sidebar.Menu>
			<Sidebar.MenuItem>
				<DropdownMenu.Root>
					<DropdownMenu.Trigger>
						{#snippet child({ props })}
							<Sidebar.MenuButton
								size="lg"
								class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
								{...props}
							>
								<Avatar.Root class="size-8 rounded-lg">
									<Avatar.Image src={user.image ?? ''} alt={user.name} />
									<Avatar.Fallback class="rounded-lg">{initials}</Avatar.Fallback>
								</Avatar.Root>
								<span class="text-sm font-medium">{user.name}</span>
								<ChevronsUpDown class="ms-auto size-4" />
							</Sidebar.MenuButton>
						{/snippet}
					</DropdownMenu.Trigger>
					<DropdownMenu.Content
						class="w-(--bits-dropdown-menu-anchor-width) min-w-56 rounded-lg"
						side={sidebar.isMobile ? 'bottom' : 'right'}
						align="end"
						sideOffset={4}
					>
						<DropdownMenu.Label class="p-0 font-normal">
							<div class="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
								<Avatar.Root class="size-8 rounded-lg">
									<Avatar.Image src={user.image ?? ''} alt={user.name} />
									<Avatar.Fallback class="rounded-lg">{initials}</Avatar.Fallback>
								</Avatar.Root>
								<div class="grid flex-1 leading-tight">
									<span class="truncate text-sm font-medium">{user.name}</span>
									<span class="truncate text-xs text-muted-foreground">{user.email}</span>
								</div>
							</div>
						</DropdownMenu.Label>
						<DropdownMenu.Separator />
						<DropdownMenu.Item onclick={() => toggleMode()}>
							{#if mode.current === 'dark'}
								<Sun />
								Light mode
							{:else}
								<Moon />
								Dark mode
							{/if}
						</DropdownMenu.Item>
						<DropdownMenu.Separator />
						<form method="POST" action="/logout" use:enhance>
							<DropdownMenu.Item>
								{#snippet child({ props })}
									<button type="submit" {...props} class={cn(props.class as string, 'w-full')}>
										<LogOut />
										Log out
									</button>
								{/snippet}
							</DropdownMenu.Item>
						</form>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Sidebar.MenuItem>
		</Sidebar.Menu>
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
