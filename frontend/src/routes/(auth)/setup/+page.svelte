<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import * as Form from '$lib/components/ui/form/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import { enhance as appEnhance, applyAction } from '$app/forms';
	import { goto, invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { signupSchema } from '$lib/schemas/signup';
	import type { ArchiveListResponse } from '$lib/types';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const step = $derived(data.user === null ? 1 : data.libraries.length === 0 ? 2 : 3);
	const scanDone = $derived(
		data.libraries.length > 0 && data.libraries.every((l) => l.last_scanned_at !== undefined)
	);

	// Step 1 — the account form. Server truth decides which step renders.
	// svelte-ignore state_referenced_locally
	const account = superForm(data.form, { validators: zod4Client(signupSchema) });
	const { form: accountForm, enhance: accountEnhance } = account;

	const STEPS = ['Account', 'Library', 'First scan'] as const;

	// Step 2 — library form.
	function typeLabel(type: string) {
		return type.charAt(0).toUpperCase() + type.slice(1);
	}

	const DEFAULT_TYPE = 'doujinshi';

	function defaultType() {
		return (
			data.libraryTypes.find((t) => t.type === DEFAULT_TYPE)?.type ??
			data.libraryTypes[0]?.type ??
			DEFAULT_TYPE
		);
	}

	let creating = $state(false);
	let createError = $state<string | null>(null);
	let libraryType = $state(defaultType());
	const libraryTypeLabel = $derived(typeLabel(libraryType));

	// Step 3 — poll until the initial scan attempt has been stamped, then read
	// the archive count once.
	let archiveCount = $state<number | null>(null);

	$effect(() => {
		if (step !== 3 || scanDone) return;
		const t = setInterval(() => invalidateAll(), 2000);
		return () => clearInterval(t);
	});

	$effect(() => {
		if (step !== 3 || !scanDone || archiveCount !== null) return;
		fetch(`/api/archives?library_id=${data.libraries[0].id}&limit=1`)
			.then((r) => (r.ok ? r.json() : null))
			.then((b: ArchiveListResponse | null) => {
				if (b) archiveCount = b.total;
			});
	});
</script>

<svelte:head>
	<title>Setup | Shoka</title>
</svelte:head>

<div class="flex min-h-screen w-full items-center justify-center px-4 py-12">
	<div class="w-full max-w-md space-y-6">
		<div class="space-y-2 text-center">
			<h1 class="text-xl font-semibold text-foreground">Set up Shoka</h1>
			<nav aria-label="Setup progress" class="flex items-center justify-center gap-2 text-xs">
				{#each STEPS as label, i (label)}
					{#if i > 0}
						<span class="text-muted-foreground">·</span>
					{/if}
					<span
						class={i + 1 === step ? 'text-foreground' : 'text-muted-foreground'}
						aria-current={i + 1 === step ? 'step' : undefined}>{label}</span
					>
				{/each}
			</nav>
		</div>

		{#if step === 1}
			<Card.Root>
				<Card.Header>
					<Card.Title>Create your account</Card.Title>
					<Card.Description>
						The first account owns this instance and becomes its administrator.
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<form method="POST" action="?/createAccount" use:accountEnhance>
						<Form.Field form={account} name="username">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Username</Form.Label>
									<Input {...props} bind:value={$accountForm.username} />
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>
						<Form.Field form={account} name="email">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Email</Form.Label>
									<Input {...props} bind:value={$accountForm.email} />
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>
						<Form.Field form={account} name="password">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Password</Form.Label>
									<Input {...props} bind:value={$accountForm.password} type="password" />
								{/snippet}
							</Form.Control>
							<Form.Description>Must be at least 8 characters long.</Form.Description>
							<Form.FieldErrors />
						</Form.Field>
						<Form.Field form={account} name="confirmPassword">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Confirm Password</Form.Label>
									<Input {...props} bind:value={$accountForm.confirmPassword} type="password" />
								{/snippet}
							</Form.Control>
							<Form.FieldErrors />
						</Form.Field>
						<Form.Button class="w-full">Create admin account</Form.Button>
					</form>
				</Card.Content>
			</Card.Root>
		{:else if step === 2}
			<Card.Root>
				<Card.Header>
					<Card.Title>Add your first library</Card.Title>
					<Card.Description>
						Shoka will scan this folder for archives and keep watching it for changes.
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<form
						method="POST"
						action="?/createLibrary"
						use:appEnhance={() => {
							creating = true;
							createError = null;
							return async ({ result }) => {
								creating = false;
								if (result.type === 'failure') {
									createError =
										(result.data?.error as string | undefined) ?? 'Failed to create library.';
									return;
								}
								if (result.type === 'success') {
									await invalidateAll();
									return;
								}
								await applyAction(result);
							};
						}}
					>
						<Field.Group>
							<Field.Field>
								<Field.Label for="setup-library-name">Name</Field.Label>
								<Input id="setup-library-name" name="name" placeholder="My Library" required />
							</Field.Field>
							<Field.Field>
								<Field.Label for="setup-library-path">Path</Field.Label>
								<Input id="setup-library-path" name="path" placeholder="/data/library" required />
							</Field.Field>
							<Field.Field>
								<Field.Label for="setup-library-type">Content type</Field.Label>
								<Select.Root type="single" name="type" bind:value={libraryType}>
									<Select.Trigger id="setup-library-type" class="w-full">
										{libraryTypeLabel}
									</Select.Trigger>
									<Select.Content>
										{#each data.libraryTypes as type (type.type)}
											<Select.Item
												value={type.type}
												label={typeLabel(type.type)}
												disabled={!type.supported}
											>
												{typeLabel(type.type)}{!type.supported ? ' (coming soon)' : ''}
											</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</Field.Field>
							{#if createError}
								<p class="text-sm text-destructive">{createError}</p>
							{/if}
						</Field.Group>

						<Button type="submit" class="mt-4 w-full" disabled={creating}>
							{creating ? 'Creating…' : 'Create library'}
						</Button>
					</form>
				</Card.Content>
			</Card.Root>
		{:else}
			<Card.Root>
				<Card.Header>
					<Card.Title>{scanDone ? 'Scan finished' : 'Scanning your library…'}</Card.Title>
					<Card.Description>{data.libraries[0].name}</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-2 text-sm">
					{#if scanDone}
						<p>
							{archiveCount === null ? 'Your library is ready.' : `${archiveCount} archives found.`}
						</p>
						{#if archiveCount === 0}
							<p class="text-muted-foreground">
								Drop files into the folder — the watcher picks them up.
							</p>
						{/if}
					{:else}
						<p class="text-muted-foreground">
							Shoka is reading the folder. A large library can take a moment.
						</p>
					{/if}
				</Card.Content>
				<Card.Footer class="gap-2">
					{#if scanDone}
						<Button class="w-full" onclick={() => goto(resolve('/'))}>Finish</Button>
					{:else}
						<Button variant="secondary" class="w-full" onclick={() => goto(resolve('/'))}>
							Skip for now
						</Button>
					{/if}
				</Card.Footer>
			</Card.Root>
		{/if}
	</div>
</div>
