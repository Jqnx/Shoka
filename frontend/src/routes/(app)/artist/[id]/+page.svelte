<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { goto, afterNavigate } from '$app/navigation';
	import { enhance, applyAction } from '$app/forms';
	import { cn } from '$lib/utils.js';
	import * as Pagination from '$lib/components/ui/pagination';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import ArchiveCard from '$lib/components/ArchiveCard.svelte';
	import ArchiveCover from '$lib/components/ArchiveCover.svelte';
	import EditArtistSheet from '$lib/components/EditArtistSheet.svelte';
	import ArtistCoverPicker from '$lib/components/ArtistCoverPicker.svelte';
	import { ChevronLeft, Ellipsis, Heart, Paintbrush, Star, Trash2 } from '@lucide/svelte';
	import { can } from '$lib/auth/permissions';

	let { data, form } = $props();

	const a = $derived(data.artist);

	const canEdit = $derived(can(page.data.user, { taxonomy: ['update'] }));
	const canDelete = $derived(can(page.data.user, { taxonomy: ['delete'] }));

	let favoriting = $state(false);
	let settingRating = $state(false);
	let hoveredStar = $state<number | null>(null);
	let deleteDialogOpen = $state(false);
	let deletingArtist = $state(false);

	// Reset only on a real page change: pagination uses `noScroll`, and
	// because the grid now sits below the header, scrolling to the top on a
	// page flip would hide the very thing the click changed.
	afterNavigate((nav) => {
		if (nav.type === 'goto' && nav.to?.url.pathname !== nav.from?.url.pathname) {
			window.scrollTo({ top: 0, behavior: 'smooth' });
		}
	});

	// A link's visible label is its hostname; a value that isn't a parseable
	// URL stays as written rather than rendering nothing.
	function linkLabel(url: string) {
		try {
			// Schemes without an authority (mailto:, data:) parse but yield an
			// empty hostname; fall back to the raw value so the link stays
			// visible and named.
			return new URL(url).hostname.replace(/^www\./, '') || url;
		} catch {
			return url;
		}
	}
</script>

<svelte:head>
	<title>{a.name} | Shoka</title>
</svelte:head>

