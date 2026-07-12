<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import { SlidersHorizontal } from '@lucide/svelte';

	const sidebar = Sidebar.useSidebar();
</script>

<Button variant="outline" size="sm" onclick={() => sidebar.toggle()} class="gap-1.5">
	<SlidersHorizontal class="size-3.5" />
</Button>
