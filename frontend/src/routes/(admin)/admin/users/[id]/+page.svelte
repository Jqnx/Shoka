<script lang="ts">
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { ADMIN_ROLE } from '$lib/auth/permissions';
	import { ArrowLeft } from '@lucide/svelte';

	let { data, form } = $props();

	const me = $derived(page.data.user);
	const isAdminTarget = $derived(data.target.role === ADMIN_ROLE);

	function initialsOf(name: string) {
		return name
			.split(' ')
			.map((part) => part[0])
			.join('')
			.slice(0, 2)
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
				{#if data.target.id === me?.id}
					<span class="text-xs font-normal text-muted-foreground">(you)</span>
				{/if}
				<Badge variant={isAdminTarget ? 'default' : 'secondary'}>
					{data.target.role ?? 'user'}
				</Badge>
				{#if data.target.banned}
					<Badge variant="destructive">banned</Badge>
				{/if}
			</div>
			<p class="truncate text-sm text-muted-foreground">{data.target.email}</p>
		</div>
	</div>

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
						await update({ reset: false });
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
