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
	favoritePages={data.favoritePages}
	initialPage={data.initialPage}
	initialCoverPage={data.archive.cover_page}
	canSetCover={data.canSetCover}
/>
