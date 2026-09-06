<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { CopyCheck } from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { coverSrc } from '$lib/types';

	let { data } = $props();

	// Ordered loosest-last: a higher per-point Hamming distance lets more
	// pairs match, trading precision for recall.
	const thresholdOptions = [
		{ value: '4', label: 'Strict (4)' },
		{ value: '6', label: 'Tight (6)' },
		{ value: '8', label: 'Balanced (8)' },
		{ value: '12', label: 'Loose (12)' },
		{ value: '16', label: 'Very loose (16)' }
	];

	const thresholdValue = $derived(String(data.threshold));
	const thresholdLabel = $derived(
		thresholdOptions.find((o) => o.value === thresholdValue)?.label ?? `Custom (${data.threshold})`
	);

	const totalArchives = $derived(
		data.groups.reduce((sum, group) => sum + group.archives.length, 0)
	);

	function libraryName(id: string) {
		return data.libraries.find((l) => l.id === id)?.name ?? 'Unknown library';
	}

	function setThreshold(value: string) {
		const params = new URLSearchParams(page.url.searchParams);
		params.set('threshold', value);
		goto(resolve(`${page.url.pathname}?${params}`), { noScroll: true, keepFocus: true });
	}

	// Send the user back here after they act on an archive, rather than to a
	// library listing they never came from.
	const from = $derived(encodeURIComponent(page.url.pathname + page.url.search));

	// p0 is the cover, the rest are that fraction through the archive. Points
	// the hash job hasn't covered are omitted from the map by the backend.
	const POINT_LABELS: Record<string, string> = {
		p0: 'cover',
		p25: '25%',
		p50: '50%',
		p75: '75%'
	};
</script>

<svelte:head>
	<title>Duplicates | Shoka Admin</title>
</svelte:head>

<div class="p-4 sm:p-6">
	<div class="flex flex-wrap items-start justify-between gap-3">
		<div>
			<h1 class="text-2xl font-bold tracking-tight">Duplicates</h1>
			<p class="mt-1 text-sm text-muted-foreground">
				Archives that look visually alike across all libraries, matched on perceptual hashes.
			</p>
		</div>

		<Select.Root type="single" value={thresholdValue} onValueChange={setThreshold}>
			<Select.Trigger class="h-8 w-[170px]">{thresholdLabel}</Select.Trigger>
			<Select.Content>
				<Select.Group>
					<Select.Label>Match threshold</Select.Label>
					{#each thresholdOptions as opt (opt.value)}
						<Select.Item value={opt.value} label={opt.label}>{opt.label}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
	</div>

	{#if data.error}
		<p class="mt-4 text-sm text-destructive">{data.error}</p>
	{/if}

	{#if data.groups.length === 0}
		<div
			class="mt-6 flex flex-col items-center justify-center rounded-xl border border-dashed border-border py-16 text-center"
		>
			<CopyCheck class="mb-3 size-10 text-muted-foreground/40" />
			<p class="text-base font-medium text-foreground">No duplicates found</p>
			<p class="mt-1 max-w-md text-sm text-muted-foreground">Nothing matched at this threshold.</p>
		</div>
	{:else}
		<p class="mt-4 text-sm text-muted-foreground">
			{data.groups.length}
			{data.groups.length === 1 ? 'group' : 'groups'} · {totalArchives} archives. Larger groups are chained
			matches, so treat them as progressively less certain.
		</p>

		<div class="mt-4 flex flex-col gap-4">
			{#each data.groups as group, i (group.archives[0].id)}
				<Card.Root>
					<Card.Header>
						<Card.Title class="text-base">
							Group {i + 1}
							<span class="ms-1 font-normal text-muted-foreground">
								· {group.archives.length} archives
							</span>
						</Card.Title>
					</Card.Header>
					<Card.Content>
						<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
							{#each group.archives as archive, j (archive.id)}
								<a
									href={resolve(`/a/${archive.id}?from=${from}`)}
									class="group flex flex-col gap-2 rounded-lg border border-border p-2 transition-colors hover:bg-accent/50"
								>
									<div class="relative aspect-[2/3] overflow-hidden rounded bg-muted">
										<img src={coverSrc(archive)} alt="" class="size-full object-cover" />
										<div class="absolute top-1 left-1">
											{#if j === 0}
												<Badge variant="secondary" class="h-5 px-1.5 text-[10px]">Reference</Badge>
											{:else}
												<Badge variant="outline" class="h-5 bg-background/90 px-1.5 text-[10px]">
													Δ{archive.distance}
												</Badge>
											{/if}
										</div>
									</div>

									<div class="min-w-0">
										<p class="line-clamp-2 text-xs font-medium group-hover:text-foreground">
											{archive.title}
										</p>
										<p class="mt-0.5 truncate text-[11px] text-muted-foreground">
											{libraryName(archive.library_id)}
										</p>
										{#if j > 0}
											<p class="mt-0.5 truncate text-[11px] text-muted-foreground/80">
												{Object.entries(archive.distances)
													.map(([point, d]) => `${POINT_LABELS[point] ?? point} ${d}`)
													.join(' · ')}
											</p>
										{/if}
									</div>
								</a>
							{/each}
						</div>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{/if}
</div>
