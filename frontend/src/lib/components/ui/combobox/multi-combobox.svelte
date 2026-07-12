<script lang="ts" module>
	export type ComboboxOption = { value: string; label: string };
</script>

<script lang="ts">
	import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
	import XIcon from '@lucide/svelte/icons/x';
	import * as Command from '$lib/components/ui/command/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
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
		value: string[];
		onValueChange: (value: string[]) => void;
		placeholder?: string;
		searchPlaceholder?: string;
		emptyText?: string;
		class?: string;
	} = $props();

	let open = $state(false);
	let triggerRef = $state<HTMLDivElement>(null!);

	function labelFor(v: string) {
		return options.find((option) => option.value === v)?.label ?? v;
	}

	function toggle(optionValue: string) {
		if (value.includes(optionValue)) {
			onValueChange(value.filter((v) => v !== optionValue));
		} else {
			onValueChange([...value, optionValue]);
		}
	}

	function remove(optionValue: string, e?: Event) {
		e?.stopPropagation();
		onValueChange(value.filter((v) => v !== optionValue));
	}
</script>

<Popover.Root bind:open>
	<Popover.Trigger bind:ref={triggerRef}>
		{#snippet child({ props })}
			<div
				{...props}
				role="combobox"
				aria-expanded={open}
				tabindex={0}
				class={cn(
					'border-input dark:bg-input/30 has-data-[state=open]:border-ring has-data-[state=open]:ring-ring/50 flex min-h-8 w-full flex-wrap items-center gap-1.5 rounded-lg border bg-transparent px-2 py-1.5 text-sm shadow-xs outline-none has-data-[state=open]:ring-3',
					className
				)}
			>
				{#if value.length > 0}
					{#each value as v (v)}
						<Badge variant="secondary" class="gap-1 py-0.5 pr-1 font-normal">
							<span class="truncate">{labelFor(v)}</span>
							<button
								type="button"
								tabindex={-1}
								class="text-muted-foreground hover:text-foreground focus-visible:text-foreground flex size-3.5 shrink-0 items-center justify-center rounded-sm outline-none"
								onclick={(e) => remove(v, e)}
							>
								<XIcon class="size-3" />
							</button>
						</Badge>
					{/each}
				{:else}
					<span class="text-muted-foreground">{placeholder}</span>
				{/if}
				<ChevronsUpDownIcon class="ms-auto size-4 shrink-0 opacity-50" />
			</div>
		{/snippet}
	</Popover.Trigger>
	<Popover.Content class="w-(--bits-popover-anchor-width) min-w-48 p-0">
		<Command.Root>
			<Command.Input placeholder={searchPlaceholder} />
			<Command.List>
				<Command.Empty>{emptyText}</Command.Empty>
				<Command.Group>
					{#each options as option (option.value)}
						<Command.Item
							value={option.value}
							keywords={[option.label]}
							data-checked={value.includes(option.value)}
							onSelect={() => toggle(option.value)}
						>
							{option.label}
						</Command.Item>
					{/each}
				</Command.Group>
			</Command.List>
		</Command.Root>
	</Popover.Content>
</Popover.Root>
