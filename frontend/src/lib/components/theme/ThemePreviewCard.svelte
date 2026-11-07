<script lang="ts">
	import { Check, Copy, Pencil, Trash2 } from '@lucide/svelte';
	import { cn } from '$lib/utils.js';
	import { TOKEN_KEYS, isValidTokenValue } from '$lib/themes/tokens';
	import type { Theme } from '$lib/themes/types';

	let {
		theme,
		active = false,
		custom = false,
		onselect,
		onedit,
		onduplicate,
		ondelete
	}: {
		theme: Theme;
		active?: boolean;
		custom?: boolean;
		onselect: () => void;
		onedit?: () => void;
		onduplicate?: () => void;
		ondelete?: () => void;
	} = $props();

	// Only known, allowlisted values reach the inline style. `@theme inline` in
	// layout.css maps `--color-background` → `var(--background)`, so the Tailwind
	// utilities inside the swatch pick these up without touching the page.
	const styleVars = $derived(
		TOKEN_KEYS.filter((k) => isValidTokenValue(k, theme.tokens[k]))
			.map((k) => `--${k}: ${theme.tokens[k]}`)
			.join('; ')
	);
</script>

<div class="group relative">
	<button
		type="button"
		onclick={onselect}
		aria-pressed={active}
		class={cn(
			'block w-full overflow-hidden rounded-lg border-2 text-left transition-colors',
			active ? 'border-primary' : 'border-border hover:border-muted-foreground/40'
		)}
	>
		<div style={styleVars} class="flex h-28 bg-background text-foreground">
			<div class="w-9 shrink-0 border-r border-border bg-sidebar"></div>
			<div class="flex flex-1 flex-col gap-1.5 p-2.5">
				<div class="rounded border border-border bg-card p-1.5 text-[10px] text-card-foreground">
					Aa
				</div>
				<div class="text-[10px] text-muted-foreground">Muted text</div>
				<div class="mt-auto flex gap-1">
					<span
						class="rounded bg-primary px-2 py-1 text-[10px] font-medium text-primary-foreground"
					>
						Button
					</span>
					<span class="rounded bg-secondary px-2 py-1 text-[10px] text-secondary-foreground">
						Alt
					</span>
				</div>
			</div>
		</div>
		<div class="flex items-center justify-between gap-2 border-t border-border bg-card px-2.5 py-2">
			<span class="truncate text-xs font-medium text-card-foreground">{theme.name}</span>
			{#if active}
				<Check class="size-3.5 shrink-0 text-primary" />
			{/if}
		</div>
	</button>

	{#if custom}
		<div
			class="absolute end-1 top-1 flex gap-0.5 rounded-md bg-background/80 p-0.5 opacity-0 backdrop-blur transition-opacity group-hover:opacity-100 focus-within:opacity-100"
		>
			<button
				type="button"
				aria-label="Edit theme"
				class="rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
				onclick={onedit}
			>
				<Pencil class="size-3.5" />
			</button>
			<button
				type="button"
				aria-label="Duplicate theme"
				class="rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
				onclick={onduplicate}
			>
				<Copy class="size-3.5" />
			</button>
			<button
				type="button"
				aria-label="Delete theme"
				class="rounded p-1 text-muted-foreground hover:bg-muted hover:text-destructive"
				onclick={ondelete}
			>
				<Trash2 class="size-3.5" />
			</button>
		</div>
	{/if}
</div>
