<script lang="ts">
	import { tick } from 'svelte';
	import { setTagsInputItem, useTagsInputRoot } from './context.svelte.js';
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		value: itemValue,
		disabled: itemDisabled = false,
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLDivElement>> & {
		value: string;
		disabled?: boolean;
	} = $props();

	const root = useTagsInputRoot();
	const item = setTagsInputItem(
		root,
		() => itemValue,
		() => itemDisabled
	);

	$effect(() => {
		if (!ref) return;
		return root.registerElement(itemValue, ref);
	});

	async function handleKeydown(e: KeyboardEvent) {
		if (item.disabled) return;
		const index = item.index;

		if (e.key === 'ArrowLeft') {
			e.preventDefault();
			if (index > 0) root.focusItemAt(index - 1);
			return;
		}
		if (e.key === 'ArrowRight') {
			e.preventDefault();
			if (index < root.value.length - 1) root.focusItemAt(index + 1);
			else root.focusInput();
			return;
		}
		if (e.key === 'Backspace' || e.key === 'Delete') {
			e.preventDefault();
			root.removeTagAt(index);
			await tick();
			if (index < root.value.length) root.focusItemAt(index);
			else root.focusInput();
			return;
		}
		if (e.key === 'Escape') {
			e.preventDefault();
			root.focusInput();
		}
	}
</script>

<div
	bind:this={ref}
	role="button"
	data-slot="tags-input-item"
	data-active={item.active ? '' : undefined}
	data-disabled={item.disabled ? '' : undefined}
	tabindex={item.disabled ? -1 : 0}
	onkeydown={handleKeydown}
	onfocus={() => {
		if (!item.disabled) root.activeIndex = item.index;
	}}
	onblur={() => {
		if (root.activeIndex === item.index) root.activeIndex = null;
	}}
	class={cn(
		'bg-secondary text-secondary-foreground data-active:ring-ring/50 flex items-center gap-1 rounded-md py-0.5 ps-2 pe-1 text-xs font-medium outline-none data-active:ring-2 data-disabled:pointer-events-none data-disabled:opacity-50',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</div>
