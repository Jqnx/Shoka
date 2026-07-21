<script lang="ts">
	import Reader from '$lib/components/reader/Reader.svelte';

	let { data } = $props();
</script>

<Reader
	archiveId={data.archive.id}
	title={data.archive.title}
	pageCount={data.archive.page_count}
	backHref="/a/{data.archive.id}"
	initialSettings={data.settings}
	initialPage={data.initialPage}
/>
