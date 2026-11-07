<script lang="ts">
	import ArchiveLibraryPage from './ArchiveLibraryPage.svelte';
	import AudioLibraryPage from './AudioLibraryPage.svelte';

	let { data, form } = $props();
</script>

{#if data.kind === 'audio'}
	<AudioLibraryPage {data} />
{:else}
	<ArchiveLibraryPage {data} {form} />
{/if}
