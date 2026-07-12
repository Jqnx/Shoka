<script lang="ts" module>
	export type ComboboxOption = { value: string; label: string };
</script>

<script lang="ts">
	import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
	import { tick } from 'svelte';
	import * as Command from '$lib/components/ui/command/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';

	let {
		options,
		value,
		onValueChange,
		placeholder = 'Select...',
		searchPlaceholder = 'Search...',
		emptyText = 'No results found.',
		class: className
	}: {
		options: ComboboxOption[];
		value: string;
		onValueChange: (value: string) => void;
		placeholder?: string;
		searchPlaceholder?: string;
		emptyText?: string;
		class?: string;
	} = $props();

	let open = $state(false);
	let triggerRef = $state<HTMLButtonElement>(null!);

	const selectedLabel = $derived(options.find((option) => option.value === value)?.label);

	function closeAndFocusTrigger() {
		open = false;
		tick().then(() => triggerRef.focus());
	}
</script>

<Popover.Root bind:open>
	<Popover.Trigger bind:ref={triggerRef}>
		{#snippet child({ props })}
			<Button
				variant="outline"
				{...props}
				role="combobox"
				aria-expanded={open}
				class={cn('w-full justify-between font-normal', className)}
			>
				<span class="truncate">{selectedLabel || placeholder}</span>
				<ChevronsUpDownIcon class="ms-2 size-4 shrink-0 opacity-50" />
			</Button>
		{/snippet}
	</Popover.Trigger>
	<Popover.Content class="w-(--bits-popover-anchor-width) min-w-48 p-0">
		<Command.Root {value}>
			<Command.Input placeholder={searchPlaceholder} />
			<Command.List>
				<Command.Empty>{emptyText}</Command.Empty>
				<Command.Group>
					{#each options as option (option.value)}
						<Command.Item
							value={option.value}
							keywords={[option.label]}
							data-checked={value === option.value}
							onSelect={() => {
								onValueChange(option.value);
								closeAndFocusTrigger();
							}}
						>
							{option.label}
						</Command.Item>
					{/each}
				</Command.Group>
			</Command.List>
		</Command.Root>
	</Popover.Content>
</Popover.Root>
