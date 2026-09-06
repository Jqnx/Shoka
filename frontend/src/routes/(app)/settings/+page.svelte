<script lang="ts">
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Form from '$lib/components/ui/form/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Plus } from '@lucide/svelte';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import { profileSchema } from '$lib/schemas/profile';
	import { changePasswordSchema } from '$lib/schemas/change-password';
	import { invalidateAll } from '$app/navigation';
	import ThemePreviewCard from '$lib/components/theme/ThemePreviewCard.svelte';
	import ThemeEditorDialog from '$lib/components/theme/ThemeEditorDialog.svelte';
	import { theme as themeController } from '$lib/themes/theme.svelte';
	import { DEFAULT_THEME_ID } from '$lib/themes/presets';
	import type { Theme } from '$lib/themes/types';

	let { data } = $props();

	let tab = $state('account');

	// --- Account: profile form ---
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

	// --- Account: password form ---
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

	// --- Appearance ---
	const activeId = $derived(themeController.active?.id ?? data.theme.id);

	let editorOpen = $state(false);
	// svelte-ignore state_referenced_locally
	let editorSeed = $state<Theme>(data.theme);
	let editorEditingId = $state<string | undefined>(undefined);
	// Bumped on every open so the editor remounts with fresh state from `seed`.
	let editorNonce = $state(0);

	function openEditor(seed: Theme, editingId: string | undefined) {
		editorSeed = seed;
		editorEditingId = editingId;
		editorNonce += 1;
		editorOpen = true;
	}

	async function post(action: string, body: Record<string, string>): Promise<boolean> {
		const fd = new FormData();
		for (const [k, v] of Object.entries(body)) fd.set(k, v);
		const res = await fetch(action, {
			method: 'POST',
			headers: { 'x-sveltekit-action': 'true' },
			body: fd
		});
		return res.ok;
	}

	function selectTheme(t: Theme) {
		if (t.id === activeId) return;
		void themeController.select(t, () => post('?/selectTheme', { themeId: t.id }));
	}

	function newTheme() {
		openEditor(themeController.active ?? data.theme, undefined);
	}

	function editTheme(t: Theme) {
		openEditor(t, t.id.replace(/^custom:/, ''));
	}

	function duplicateTheme(t: Theme) {
		openEditor(t, undefined);
	}

	async function deleteTheme(t: Theme) {
		if (!confirm(`Delete "${t.name}"?`)) return;
		const ok = await post('?/deleteTheme', { id: t.id.replace(/^custom:/, '') });
		if (!ok) return;
		if (themeController.active?.id === t.id) {
			const fallback = data.presetThemes.find((p) => p.id === DEFAULT_THEME_ID);
			if (fallback) themeController.apply(fallback);
		}
		await invalidateAll();
	}
</script>

<svelte:head>
	<title>Settings | Shoka</title>
</svelte:head>

<div class="mx-auto max-w-2xl p-4 sm:p-6">
	<Tabs.Root bind:value={tab} class="gap-6">
		<Tabs.List>
			<Tabs.Trigger value="account">Account</Tabs.Trigger>
			<Tabs.Trigger value="appearance">Appearance</Tabs.Trigger>
		</Tabs.List>

		<Tabs.Content value="account">
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
		</Tabs.Content>

		<Tabs.Content value="appearance">
			<div class="flex items-center justify-between gap-3">
				<div>
					<h1 class="text-2xl font-bold tracking-tight">Appearance</h1>
					<p class="mt-1 text-sm text-muted-foreground">Pick a theme or build your own.</p>
				</div>
				<Button size="sm" onclick={newTheme} class="gap-1.5">
					<Plus class="size-4" />
					New theme
				</Button>
			</div>

			{#if data.customThemes.length > 0}
				<h2 class="mt-8 mb-3 text-sm font-semibold text-muted-foreground">Your themes</h2>
				<div class="grid grid-cols-2 gap-3 sm:grid-cols-3">
					{#each data.customThemes as t (t.id)}
						<ThemePreviewCard
							theme={t}
							custom
							active={t.id === activeId}
							onselect={() => selectTheme(t)}
							onedit={() => editTheme(t)}
							onduplicate={() => duplicateTheme(t)}
							ondelete={() => deleteTheme(t)}
						/>
					{/each}
				</div>
			{/if}

			<h2 class="mt-8 mb-3 text-sm font-semibold text-muted-foreground">Built-in themes</h2>
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-3">
				{#each data.presetThemes as t (t.id)}
					<ThemePreviewCard theme={t} active={t.id === activeId} onselect={() => selectTheme(t)} />
				{/each}
			</div>

			{#key editorNonce}
				<ThemeEditorDialog bind:open={editorOpen} seed={editorSeed} editingId={editorEditingId} />
			{/key}
		</Tabs.Content>
	</Tabs.Root>
</div>
