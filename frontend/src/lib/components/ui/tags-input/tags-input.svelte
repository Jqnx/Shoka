<script lang="ts">
	import { setTagsInputRoot } from './context.svelte.js';
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		value = $bindable([]),
		onValueChange,
		max,
		duplicate = false,
		delimiter = ',',
		addOnPaste = false,
		addOnTab = false,
		disabled = false,
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLDivElement>> & {
		value?: string[];
		onValueChange?: (value: string[]) => void;
		max?: number;
		duplicate?: boolean;
		delimiter?: string | RegExp;
		addOnPaste?: boolean;
		addOnTab?: boolean;
		disabled?: boolean;
	} = $props();

	const root = setTagsInputRoot({
		value: () => value,
		onValueChange: (next) => {
			value = next;
			onValueChange?.(next);
		},
		max: () => max,
		duplicate: () => duplicate,
		delimiter: () => delimiter,
		addOnPaste: () => addOnPaste,
		addOnTab: () => addOnTab,
		disabled: () => disabled
	});
</script>

<div
	bind:this={ref}
	data-slot="tags-input"
	data-disabled={disabled ? '' : undefined}
	class={cn(
		'border-input dark:bg-input/30 has-[input:focus-visible]:border-ring has-[input:focus-visible]:ring-ring/50 flex min-h-8 w-full flex-wrap items-center gap-1.5 rounded-lg border bg-transparent px-2 py-1.5 text-sm shadow-xs has-[input:focus-visible]:ring-3 data-disabled:pointer-events-none data-disabled:opacity-50',
		className
	)}
	onclick={(e) => {
		if (disabled) return;
		if (e.target === e.currentTarget) root.focusInput();
	}}
	{...restProps}
>
	{@render children?.()}
</div>
