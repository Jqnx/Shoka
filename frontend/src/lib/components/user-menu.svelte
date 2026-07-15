<script lang="ts">
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { cn } from '$lib/utils.js';
	import { toggleMode, mode } from 'mode-watcher';
	import { enhance } from '$app/forms';
	import { ChevronsUpDown, LogOut, Moon, Sun } from '@lucide/svelte';
	import type { User } from 'better-auth';

	let { user }: { user: User } = $props();

	const sidebar = Sidebar.useSidebar();

	const initials = $derived(
		user.name
			.split(' ')
			.map((part: string) => part[0])
			.slice(0, 2)
			.join('')
			.toUpperCase()
	);
</script>

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
