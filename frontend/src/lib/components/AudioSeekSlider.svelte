<script lang="ts">
	import { Slider } from '$lib/components/ui/slider/index.js';
	import { player } from '$lib/player.svelte';

	let { class: className = '' }: { class?: string } = $props();

	let scrub = $state<number | null>(null);

	const max = $derived(Math.max(player.duration, 1));
	// Floor to the 1s step. A fractional value is not a valid step, so
	// bits-ui rewrites it through `value` and fires onValueChange, which
	// would pin `scrub` and freeze the thumb while playback advances.
	const position = $derived(Math.floor(Math.min(player.currentTime, max)));
</script>

<Slider
	type="single"
	value={scrub ?? position}
	{max}
	step={1}
	disabled={player.duration <= 0}
	ariaLabel="Seek"
	class={className}
	onValueChange={(value: number) => (scrub = value)}
	onValueCommit={(value: number) => {
		player.seek(value);
		scrub = null;
	}}
/>
