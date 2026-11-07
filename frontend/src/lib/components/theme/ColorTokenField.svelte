<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { toHex, toOklchString } from '$lib/themes/color';

	let {
		token,
		label,
		value = $bindable('')
	}: { token: string; label: string; value: string } = $props();

	// `<input type="color">` speaks `#rrggbb` only; the text input keeps the raw
	// (canonical `oklch()`) value.
	const swatch = $derived(toHex(value));

	function onSwatch(event: Event) {
		const hex = (event.currentTarget as HTMLInputElement).value;
		value = toOklchString(hex) ?? hex;
	}
</script>

<div class="flex items-center gap-2">
	<label
		for={`tok-${token}`}
		class="w-44 shrink-0 truncate font-mono text-xs text-muted-foreground"
	>
		--{token}
	</label>
	<input
		type="color"
		aria-label={`${label} color picker`}
		value={swatch}
		oninput={onSwatch}
		class="size-8 shrink-0 cursor-pointer rounded border border-border bg-transparent p-0.5"
	/>
	<Input id={`tok-${token}`} class="h-8 font-mono text-xs" bind:value />
</div>
