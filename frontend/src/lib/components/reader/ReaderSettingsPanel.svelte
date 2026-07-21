<script lang="ts">
	import type {
		ReaderSettings,
		ViewMode,
		ReadingDirection,
		PageLayout,
		FitMode,
		ReaderBackground
	} from '$lib/types';
	import { Button } from '$lib/components/ui/button/index.js';

	let {
		settings,
		onChange
	}: {
		settings: ReaderSettings;
		onChange: (patch: Partial<ReaderSettings>) => void;
	} = $props();

	// Snippets can't take a generic type parameter, so `set` is typed loosely
	// here - each call site below passes a properly-typed callback, so the
	// actual `onChange` calls stay fully typed regardless.
	type Option = { value: string; label: string };
</script>

{#snippet segmented(label: string, value: string, options: Option[], set: (v: string) => void)}
	<div>
		<p class="mb-1.5 text-xs font-medium text-muted-foreground">{label}</p>
		<div class="flex gap-1">
			{#each options as opt (opt.value)}
				<Button
					type="button"
					size="sm"
					variant={value === opt.value ? 'secondary' : 'ghost'}
					class="flex-1"
					onclick={() => set(opt.value)}
				>
					{opt.label}
				</Button>
			{/each}
		</div>
	</div>
{/snippet}

<div class="flex flex-col gap-3">
	{@render segmented(
		'View mode',
		settings.view_mode,
		[
			{ value: 'paged', label: 'Paged' },
			{ value: 'continuous', label: 'Continuous' }
		],
		(v) => onChange({ view_mode: v as ViewMode })
	)}

	{#if settings.view_mode === 'paged'}
		{@render segmented(
			'Reading direction',
			settings.reading_direction,
			[
				{ value: 'rtl', label: 'Right to left' },
				{ value: 'ltr', label: 'Left to right' }
			],
			(v) => onChange({ reading_direction: v as ReadingDirection })
		)}

		{@render segmented(
			'Page layout',
			settings.page_layout,
			[
				{ value: 'single', label: 'Single' },
				{ value: 'double', label: 'Double' }
			],
			(v) => onChange({ page_layout: v as PageLayout })
		)}
	{/if}

	{@render segmented(
		'Fit',
		settings.fit_mode,
		[
			{ value: 'width', label: 'Width' },
			{ value: 'height', label: 'Height' },
			{ value: 'original', label: 'Original' }
		],
		(v) => onChange({ fit_mode: v as FitMode })
	)}

	{@render segmented(
		'Background',
		settings.background,
		[
			{ value: 'black', label: 'Black' },
			{ value: 'gray', label: 'Gray' },
			{ value: 'white', label: 'White' }
		],
		(v) => onChange({ background: v as ReaderBackground })
	)}
</div>
