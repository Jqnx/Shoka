<script lang="ts">
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import * as Form from '$lib/components/ui/form/index.js';
	import { ADMIN_ROLE, DEFAULT_ROLE, can } from '$lib/auth/permissions';
	import { adminPasswordSchema, adminUsernameSchema } from '$lib/schemas/admin-user';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import { ArrowLeft, ShieldCheck, User } from '@lucide/svelte';

	let { data, form } = $props();

	const me = $derived(page.data.user);
	const isAdminTarget = $derived(data.target.role === ADMIN_ROLE);
	const isSelf = $derived(data.target.id === me?.id);
	const canEdit = $derived(can(me, { user: ['update'] }));
	const canSetPassword = $derived(can(me, { user: ['set-password'] }));
	const canSetRole = $derived(can(me, { user: ['set-role'] }));
	const canBan = $derived(can(me, { user: ['ban'] }));
	const canManageStatus = $derived(canSetRole || canBan);

	// --- Username ---
	let usernameSaved = $state(false);
	// svelte-ignore state_referenced_locally
	const username = superForm(data.usernameForm, {
		id: 'username',
		validators: zod4Client(adminUsernameSchema),
		// The new value is the server's, not the seeded one: resetting would put
		// the old username back after a successful save.
		resetForm: false,
		onSubmit: () => {
			usernameSaved = false;
		},
		onUpdated: ({ form }) => {
			usernameSaved = form.valid;
		}
	});
	const { form: usernameData, enhance: usernameEnhance, submitting: usernameSubmitting } = username;

	// --- Password ---
	let passwordSaved = $state(false);
	// svelte-ignore state_referenced_locally
	const password = superForm(data.passwordForm, {
		id: 'password',
		validators: zod4Client(adminPasswordSchema),
		onSubmit: () => {
			passwordSaved = false;
		},
		onUpdated: ({ form }) => {
			passwordSaved = form.valid;
		}
	});
	const { form: passwordData, enhance: passwordEnhance, submitting: passwordSubmitting } = password;

	// The account switches and Save access invalidate the page, and the load then
	// re-seeds both superforms - discarding whatever the admin had typed in the
	// other cards. Snapshot the two edit forms across such an update; only the
	// username/password forms themselves replace their own values (on submit).
	async function updateKeepingEdits(
		update: (options?: { reset?: boolean; invalidateAll?: boolean }) => Promise<void>,
		options?: { reset?: boolean }
	) {
		const usernameEdit = { ...$usernameData };
		const passwordEdit = { ...$passwordData };

		await update(options);

		$usernameData = usernameEdit;
		$passwordData = passwordEdit;
	}

	// --- Role and ban toggles ---
	// Each switch drives its own form: the optimistic value is what the hidden
	// input submits, and resetting it to null after update() hands the switch
	// back to the server's value (so a rejected change visibly reverts).
	let roleForm = $state<HTMLFormElement>(null!);
	let roleTarget = $state<boolean | null>(null);
	let roleSaving = $state(false);
	const roleChecked = $derived(roleTarget ?? isAdminTarget);

	let banForm = $state<HTMLFormElement>(null!);
	let banTarget = $state<boolean | null>(null);
	let banSaving = $state(false);
	const banChecked = $derived(banTarget ?? Boolean(data.target.banned));

	function toggleRole(checked: boolean) {
		roleTarget = checked;
		roleForm?.requestSubmit();
	}

	function toggleBan(checked: boolean) {
		banTarget = checked;
		banForm?.requestSubmit();
	}

	function actionError(action: string) {
		if (!form || !('action' in form) || !('error' in form)) return undefined;

		return form.action === action ? form.error : undefined;
	}

	function initialsOf(name: string) {
		return name
			.split(' ')
			.map((part) => part[0])
			.slice(0, 2)
			.join('')
			.toUpperCase();
	}
</script>

<svelte:head>
	<title>{data.target.name} | Shoka Admin</title>
</svelte:head>

