<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { can, ADMIN_ROLE, DEFAULT_ROLE } from '$lib/auth/permissions';
	import { Ellipsis, LogOut, ShieldCheck, ShieldOff, Trash2, UserPlus } from '@lucide/svelte';
	import type { ManagedUser } from './+page.server';

	let { data, form } = $props();

	const me = $derived(page.data.user);
	const canCreate = $derived(can(me, { user: ['create'] }));
	const canSetRole = $derived(can(me, { user: ['set-role'] }));
	const canBan = $derived(can(me, { user: ['ban'] }));
	const canRevoke = $derived(can(me, { session: ['revoke'] }));
	const canDelete = $derived(can(me, { user: ['delete'] }));

	let createOpen = $state(false);
	let createRole = $state(DEFAULT_ROLE);
	let deleteTarget = $state<ManagedUser | null>(null);

	const lastPage = $derived(Math.max(1, Math.ceil(data.total / data.pageSize)));

	function roleOf(user: ManagedUser) {
		return user.role ?? DEFAULT_ROLE;
	}

	function pageHref(target: number) {
		const params = new URLSearchParams();
		if (data.search) params.set('q', data.search);
		if (target > 1) params.set('page', String(target));
		const query = params.toString();
		return query ? `/admin/users?${query}` : '/admin/users';
	}
</script>

<svelte:head>
	<title>Users | Shoka Admin</title>
</svelte:head>

