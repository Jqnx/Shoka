<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Form from '$lib/components/ui/form/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import { profileSchema } from '$lib/schemas/profile';
	import { changePasswordSchema } from '$lib/schemas/change-password';

	let { data } = $props();

	let profileSaved = $state(false);
	// svelte-ignore state_referenced_locally
	const profile = superForm(data.profileForm, {
		id: 'profile',
		validators: zod4Client(profileSchema),
		resetForm: false,
		onSubmit: () => {
			profileSaved = false;
		},
		onUpdated: ({ form }) => {
			profileSaved = form.valid;
		}
	});
	const { form: profileData, enhance: profileEnhance, submitting: profileSubmitting } = profile;

	let passwordSaved = $state(false);
	// svelte-ignore state_referenced_locally
	const password = superForm(data.passwordForm, {
		id: 'password',
		validators: zod4Client(changePasswordSchema),
		onSubmit: () => {
			passwordSaved = false;
		},
		onUpdated: ({ form }) => {
			passwordSaved = form.valid;
		}
	});
	const { form: passwordData, enhance: passwordEnhance, submitting: passwordSubmitting } = password;
</script>

<svelte:head>
	<title>Settings | Shoka</title>
</svelte:head>

<div>
	<h1 class="text-2xl font-bold tracking-tight">Account</h1>
	<p class="mt-1 text-sm text-muted-foreground">Manage your account details.</p>

	<Card.Root class="mt-6">
		<form class="contents" method="POST" action="?/updateProfile" use:profileEnhance>
			<Card.Header>
				<Card.Title>Profile</Card.Title>
				<Card.Description>Your display name and username.</Card.Description>
			</Card.Header>
			<Card.Content>
				<Form.Field form={profile} name="name">
					<Form.Control>
						{#snippet children({ props })}
							<Form.Label>Name</Form.Label>
							<Input {...props} bind:value={$profileData.name} />
						{/snippet}
					</Form.Control>
					<Form.FieldErrors />
				</Form.Field>
				<Form.Field form={profile} name="username">
					<Form.Control>
						{#snippet children({ props })}
							<Form.Label>Username</Form.Label>
							<Input {...props} bind:value={$profileData.username} />
						{/snippet}
					</Form.Control>
					<Form.FieldErrors />
				</Form.Field>
			</Card.Content>
			<Card.Footer class="gap-3">
				<Button type="submit" size="sm" disabled={$profileSubmitting}>
					{$profileSubmitting ? 'Saving…' : 'Save'}
				</Button>
				{#if profileSaved}
					<p class="text-sm text-muted-foreground">Saved.</p>
				{/if}
			</Card.Footer>
		</form>
	</Card.Root>

	<Card.Root class="mt-6">
		<form class="contents" method="POST" action="?/changePassword" use:passwordEnhance>
			<Card.Header>
				<Card.Title>Change password</Card.Title>
				<Card.Description>Update the password used to sign in.</Card.Description>
			</Card.Header>
			<Card.Content>
				<Form.Field form={password} name="currentPassword">
					<Form.Control>
						{#snippet children({ props })}
							<Form.Label>Current password</Form.Label>
							<Input {...props} type="password" bind:value={$passwordData.currentPassword} />
						{/snippet}
					</Form.Control>
					<Form.FieldErrors />
				</Form.Field>
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
					{$passwordSubmitting ? 'Saving…' : 'Change password'}
				</Button>
				{#if passwordSaved}
					<p class="text-sm text-muted-foreground">Password changed.</p>
				{/if}
			</Card.Footer>
		</form>
	</Card.Root>
</div>
