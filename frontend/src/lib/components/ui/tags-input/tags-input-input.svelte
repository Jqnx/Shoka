<script lang="ts">
	import { useTagsInputRoot } from './context.svelte.js';
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLInputAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		value = $bindable(''),
		class: className,
		...restProps
	}: WithElementRef<Omit<HTMLInputAttributes, 'value'>, HTMLInputElement> & { value?: string } =
		$props();

	const root = useTagsInputRoot();

	$effect(() => {
		root.inputRef = ref;
		return () => {
			if (root.inputRef === ref) root.inputRef = null;
		};
	});

	function commit() {
		if (root.addTag(value)) {
			value = '';
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			commit();
			return;
		}
		if (e.key === 'Tab' && root.addOnTab && value.trim()) {
			e.preventDefault();
			commit();
			return;
		}
		if ((e.key === 'Backspace' || e.key === 'ArrowLeft') && value === '' && root.value.length > 0) {
			e.preventDefault();
			root.focusLast();
		}
	}

	function handlePaste(e: ClipboardEvent) {
		if (!root.addOnPaste) return;
		const text = e.clipboardData?.getData('text');
		if (!text) return;
		e.preventDefault();
		root.addFromDelimited(text);
	}
</script>

<input
	bind:this={ref}
	bind:value
	data-slot="tags-input-input"
	disabled={root.disabled}
	class={cn(
		'placeholder:text-muted-foreground min-w-24 flex-1 bg-transparent text-sm outline-none disabled:cursor-not-allowed',
		className
	)}
	onkeydown={handleKeydown}
	onpaste={handlePaste}
	{...restProps}
/>
