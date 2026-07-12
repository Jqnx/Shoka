<script lang="ts">
	import { useTagsInputRoot } from './context.svelte.js';
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLButtonAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLButtonAttributes> = $props();

	const root = useTagsInputRoot();
</script>

<button
	bind:this={ref}
	type="button"
	data-slot="tags-input-clear"
	disabled={root.disabled || root.value.length === 0}
	class={cn(
		'text-muted-foreground hover:text-foreground text-xs outline-none disabled:pointer-events-none disabled:opacity-50',
		className
	)}
	onclick={() => root.clear()}
	{...restProps}
>
	{@render children?.()}
</button>