<div class="p-4 sm:p-6">
	<div class="flex flex-wrap items-center justify-between gap-2">
		<div>
			<h1 class="text-2xl font-bold tracking-tight">Users</h1>
			<p class="mt-1 text-sm text-muted-foreground">
				{data.total}
				{data.total === 1 ? 'account' : 'accounts'}. Admins can manage the whole instance.
			</p>
		</div>

		{#if canCreate}
			<Dialog.Root bind:open={createOpen}>
				<Dialog.Trigger>
					{#snippet child({ props })}
						<Button {...props} size="sm">
							<UserPlus class="size-3.5" />
							Add user
						</Button>
					{/snippet}
				</Dialog.Trigger>
				<Dialog.Content class="sm:max-w-md">
					<Dialog.Header>
						<Dialog.Title>Add user</Dialog.Title>
						<Dialog.Description>Creates an account with a password you choose.</Dialog.Description>
					</Dialog.Header>
					<form
						method="POST"
						action="?/create"
						class="flex flex-col gap-4"
						use:enhance={() => {
							return async ({ result, update }) => {
								if (result.type === 'success') createOpen = false;
								await update();
							};
						}}
					>
						<div class="flex flex-col gap-1.5">
							<Label for="new-username">Username</Label>
							<Input id="new-username" name="username" required maxlength={50} />
						</div>
						<div class="flex flex-col gap-1.5">
							<Label for="new-email">Email</Label>
							<Input id="new-email" name="email" type="email" required />
						</div>
						<div class="flex flex-col gap-1.5">
							<Label for="new-password">Password</Label>
							<Input id="new-password" name="password" type="password" required minlength={8} />
						</div>
						<div class="flex flex-col gap-1.5">
							<Label for="new-role">Role</Label>
							<input type="hidden" name="role" value={createRole} />
							<Select.Root type="single" bind:value={createRole}>
								<Select.Trigger id="new-role" class="w-full">{createRole}</Select.Trigger>
								<Select.Content>
									<Select.Item value={DEFAULT_ROLE} label="user">user</Select.Item>
									<Select.Item value={ADMIN_ROLE} label="admin">admin</Select.Item>
								</Select.Content>
							</Select.Root>
						</div>
						<Dialog.Footer>
							<Button type="submit">Create</Button>
						</Dialog.Footer>
					</form>
				</Dialog.Content>
			</Dialog.Root>
		{/if}
	</div>

	{#if form?.error}
		<p class="mt-4 text-sm text-destructive">{form.error}</p>
	{/if}

	<form method="GET" class="mt-6 flex max-w-sm gap-2">
		<Input name="q" placeholder="Search by name" value={data.search} />
		<Button type="submit" variant="outline">Search</Button>
	</form>

	<div class="mt-4 flex flex-col gap-2">
		{#each data.users as u (u.id)}
			{@const role = roleOf(u)}
			<div class="flex items-center gap-3 rounded-lg border border-border bg-background p-3">
				<div class="min-w-0 flex-1">
					<p class="flex items-center gap-2 truncate text-sm font-medium">
						{u.name}
						{#if u.id === me?.id}
							<span class="text-xs text-muted-foreground">(you)</span>
						{/if}
					</p>
					<p class="truncate text-xs text-muted-foreground">{u.email}</p>
				</div>

				<Badge variant={role === ADMIN_ROLE ? 'default' : 'secondary'}>{role}</Badge>
				{#if u.banned}
					<Badge variant="destructive">banned</Badge>
				{/if}

				<DropdownMenu.Root>
					<DropdownMenu.Trigger>
						{#snippet child({ props })}
							<Button {...props} variant="outline" size="icon" aria-label="User actions">
								<Ellipsis class="size-3.5" />
							</Button>
						{/snippet}
					</DropdownMenu.Trigger>
					<DropdownMenu.Content align="end" class="min-w-48">
						{#if canSetRole && u.id !== me?.id}
							<form method="POST" action="?/setRole" use:enhance>
								<input type="hidden" name="id" value={u.id} />
								<input
									type="hidden"
									name="role"
									value={role === ADMIN_ROLE ? DEFAULT_ROLE : ADMIN_ROLE}
								/>
								<DropdownMenu.Item>
									{#snippet child({ props })}
										<button type="submit" {...props} class="w-full">
											{#if role === ADMIN_ROLE}
												<ShieldOff />
												Demote to user
											{:else}
												<ShieldCheck />
												Promote to admin
											{/if}
										</button>
									{/snippet}
								</DropdownMenu.Item>
							</form>
						{/if}

						{#if canBan && u.id !== me?.id}
							<form method="POST" action="?/setBanned" use:enhance>
								<input type="hidden" name="id" value={u.id} />
								<input type="hidden" name="banned" value={u.banned ? 'false' : 'true'} />
								<DropdownMenu.Item>
									{#snippet child({ props })}
										<button type="submit" {...props} class="w-full">
											<ShieldOff />
											{u.banned ? 'Lift ban' : 'Ban'}
										</button>
									{/snippet}
								</DropdownMenu.Item>
							</form>
						{/if}

						{#if canRevoke}
							<form method="POST" action="?/revokeSessions" use:enhance>
								<input type="hidden" name="id" value={u.id} />
								<DropdownMenu.Item>
									{#snippet child({ props })}
										<button type="submit" {...props} class="w-full">
											<LogOut />
											Sign out everywhere
										</button>
									{/snippet}
								</DropdownMenu.Item>
							</form>
						{/if}

						{#if canDelete && u.id !== me?.id}
							<DropdownMenu.Separator />
							<DropdownMenu.Item variant="destructive" onSelect={() => (deleteTarget = u)}>
								<Trash2 />
								Delete
							</DropdownMenu.Item>
						{/if}
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</div>
		{/each}
	</div>

	{#if lastPage > 1}
		<div class="mt-4 flex items-center justify-between text-sm">
			<Button variant="outline" size="sm" href={pageHref(data.page - 1)} disabled={data.page <= 1}>
				Previous
			</Button>
			<span class="text-muted-foreground">Page {data.page} of {lastPage}</span>
			<Button
				variant="outline"
				size="sm"
				href={pageHref(data.page + 1)}
				disabled={data.page >= lastPage}
			>
				Next
			</Button>
		</div>
	{/if}
</div>

<AlertDialog.Root open={deleteTarget !== null} onOpenChange={(v) => !v && (deleteTarget = null)}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Delete {deleteTarget?.name}?</AlertDialog.Title>
			<AlertDialog.Description>
				This removes the account, its sessions, and its per-user data. It cannot be undone.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<form
			method="POST"
			action="?/delete"
			use:enhance={() => {
				return async ({ update }) => {
					deleteTarget = null;
					await update();
				};
			}}
		>
			<input type="hidden" name="id" value={deleteTarget?.id ?? ''} />
			<AlertDialog.Footer>
				<AlertDialog.Cancel type="button">Cancel</AlertDialog.Cancel>
				<Button type="submit" variant="destructive">Delete</Button>
			</AlertDialog.Footer>
		</form>
	</AlertDialog.Content>
</AlertDialog.Root>
