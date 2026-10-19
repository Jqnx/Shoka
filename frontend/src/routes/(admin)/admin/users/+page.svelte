<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Form from '$lib/components/ui/form/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { can, ADMIN_ROLE, DEFAULT_ROLE } from '$lib/auth/permissions';
	import { superForm, type Infer } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import { createUserSchema, type CreateUserSchema } from '$lib/schemas/create-user';
	import { cn } from '$lib/utils.js';
	import { Ellipsis, Ban, LogOut, ShieldCheck, ShieldOff, Trash2, UserPlus } from '@lucide/svelte';
	import type { ManagedUser } from './+page.server';

	let { data, form } = $props();

	const me = $derived(page.data.user);
	const canCreate = $derived(can(me, { user: ['create'] }));
	const canSetRole = $derived(can(me, { user: ['set-role'] }));
	const canBan = $derived(can(me, { user: ['ban'] }));
	const canRevoke = $derived(can(me, { session: ['revoke'] }));
	const canDelete = $derived(can(me, { user: ['delete'] }));

	let createOpen = $state(false);
	let deleteTarget = $state<ManagedUser | null>(null);
	let saving = $state(false);
	let signupForm = $state<HTMLFormElement>(null!);
	// Optimistic value for "sign-ups disabled" while the toggle action is in
	// flight; null means the server's value. Reset after update() so a failed
	// action reverts.
	let signupTarget = $state<boolean | null>(null);
	const signupChecked = $derived(signupTarget ?? !data.signupEnabled);

	function toggleSignup(checked: boolean) {
		signupTarget = checked;
		signupForm?.requestSubmit();
	}

	// svelte-ignore state_referenced_locally
	const createUser = superForm(data.createForm, {
		id: 'createUser',
		validators: zod4Client(createUserSchema),
		// A failed create leaves the dialog open with the entered values and the
		// field errors; a successful one closes it and clears the form, so the
		// next account starts from the defaults.
		onResult: ({ result }) => {
			if (result.type === 'success') createOpen = false;
		}
	});
	const {
		form: createUserData,
		enhance: createUserEnhance,
		submitting: createUserSubmitting
	} = createUser;

	// bits-ui hands the select's value back as a plain string; the schema's
	// enum is what narrows it.
	function setCreateUserRole(value: string) {
		$createUserData.role = value as Infer<CreateUserSchema>['role'];
	}

	const lastPage = $derived(Math.max(1, Math.ceil(data.total / data.pageSize)));

	function roleOf(user: ManagedUser) {
		return user.role ?? DEFAULT_ROLE;
	}

	// An action failure that names a row (`{ action, id, error }`) is rendered on
	// that row, beside the menu that produced it. One that names no row keeps a
	// slot of its own: the create dialog's failures belong in the dialog (which
	// stays open on a failure, and the page-level slot is behind it) and a
	// permission guard's belong on the page. So does one whose row is no longer
	// on the page (deleted by someone else since this list was loaded): a message
	// nobody can see is worse than a message in the wrong place.
	const failedRowID = $derived(form && 'id' in form ? form.id : undefined);
	const dialogError = $derived(
		form && 'error' in form && 'action' in form && form.action === 'create' ? form.error : undefined
	);
	const pageError = $derived(
		form &&
			'error' in form &&
			dialogError === undefined &&
			(failedRowID === undefined || !data.users.some((u) => u.id === failedRowID))
			? form.error
			: undefined
	);

	function rowError(userID: string): string | undefined {
		if (!form || !('error' in form) || !('id' in form)) return undefined;

		return form.id === userID ? form.error : undefined;
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
					<form method="POST" action="?/create" class="flex flex-col gap-4" use:createUserEnhance>
						<Form.Field form={createUser} name="username">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Username</Form.Label>
									<Input {...props} bind:value={$createUserData.username} />
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>
						<Form.Field form={createUser} name="email">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Email</Form.Label>
									<Input {...props} bind:value={$createUserData.email} />
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>
						<Form.Field form={createUser} name="password">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Password</Form.Label>
									<Input {...props} type="password" bind:value={$createUserData.password} />
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>
						<Form.Field form={createUser} name="role">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Role</Form.Label>
									<Select.Root
										type="single"
										name={props.name}
										value={$createUserData.role}
										onValueChange={setCreateUserRole}
									>
										<Select.Trigger {...props} class="w-full">
											{$createUserData.role}
										</Select.Trigger>
										<Select.Content>
											<Select.Item value={DEFAULT_ROLE} label="user">user</Select.Item>
											<Select.Item value={ADMIN_ROLE} label="admin">admin</Select.Item>
										</Select.Content>
									</Select.Root>
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>
						{#if dialogError}
							<p class="text-sm text-destructive">{dialogError}</p>
						{/if}
						<Dialog.Footer>
							<Form.Button disabled={$createUserSubmitting}>
								{$createUserSubmitting ? 'Creating…' : 'Create'}
							</Form.Button>
						</Dialog.Footer>
					</form>
				</Dialog.Content>
			</Dialog.Root>
		{/if}
	</div>

	{#if pageError}
		<p class="mt-4 text-sm text-destructive">{pageError}</p>
	{/if}

	<div class="mt-6 flex flex-wrap items-center justify-between gap-3">
		<form method="GET" class="flex max-w-sm gap-2">
			<Input name="q" placeholder="Search by name" value={data.search} />
			<Button type="submit" variant="outline">Search</Button>
		</form>

		{#if data.canToggleSignup}
			<form
				bind:this={signupForm}
				method="POST"
				action="?/setSignupEnabled"
				class="ml-auto flex items-center gap-2"
				use:enhance={() => {
					saving = true;
					return async ({ update }) => {
						await update();
						saving = false;
						signupTarget = null;
					};
				}}
			>
				<input type="hidden" name="enabled" value={String(!data.signupEnabled)} />
				<Switch
					id="disable-signup"
					checked={signupChecked}
					disabled={saving}
					onCheckedChange={toggleSignup}
				/>
				<Label for="disable-signup" class="text-sm font-normal">Disable sign-up</Label>
			</form>
		{/if}
	</div>

	<div class="mt-4 flex flex-col gap-2">
		{#each data.users as u (u.id)}
			{@const role = roleOf(u)}
			{@const error = rowError(u.id)}
			<div class="rounded-xl border border-border p-4">
				<div class="flex items-center gap-3">
					<div class="min-w-0 flex-1">
						<p class="flex items-center gap-2 truncate text-sm font-medium">
							{u.name}
							{#if u.id === me?.id}
								<span class="text-xs text-muted-foreground">(you)</span>
							{/if}
						</p>
						<p class="truncate text-xs text-muted-foreground">{u.email}</p>
					</div>

					<Badge class="h-8" variant={role === ADMIN_ROLE ? 'default' : 'secondary'}>{role}</Badge>
					{#if u.banned}
						<Badge class="h-8" variant="destructive">banned</Badge>
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
											<button type="submit" {...props} class={cn(props.class as string, 'w-full')}>
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
											<button type="submit" {...props} class={cn(props.class as string, 'w-full')}>
												{#if u.banned}
													<ShieldCheck />
													Lift ban
												{:else}
													<Ban />
													Ban
												{/if}
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
											<button type="submit" {...props} class={cn(props.class as string, 'w-full')}>
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
				{#if error}
					<p class="mt-2 text-xs text-destructive">{error}</p>
				{/if}
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
