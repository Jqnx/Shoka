<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as TagsInput from '$lib/components/ui/tags-input/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { enhance } from '$app/forms';
	import { flushSync } from 'svelte';
	import { SlidersHorizontal } from '@lucide/svelte';
	import { sourceInfo } from '$lib/metadata-sources';
	import type { LibrarySource } from '$lib/types';

	let {
		source,
		error
	}: {
		source: LibrarySource;
		error?: string;
	} = $props();

	const info = $derived(sourceInfo(source.source));
	const formId = $derived(`source-form-${source.source}`);

	// Uncontrolled-after-mount: seeded from the prop, then edited locally
	// until the form is submitted (mirrors the artistInput pattern in
	// ArchivesSidebar.svelte).
	// svelte-ignore state_referenced_locally
	let enabled = $state(source.enabled);
	// svelte-ignore state_referenced_locally
	let cookies = $state(source.cookies ?? '');
	// svelte-ignore state_referenced_locally
	let apiKey = $state(source.api_key ?? '');
	// svelte-ignore state_referenced_locally
	let magazineBlocklist = $state(source.magazine_blocklist);
	// svelte-ignore state_referenced_locally
	let miscBlocklist = $state(source.misc_blocklist);

	let saving = $state(false);
	let settingsOpen = $state(false);
	let formEl = $state<HTMLFormElement>(null!);

	const hasFields = $derived(info.fields.length > 0);

	// The switch has no separate Save button — it submits itself immediately.
	// requestSubmit() reads the hidden "enabled" input's DOM value
	// synchronously, but Svelte's reactive DOM updates are batched into a
	// microtask - without flushSync, the form would still see the *previous*
	// value at submit time, one toggle behind what was actually clicked.
	function handleEnabledChange(value: boolean) {
		enabled = value;
		flushSync();
		formEl?.requestSubmit();
	}
</script>

<Card.Root class="gap-0 overflow-hidden p-0">
	<form
		bind:this={formEl}
		id={formId}
		method="POST"
		action="?/updateSource"
		use:enhance={() => {
			saving = true;
			return async ({ result, update }) => {
				saving = false;
				await update();
				if (result.type === 'success') settingsOpen = false;
			};
		}}
	>
		<input type="hidden" name="source" value={source.source} />
		<input type="hidden" name="enabled" value={String(enabled)} />
		<!--
			These mirror the settings-dialog fields but stay mounted even while
			the dialog is closed, unlike the dialog's own inputs (which live in a
			portal that unmounts on close). The enabled-only toggle below submits
			this form directly without opening the dialog - without these, that
			submission would carry no cookies/api_key/blocklist values at all,
			and the server would overwrite the stored ones with empty values.
		-->
		{#if info.fields.includes('cookies')}
			<input type="hidden" name="cookies" value={cookies} />
		{/if}
		{#if info.fields.includes('api_key')}
			<input type="hidden" name="api_key" value={apiKey} />
		{/if}
		{#if info.fields.includes('blocklists')}
			{#each magazineBlocklist as tag (tag)}
				<input type="hidden" name="magazine_blocklist" value={tag} />
			{/each}
			{#each miscBlocklist as tag (tag)}
				<input type="hidden" name="misc_blocklist" value={tag} />
			{/each}
		{/if}

		<div
			class="relative flex aspect-[4/3] items-center justify-center {info.image
				? ''
				: info.color}"
		>
			{#if info.image}
				<img src={info.image} alt="" class="size-full object-cover" />
			{:else}
				<info.icon class="size-10 text-white/90" />
			{/if}

			{#if hasFields}
				<Button
					type="button"
					variant="ghost"
					size="icon-sm"
					class="absolute top-2 right-2 text-white hover:bg-white/20 hover:text-white"
					onclick={() => (settingsOpen = true)}
				>
					<SlidersHorizontal />
					<span class="sr-only">Configure {info.label}</span>
				</Button>
			{/if}
		</div>

		<div class="flex items-center justify-between gap-2 p-3">
			<span class="truncate text-sm font-medium">{info.label}</span>
			<Switch checked={enabled} disabled={saving} onCheckedChange={handleEnabledChange} />
		</div>

		{#if error}
			<p class="px-3 pb-3 text-xs text-destructive">{error}</p>
		{/if}
	</form>
</Card.Root>

{#if hasFields}
	<Dialog.Root bind:open={settingsOpen}>
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>{info.label} settings</Dialog.Title>
				{#if info.description}
					<Dialog.Description>{info.description}</Dialog.Description>
				{/if}
			</Dialog.Header>

			<div class="flex flex-col gap-4">
				{#if info.fields.includes('cookies')}
					<div class="space-y-1.5">
						<Label for="{source.source}-cookies">Session cookie</Label>
						<Textarea
							id="{source.source}-cookies"
							bind:value={cookies}
							placeholder="ipb_member_id=...; ipb_pass_hash=..."
							class="font-mono text-xs"
							rows={3}
						/>
					</div>
				{/if}

				{#if info.fields.includes('api_key')}
					<div class="space-y-1.5">
						<Label for="{source.source}-api-key">API key</Label>
						<Input
							id="{source.source}-api-key"
							bind:value={apiKey}
							placeholder="API key"
							class="font-mono text-xs"
						/>
					</div>
				{/if}

				{#if info.fields.includes('blocklists')}
					<div class="space-y-1.5">
						<Label>Magazine blocklist</Label>
						<TagsInput.Root
							value={magazineBlocklist}
							onValueChange={(v) => (magazineBlocklist = v)}
							addOnPaste
						>
							{#each magazineBlocklist as tag (tag)}
								<TagsInput.Item value={tag}>
									<TagsInput.ItemText />
									<TagsInput.ItemDelete />
								</TagsInput.Item>
							{/each}
							<TagsInput.Input placeholder="Add magazine name…" />
						</TagsInput.Root>
					</div>

					<div class="space-y-1.5">
						<Label>Misc tag blocklist</Label>
						<TagsInput.Root
							value={miscBlocklist}
							onValueChange={(v) => (miscBlocklist = v)}
							addOnPaste
						>
							{#each miscBlocklist as tag (tag)}
								<TagsInput.Item value={tag}>
									<TagsInput.ItemText />
									<TagsInput.ItemDelete />
								</TagsInput.Item>
							{/each}
							<TagsInput.Input placeholder="Add tag…" />
						</TagsInput.Root>
					</div>
				{/if}
			</div>

			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}

			<Dialog.Footer showCloseButton>
				<Button type="submit" form={formId} disabled={saving}>
					{saving ? 'Saving…' : 'Save'}
				</Button>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{/if}