<div class="mx-auto max-w-screen-2xl px-4 py-8 sm:px-6">
	<!-- Back -->
	<a
		href={data.backHref}
		class="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
	>
		<ChevronLeft class="size-4" />
		Artists
	</a>

	<div class="flex flex-col gap-8 md:flex-row">
		<!-- Cover -->
		<div class="w-full shrink-0 md:w-56 lg:w-64">
			<div
				class="relative aspect-[2/3] w-full overflow-hidden rounded-lg border border-border bg-muted"
			>
				{#if a.cover}
					<ArchiveCover
						archive={{
							id: a.cover.archive_id,
							cover_page: a.cover.page,
							thumbs_ready: a.cover.thumbs_ready
						}}
						alt={a.name}
						class="size-full object-cover"
					/>
				{:else}
					<div class="absolute inset-0 flex items-center justify-center">
						<Paintbrush class="size-10 text-muted-foreground/40" />
					</div>
				{/if}
			</div>

			{#if canEdit}
				<ArtistCoverPicker artistId={a.id} currentArchiveId={a.cover?.archive_id} />
			{/if}
		</div>

		<!-- Metadata -->
		<div class="min-w-0 flex-1">
			<div class="flex items-start justify-between gap-3">
				<h1 class="text-2xl leading-tight font-bold">{a.name}</h1>
				<div class="flex shrink-0 items-center gap-1.5">
					<form
						method="POST"
						action="?/toggleFavorite"
						use:enhance={() => {
							favoriting = true;
							return async ({ update }) => {
								favoriting = false;
								await update();
							};
						}}
					>
						<input type="hidden" name="favorited" value={a.is_favorited} />
						<Button
							type="submit"
							variant="outline"
							size="icon"
							disabled={favoriting}
							aria-label={a.is_favorited ? 'Remove from favorites' : 'Add to favorites'}
							aria-pressed={a.is_favorited}
						>
							<Heart class={cn('size-4', a.is_favorited && 'fill-red-500 stroke-red-500')} />
						</Button>
					</form>
					{#if canEdit}
						<EditArtistSheet artist={a} />
					{/if}
					{#if canDelete}
						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								{#snippet child({ props })}
									<Button {...props} variant="outline" size="icon" aria-label="More actions">
										<Ellipsis class="size-4" />
									</Button>
								{/snippet}
							</DropdownMenu.Trigger>
							<DropdownMenu.Content align="end" class="w-auto min-w-44">
								<DropdownMenu.Item variant="destructive" onSelect={() => (deleteDialogOpen = true)}>
									<Trash2 />
									Delete
								</DropdownMenu.Item>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					{/if}

					<AlertDialog.Root bind:open={deleteDialogOpen}>
						<AlertDialog.Content>
							<AlertDialog.Header>
								<AlertDialog.Title>Delete "{a.name}"?</AlertDialog.Title>
								<AlertDialog.Description>
									This removes the artist and its links from Shoka; archives are untouched.
								</AlertDialog.Description>
							</AlertDialog.Header>
							<form
								method="POST"
								action="?/deleteArtist"
								use:enhance={() => {
									deletingArtist = true;
									return async ({ result }) => {
										deletingArtist = false;
										await applyAction(result);
									};
								}}
							>
								<AlertDialog.Footer>
									<AlertDialog.Cancel type="button">Cancel</AlertDialog.Cancel>
									<AlertDialog.Action type="submit" variant="destructive" disabled={deletingArtist}>
										{deletingArtist ? 'Deleting…' : 'Delete'}
									</AlertDialog.Action>
								</AlertDialog.Footer>
							</form>
						</AlertDialog.Content>
					</AlertDialog.Root>
				</div>
			</div>

			<div
				role="group"
				aria-label="Rating"
				class="mt-2 flex items-center gap-0.5"
				onmouseleave={() => (hoveredStar = null)}
			>
				{#each [1, 2, 3, 4, 5] as n (n)}
					<form
						method="POST"
						action={a.rating === n ? '?/clearRating' : '?/setRating'}
						use:enhance={() => {
							settingRating = true;
							return async ({ update }) => {
								settingRating = false;
								await update();
							};
						}}
					>
						{#if a.rating !== n}
							<input type="hidden" name="rating" value={n} />
						{/if}
						<button
							type="submit"
							disabled={settingRating}
							class="p-0.5 text-muted-foreground transition-colors hover:text-amber-400 disabled:opacity-50"
							aria-pressed={a.rating === n}
							aria-label={a.rating === n
								? `Clear rating (${n} star${n === 1 ? '' : 's'})`
								: `Rate ${n} star${n === 1 ? '' : 's'}`}
							onmouseenter={() => (hoveredStar = n)}
						>
							<Star
								class={cn(
									'size-5',
									(hoveredStar ?? a.rating ?? 0) >= n
										? 'fill-amber-400 stroke-amber-400'
										: 'fill-none'
								)}
							/>
						</button>
					</form>
				{/each}
			</div>

			<dl class="mt-4 grid grid-cols-1 gap-x-6 gap-y-2 text-sm sm:grid-cols-[7rem_1fr]">
				{#if (a.aliases ?? []).length}
					<dt class="text-muted-foreground">Aliases</dt>
					<dd class="flex flex-wrap gap-1.5">
						{#each a.aliases ?? [] as alias (alias)}
							<Badge variant="secondary">{alias}</Badge>
						{/each}
					</dd>
				{/if}

				{#if (a.circles ?? []).length}
					<dt class="text-muted-foreground">Circles</dt>
					<dd class="flex flex-wrap gap-x-2">
						{#each a.circles ?? [] as circle (circle.id)}
							<a
								href={resolve(`/circle/${encodeURIComponent(circle.name)}`)}
								class="hover:underline"
							>
								{circle.name}
							</a>
						{/each}
					</dd>
				{/if}

				{#if (a.urls ?? []).length}
					<dt class="text-muted-foreground">Links</dt>
					<dd class="flex flex-wrap gap-x-2">
						{#each a.urls ?? [] as url (url)}
							<!-- eslint-disable svelte/no-navigation-without-resolve -- external artist link, not an app route -->
							<a
								href={url}
								target="_blank"
								rel="noopener noreferrer"
								class="text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
							>
								{linkLabel(url)}
							</a>
							<!-- eslint-enable svelte/no-navigation-without-resolve -->
						{/each}
					</dd>
				{/if}
			</dl>

			{#if form && 'error' in form && form.error}
				<p class="mt-3 text-sm text-destructive">{form.error}</p>
			{/if}
		</div>
	</div>

	<Separator class="my-8" />

	<div class="mb-4 flex items-center gap-2">
		<h2 class="text-lg font-semibold">Archives</h2>
		<span class="text-sm text-muted-foreground">({data.total})</span>
	</div>

	{#if data.archives.length === 0}
		<p class="text-muted-foreground">No archives found.</p>
	{:else}
		<div
			class="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8"
		>
			{#each data.archives as archive (archive.id)}
				<ArchiveCard {archive} />
			{/each}
		</div>
	{/if}

	{#if data.total > data.limit}
		<div class="mt-8">
			<Pagination.Root
				count={data.total}
				perPage={data.limit}
				page={data.page}
				onPageChange={(p) => {
					const params = new URLSearchParams(page.url.searchParams);
					params.set('page', String(p));
					goto(resolve(`/artist/${a.id}?${params}`), { noScroll: true });
				}}
			>
				{#snippet children({ pages })}
					<Pagination.Content>
						<Pagination.Item>
							<Pagination.Previous />
						</Pagination.Item>
						{#each pages as p (p.key)}
							{#if p.type === 'ellipsis'}
								<Pagination.Item>
									<Pagination.Ellipsis />
								</Pagination.Item>
							{:else}
								<Pagination.Item>
									<Pagination.Link page={p} isActive={data.page === p.value} />
								</Pagination.Item>
							{/if}
						{/each}
						<Pagination.Item>
							<Pagination.Next />
						</Pagination.Item>
					</Pagination.Content>
				{/snippet}
			</Pagination.Root>
		</div>
	{/if}
</div>
