<script lang="ts">
	import { cn } from '$lib/utils.js';

	// Single-line label that reveals its tail by scrolling when hovered.
	// Distance is measured after mount and on resize (the player bar follows
	// the sidebar width), so only genuinely truncated text moves at all.
	let { text, class: className = '' }: { text: string; class?: string } = $props();

	let root = $state<HTMLSpanElement>();
	let inner = $state<HTMLSpanElement>();
	let distance = $state(0);

	$effect(() => {
		const rootEl = root;
		const innerEl = inner;
		if (!rootEl || !innerEl) return;

		const measure = () => {
			distance = Math.max(0, innerEl.offsetWidth - rootEl.clientWidth);
		};
		measure();

		const observer = new ResizeObserver(measure);
		observer.observe(rootEl);
		observer.observe(innerEl);
		return () => observer.disconnect();
	});

	// ~60px per second, with a floor so a small overflow is still readable.
	const duration = $derived(Math.max(600, Math.round((distance / 60) * 1000)));
</script>

<span
	bind:this={root}
	class={cn('scroll-title flex overflow-hidden whitespace-nowrap', className)}
	style:--scroll-distance="-{distance}px"
	style:--scroll-duration="{duration}ms"
>
	<span bind:this={inner} class="scroll-text">{text}</span>
</span>

<style>
	.scroll-text {
		display: block;
		flex: none;
		transform: translateX(0);
		transition: transform 200ms ease-out;
	}

	.scroll-title:hover .scroll-text {
		transform: translateX(var(--scroll-distance, 0px));
		transition: transform var(--scroll-duration, 0ms) linear;
	}

	@media (prefers-reduced-motion: reduce) {
		.scroll-text,
		.scroll-title:hover .scroll-text {
			transform: none;
			transition: none;
		}
	}
</style>
