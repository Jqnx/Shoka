<script lang="ts">
	import SignupForm from '$lib/components/signup-form.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<div class="flex h-screen w-full items-center justify-center px-4">
	<SignupForm {data} />
</div>
