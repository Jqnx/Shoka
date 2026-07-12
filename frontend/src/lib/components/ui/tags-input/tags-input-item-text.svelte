<script lang="ts">
	import { useTagsInputItem } from './context.svelte.js';
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLSpanElement>> = $props();

	const item = useTagsInputItem();
</script>

<span
	bind:this={ref}
	data-slot="tags-input-item-text"
	class={cn('truncate', className)}
	{...restProps}
>
	{item.value}
</span>
