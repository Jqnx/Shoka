<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import UserMenu from '$lib/components/user-menu.svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import {
		BookOpen,
		Drama,
		House,
		Library,
		Paintbrush,
		Settings,
		Tags,
		Users
	} from '@lucide/svelte';
	import type { User } from 'better-auth';
	import type { Library as LibraryType } from '$lib/types';
	import type { ComponentProps } from 'svelte';

	let {
		ref = $bindable(null),
		user,
		libraries,
		...restProps
	}: ComponentProps<typeof Sidebar.Root> & {
		user: User;
		libraries: LibraryType[];
	} = $props();

	const generalLinks = [{ path: '/', label: 'Home', icon: House }];

	const adminLinks = [{ path: '/admin', label: 'Admin', icon: Settings }];

	const metadataLinks = [
		{ path: '/artist', label: 'Artists', icon: Paintbrush },
		{ path: '/tag', label: 'Tags', icon: Tags },
		{ path: '/character', label: 'Characters', icon: Users },
		{ path: '/parody', label: 'Parodies', icon: Drama }
	];
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
			<Sidebar.GroupLabel>Libraries</Sidebar.GroupLabel>
			{#if libraries.length === 0}
				<p class="px-2 text-xs text-sidebar-foreground/60">
					No libraries yet. <a href={resolve('/admin/libraries')} class="underline">Add one</a>.
				</p>
			{:else}
				<Sidebar.Menu>
					{#each libraries as library (library.id)}
						{@const isActive = page.url.pathname === `/${library.id}`}
						<Sidebar.MenuItem>
							<Sidebar.MenuButton {isActive}>
								{#snippet child({ props })}
									<a href={resolve(`/${library.id}`)} data-sveltekit-preload-data="tap" {...props}>
										<Library />
										<span>{library.name}</span>
									</a>
								{/snippet}
							</Sidebar.MenuButton>
						</Sidebar.MenuItem>
					{/each}
				</Sidebar.Menu>
			{/if}
		</Sidebar.Group>

		<Sidebar.Group>
			<Sidebar.GroupLabel>Metadata</Sidebar.GroupLabel>
			{@render navMenu(metadataLinks)}
		</Sidebar.Group>
	</Sidebar.Content>
	<Sidebar.Footer>
		{@render navMenu(adminLinks)}
		<UserMenu {user} />
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
