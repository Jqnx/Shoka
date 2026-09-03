<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';
	import { sourceInfo } from '$lib/metadata-sources';
	import type { ArchiveSourceLink } from '$lib/types';

	let { links }: { links: ArchiveSourceLink[] | null } = $props();
</script>

{#if links && links.length > 0}
	<div class="flex flex-wrap gap-2">
		<!-- Keyed on URL, not source: an archive can carry several links from
		     the same site (e.g. a two-volume release). -->
		{#each links as link (link.url)}
			{@const info = sourceInfo(link.source)}
			<Tooltip.Root>
				<Tooltip.Trigger>
					{#snippet child({ props })}
						<!-- eslint-disable svelte/no-navigation-without-resolve -- link.url is an external gallery URL, not an app route -->
						<a
							{...props}
							href={link.url}
							target="_blank"
							rel="noopener noreferrer"
							aria-label={info.label}
							class="flex size-7 items-center justify-center overflow-hidden rounded-md text-white/90 transition-opacity hover:opacity-80 {info.image
								? ''
								: info.color}"
						>
							{#if info.image}
								<img src={info.image} alt={info.label} class="size-full object-cover" />
							{:else}
								<info.icon class="size-4" />
							{/if}
						</a>
						<!-- eslint-enable svelte/no-navigation-without-resolve -->
					{/snippet}
				</Tooltip.Trigger>
				<Tooltip.Content>{info.label}</Tooltip.Content>
			</Tooltip.Root>
		{/each}
	</div>
{/if}
