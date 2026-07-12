<script lang="ts">
	import { useTagsInputItem } from './context.svelte.js';
	import { cn, type WithElementRef } from '$lib/utils.js';
	import XIcon from '@lucide/svelte/icons/x';
	import type { HTMLButtonAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLButtonAttributes> = $props();

	const item = useTagsInputItem();
</script>

<button
	bind:this={ref}
	type="button"
	data-slot="tags-input-item-delete"
	tabindex={-1}
	disabled={item.disabled}
	class={cn(
		'text-muted-foreground hover:text-foreground focus-visible:text-foreground flex size-3.5 shrink-0 items-center justify-center rounded-sm outline-none disabled:pointer-events-none',
		className
	)}
	onclick={(e) => {
		e.stopPropagation();
		item.root.removeTagAt(item.index);
	}}
	{...restProps}
>
	{#if children}
		{@render children()}
	{:else}
		<XIcon class="size-3" />
	{/if}
</button>