<div class="mx-auto max-w-3xl p-4 sm:p-6">
	<a
		href={resolve('/admin/users')}
		class="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
	>
		<ArrowLeft class="size-3.5" />
		Users
	</a>

	<div class="mt-3 flex items-center gap-3">
		<Avatar.Root class="size-12">
			<Avatar.Image src={data.target.image ?? ''} alt={data.target.name} />
			<Avatar.Fallback class="bg-muted text-sm font-semibold select-none">
				{initialsOf(data.target.name)}
			</Avatar.Fallback>
		</Avatar.Root>
		<div class="min-w-0">
			<div class="flex flex-wrap items-center gap-2">
				<h1 class="text-2xl font-bold tracking-tight">{data.target.name}</h1>
				{#if isAdminTarget}
					<span class="inline-flex shrink-0 items-center text-primary" title="Administrator">
						<ShieldCheck class="size-4" aria-hidden="true" />
						<span class="sr-only">Administrator</span>
					</span>
				{:else}
					<span class="inline-flex shrink-0 items-center text-muted-foreground" title="User">
						<User class="size-4" aria-hidden="true" />
						<span class="sr-only">User</span>
					</span>
				{/if}
				{#if data.target.banned}
					<Badge variant="destructive">banned</Badge>
				{/if}
			</div>
		</div>
	</div>

	{#if canEdit}
		<Card.Root class="mt-6">
			<form class="contents" method="POST" action="?/updateUsername" use:usernameEnhance>
				<Card.Header>
					<Card.Title>Username</Card.Title>
				</Card.Header>
				<Card.Content>
					<Form.Field form={username} name="username">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Username</Form.Label>
								<Input {...props} bind:value={$usernameData.username} />
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>
				</Card.Content>
				<Card.Footer class="gap-3">
					<Button type="submit" size="sm" disabled={$usernameSubmitting}>
						{$usernameSubmitting ? 'Saving…' : 'Save'}
					</Button>
					{#if actionError('updateUsername')}
						<p class="text-sm text-destructive">{actionError('updateUsername')}</p>
					{:else if usernameSaved}
						<p class="text-sm text-muted-foreground">Saved.</p>
					{/if}
				</Card.Footer>
			</form>
		</Card.Root>
	{/if}

	{#if canSetPassword}
		<Card.Root class="mt-6">
			<form class="contents" method="POST" action="?/setPassword" use:passwordEnhance>
				<Card.Header>
					<Card.Title>Password</Card.Title>
					<Card.Description>
						Set a new password for this user. Their signed-in sessions stay active.
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<Form.Field form={password} name="newPassword">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>New password</Form.Label>
								<Input {...props} type="password" bind:value={$passwordData.newPassword} />
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>
					<Form.Field form={password} name="confirmPassword">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Confirm new password</Form.Label>
								<Input {...props} type="password" bind:value={$passwordData.confirmPassword} />
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>
				</Card.Content>
				<Card.Footer class="gap-3">
					<Button type="submit" size="sm" disabled={$passwordSubmitting}>
						{$passwordSubmitting ? 'Saving…' : 'Set password'}
					</Button>
					{#if actionError('setPassword')}
						<p class="text-sm text-destructive">{actionError('setPassword')}</p>
					{:else if passwordSaved}
						<p class="text-sm text-muted-foreground">Password updated.</p>
					{/if}
				</Card.Footer>
			</form>
		</Card.Root>
	{/if}

	{#if canManageStatus}
		<Card.Root class="mt-6">
			<Card.Header>
				<Card.Title>Account status</Card.Title>
				<Card.Description>
					Admins manage the whole instance. A disabled account cannot sign in.
				</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-5">
				{#if canSetRole}
					<form
						bind:this={roleForm}
						method="POST"
						action="?/setRole"
						class="flex items-start justify-between gap-4"
						use:enhance={() => {
							roleSaving = true;
							return async ({ update }) => {
								await updateKeepingEdits(update);
								roleSaving = false;
								roleTarget = null;
							};
						}}
					>
						<!--
						The hidden value is the desired *new* state, derived from the
						server's data: requestSubmit() submits before Svelte flushes the
						DOM, so reading it from the optimistic state would send the old
						value and turn the toggle into a no-op.
					-->
						<input type="hidden" name="role" value={isAdminTarget ? DEFAULT_ROLE : ADMIN_ROLE} />
						<div class="min-w-0 space-y-1">
							<Label for="user-admin" class="text-sm font-medium">Administrator</Label>
							<p class="text-xs text-muted-foreground">
								{isSelf
									? 'You cannot change your own role.'
									: isAdminTarget
										? 'At least one admin must remain.'
										: 'Can manage libraries, metadata and users.'}
							</p>
						</div>
						<Switch
							id="user-admin"
							checked={roleChecked}
							disabled={isSelf || roleSaving}
							onCheckedChange={toggleRole}
						/>
					</form>
				{/if}

				{#if canBan}
					<form
						bind:this={banForm}
						method="POST"
						action="?/setBanned"
						class="flex items-start justify-between gap-4"
						use:enhance={() => {
							banSaving = true;
							return async ({ update }) => {
								await updateKeepingEdits(update);
								banSaving = false;
								banTarget = null;
							};
						}}
					>
						<input type="hidden" name="banned" value={String(!data.target.banned)} />
						<div class="min-w-0 space-y-1">
							<Label for="user-disabled" class="text-sm font-medium">Disabled</Label>
							<p class="text-xs text-muted-foreground">
								{isSelf
									? 'You cannot disable your own account.'
									: 'Blocks sign-in and signs the user out everywhere.'}
							</p>
						</div>
						<Switch
							id="user-disabled"
							checked={banChecked}
							disabled={isSelf || banSaving}
							onCheckedChange={toggleBan}
						/>
					</form>
				{/if}
			</Card.Content>
			{#if actionError('setRole') || actionError('setBanned')}
				<Card.Footer>
					<p class="text-sm text-destructive">
						{actionError('setRole') ?? actionError('setBanned')}
					</p>
				</Card.Footer>
			{/if}
		</Card.Root>
	{/if}

	<Card.Root class="mt-6">
		<Card.Header>
			<Card.Title>Library access</Card.Title>
			<Card.Description>
				{#if isAdminTarget}
					Admins can access every library.
				{:else}
					Choose which libraries this user may browse. A user without a grant sees nothing.
				{/if}
			</Card.Description>
		</Card.Header>

		{#if !isAdminTarget}
			<form
				class="contents"
				method="POST"
				action="?/saveAccess"
				use:enhance={() =>
					async ({ update }) => {
						// SvelteKit's default callback resets the form after a
						// successful action. The checkboxes are controlled values,
						// and a native reset would clear their hidden inputs
						// without re-rendering them, so the next Save would submit
						// unchecked even though the UI still shows the grants.
						await updateKeepingEdits(update, { reset: false });
					}}
			>
				<Card.Content>
					{#if data.libraries.length === 0}
						<p class="text-sm text-muted-foreground">No libraries yet.</p>
					{:else}
						<div class="space-y-2">
							{#each data.libraries as library (library.id)}
								<label class="flex items-center gap-2 text-sm">
									<Checkbox
										name="libraryIds"
										value={library.id}
										checked={data.grantedLibraryIds.includes(library.id)}
									/>
									<span class="truncate">{library.name}</span>
								</label>
							{/each}
						</div>
					{/if}
				</Card.Content>
				<Card.Footer class="flex items-center gap-3">
					<Button type="submit" size="sm" disabled={data.libraries.length === 0}>
						Save access
					</Button>
					{#if form?.action === 'saveAccess' && 'error' in form}
						<p class="text-sm text-destructive">{form.error}</p>
					{:else if form?.action === 'saveAccess'}
						<p class="text-sm text-muted-foreground">Saved.</p>
					{/if}
				</Card.Footer>
			</form>
		{/if}
	</Card.Root>
</div>
