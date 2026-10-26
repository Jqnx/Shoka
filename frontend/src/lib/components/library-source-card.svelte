<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as TagsInput from '$lib/components/ui/tags-input/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { enhance } from '$app/forms';
	import { flushSync } from 'svelte';
	import { ArrowDown, ArrowUp, GripVertical, Pin, SlidersHorizontal } from '@lucide/svelte';
	import { sourceInfo } from '$lib/metadata-sources';
	import type { LibrarySource } from '$lib/types';

	let {
		source,
		error,
		isRemote,
		canMoveUp,
		canMoveDown,
		onMoveUp,
		onMoveDown,
		onDragStart,
		onDrop
	}: {
		source: LibrarySource;
		error?: string;
		isRemote: boolean;
		canMoveUp: boolean;
		canMoveDown: boolean;
		onMoveUp: () => void;
		onMoveDown: () => void;
		onDragStart: () => void;
		onDrop: () => void;
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

	function handleDragStart(event: DragEvent) {
		if (event.dataTransfer) {
			event.dataTransfer.effectAllowed = 'move';
			// Firefox refuses to start an HTML5 drag unless dragstart sets
			// some data; the payload itself is unused.
			event.dataTransfer.setData('text/plain', source.source);
		}

		onDragStart();
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		onDrop();
	}
</script>

<div
	class="flex items-center gap-3 rounded-lg border bg-card p-2"
	role="listitem"
	draggable={isRemote ? true : undefined}
	ondragstart={isRemote ? handleDragStart : undefined}
	ondragover={isRemote ? (event: DragEvent) => event.preventDefault() : undefined}
	ondrop={isRemote ? handleDrop : undefined}
>
	<div class="flex size-6 shrink-0 items-center justify-center text-muted-foreground">
		{#if isRemote}
			<GripVertical class="size-4 cursor-grab" aria-hidden="true" />
		{:else}
			<Pin class="size-4" aria-hidden="true" />
		{/if}
	</div>

	<div
		class="flex size-9 shrink-0 items-center justify-center overflow-hidden rounded-md {info.image
			? ''
			: info.color}"
	>
		{#if info.image}
			<img src={info.image} alt="" class="size-full object-cover" />
		{:else}
			<info.icon class="size-5 text-white/90" />
		{/if}
	</div>

	<div class="min-w-0 flex-1">
		<p class="truncate text-sm font-medium">{info.label}</p>
	</div>

	<form
		bind:this={formEl}
		id={formId}
		method="POST"
		action="?/updateSource"
		class="flex shrink-0 items-center gap-1"
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

		{#if isRemote}
			<Button type="button" variant="ghost" size="icon-sm" disabled={!canMoveUp} onclick={onMoveUp}>
				<ArrowUp />
				<span class="sr-only">Move {info.label} up</span>
			</Button>
			<Button
				type="button"
				variant="ghost"
				size="icon-sm"
				disabled={!canMoveDown}
				onclick={onMoveDown}
			>
				<ArrowDown />
				<span class="sr-only">Move {info.label} down</span>
			</Button>
		{/if}

		{#if hasFields}
			<Button type="button" variant="ghost" size="icon-sm" onclick={() => (settingsOpen = true)}>
				<SlidersHorizontal />
				<span class="sr-only">Configure {info.label}</span>
			</Button>
		{/if}

		<Switch checked={enabled} disabled={saving} onCheckedChange={handleEnabledChange} />
	</form>
</div>

{#if error}
	<p class="px-2 text-xs text-destructive">{error}</p>
{/if}

{#if hasFields}
	<Dialog.Root bind:open={settingsOpen}>
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>{info.label} settings</Dialog.Title>
				{#if info.help}
					<Dialog.Description>
						{info.help.text}
						<!-- eslint-disable svelte/no-navigation-without-resolve -- external settings page, not an app route -->
						<a href={info.help.href} target="_blank" rel="noopener noreferrer">{info.help.href}</a>
						<!-- eslint-enable svelte/no-navigation-without-resolve -->
					</Dialog.Description>
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
