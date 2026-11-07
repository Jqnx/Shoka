<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { theme } from '$lib/themes/theme.svelte';

	let { children, data } = $props();

	// Record the SSR-resolved theme so client-side switching has a baseline to
	// revert to. Deliberately does NOT touch the DOM - the injected `<style>`
	// block is already correct, so painting here would only risk a flash.
	$effect(() => {
		theme.hydrate(data.theme);
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}
