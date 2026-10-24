<script lang="ts">
	import FavoritePagesViewer from '$lib/components/favorites/FavoritePagesViewer.svelte';

	let { data } = $props();
</script>

<FavoritePagesViewer items={data.items} initialAt={data.at} initialSettings={data.settings} />
