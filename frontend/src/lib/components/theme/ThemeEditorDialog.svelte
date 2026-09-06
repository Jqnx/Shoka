<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import ColorTokenField from './ColorTokenField.svelte';
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import type { SubmitFunction } from '@sveltejs/kit';
	import {
		TOKEN_GROUPS,
		TOKEN_KEYS,
		isValidTokenValue,
		type ThemeTokens
	} from '$lib/themes/tokens';
	import { parseThemeCss } from '$lib/themes/import';
	import { theme as themeController } from '$lib/themes/theme.svelte';
	import type { Appearance, Theme } from '$lib/themes/types';

	let {
		open = $bindable(false),
		seed,
		editingId,
		onsaved
	}: {
		open?: boolean;
		seed: Theme;
		editingId?: string;
		onsaved?: (savedId: string) => void;
	} = $props();

	// The parent remounts this component (via `{#key}`) each time it opens, so
	// these initialisers run fresh against the current `seed`.
	// svelte-ignore state_referenced_locally
	let name = $state(editingId ? seed.name : `${seed.name} copy`);
	// svelte-ignore state_referenced_locally
	let appearance = $state<Appearance>(seed.appearance);
	// svelte-ignore state_referenced_locally
	let tokens = $state<ThemeTokens>({ ...seed.tokens });

	let pasteText = $state('');
	let pasteMsg = $state('');
	let copied = $state(false);
	let saving = $state(false);
	let error = $state<string | null>(null);

	// The whole app is the live preview - a container-scoped preview would not
	// reach this dialog's own portalled content. Revert on close / unmount.
	$effect(() => {
		if (!open) return;
		themeController.preview($state.snapshot(tokens), appearance);
		return () => themeController.restore();
	});

	const invalidCount = $derived(TOKEN_KEYS.filter((k) => !isValidTokenValue(k, tokens[k])).length);
	const canSave = $derived(name.trim().length > 0 && invalidCount === 0);

	const payload = $derived(
		JSON.stringify({
			...(editingId ? { id: editingId } : {}),
			name: name.trim(),
			appearance,
			tokens
		})
	);

	function applyPaste() {
		const result = parseThemeCss(pasteText, $state.snapshot(tokens));
		tokens = result.tokens;
		const kept = `Applied ${result.matched.length} token${result.matched.length === 1 ? '' : 's'}`;
		pasteMsg = result.rejected.length
			? `${kept}, skipped ${result.rejected.length} invalid`
			: result.matched.length
				? kept
				: 'No known tokens found';
	}

	async function copyCss() {
		const body = TOKEN_KEYS.map((k) => `  --${k}: ${tokens[k]};`).join('\n');
		try {
			await navigator.clipboard.writeText(`:root {\n${body}\n}`);
			copied = true;
			setTimeout(() => (copied = false), 1500);
		} catch {
			/* clipboard unavailable */
		}
	}

	const handleSave: SubmitFunction = () => {
		saving = true;
		error = null;
		return async ({ result }) => {
			saving = false;
			if (result.type === 'failure') {
				error = (result.data?.error as string | undefined) ?? 'Failed to save theme.';
				return;
			}
			if (result.type === 'success') {
				const savedId = (result.data?.saved as string) ?? '';
				if (themeController.active?.id === savedId) {
					themeController.apply({
						id: savedId,
						name: name.trim(),
						family: 'Custom',
						appearance,
						tokens: { ...tokens }
					});
				}
				open = false;
				await invalidateAll();
				onsaved?.(savedId);
			}
		};
	};
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="flex max-h-[90vh] flex-col sm:max-w-2xl">
		<Dialog.Header>
			<Dialog.Title>{editingId ? 'Edit theme' : 'New theme'}</Dialog.Title>
			<Dialog.Description>
				The whole app previews your changes live. Cancel reverts to the saved theme.
			</Dialog.Description>
		</Dialog.Header>

		<form
			method="POST"
			action="/settings?/saveTheme"
			use:enhance={handleSave}
			class="flex min-h-0 flex-1 flex-col gap-4"
		>
			<input type="hidden" name="payload" value={payload} />

			<div class="flex flex-wrap items-end gap-4">
				<div class="flex-1">
					<Label for="theme-name" class="text-xs">Name</Label>
					<Input id="theme-name" bind:value={name} maxlength={40} class="mt-1 h-8" />
				</div>
				<div class="flex items-center gap-2">
					<Switch
						id="theme-appearance"
						checked={appearance === 'dark'}
						onCheckedChange={(v) => (appearance = v ? 'dark' : 'light')}
					/>
					<Label for="theme-appearance" class="text-xs">Dark appearance</Label>
				</div>
			</div>

			<div class="flex min-h-0 flex-1 flex-col gap-5 overflow-y-auto pe-1">
				{#each TOKEN_GROUPS as group (group.label)}
					<Field.Group>
						<p class="text-xs font-semibold tracking-wide text-muted-foreground uppercase">
							{group.label}
						</p>
						{#each group.tokens as token (token)}
							<ColorTokenField {token} label={token} bind:value={tokens[token]} />
						{/each}
					</Field.Group>
				{/each}

				<Field.Group>
					<p class="text-xs font-semibold tracking-wide text-muted-foreground uppercase">Shape</p>
					<div class="flex items-center gap-2">
						<label
							for="tok-radius"
							class="w-44 shrink-0 truncate font-mono text-xs text-muted-foreground"
						>
							--radius
						</label>
						<Input id="tok-radius" class="h-8 font-mono text-xs" bind:value={tokens.radius} />
					</div>
				</Field.Group>

				<Field.Group>
					<p class="text-xs font-semibold tracking-wide text-muted-foreground uppercase">
						Import / export
					</p>
					<Textarea
						bind:value={pasteText}
						rows={3}
						placeholder={'Paste a :root { … } block to import token values'}
						class="font-mono text-xs"
					/>
					<div class="flex items-center gap-2">
						<Button type="button" variant="outline" size="sm" onclick={applyPaste}>Paste CSS</Button
						>
						<Button type="button" variant="outline" size="sm" onclick={copyCss}>
							{copied ? 'Copied' : 'Copy CSS'}
						</Button>
						{#if pasteMsg}
							<span class="text-xs text-muted-foreground">{pasteMsg}</span>
						{/if}
					</div>
				</Field.Group>
			</div>

			{#if invalidCount > 0}
				<p class="text-sm text-destructive">
					{invalidCount} token{invalidCount === 1 ? ' has' : 's have'} an invalid value.
				</p>
			{/if}
			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}

			<Dialog.Footer showCloseButton>
				<Button type="submit" disabled={saving || !canSave}>
					{saving ? 'Saving…' : 'Save theme'}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
