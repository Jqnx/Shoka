<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Plus } from '@lucide/svelte';
	import { invalidateAll } from '$app/navigation';
	import ThemePreviewCard from '$lib/components/theme/ThemePreviewCard.svelte';
	import ThemeEditorDialog from '$lib/components/theme/ThemeEditorDialog.svelte';
	import { theme as themeController } from '$lib/themes/theme.svelte';
	import { DEFAULT_THEME_ID } from '$lib/themes/presets';
	import type { Theme } from '$lib/themes/types';

	let { data } = $props();

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
	<title>Appearance | Shoka</title>
</svelte:head>

<div>
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
</div>

{#key editorNonce}
	<ThemeEditorDialog bind:open={editorOpen} seed={editorSeed} editingId={editorEditingId} />
{/key}
