<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { enhance } from '$app/forms';
	import { cn } from '$lib/utils.js';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import PageThumbnailGallery from '$lib/components/PageThumbnailGallery.svelte';
	import EditArchiveSheet from '$lib/components/EditArchiveSheet.svelte';
	import {
		BookOpen,
		ChevronLeft,
		Calendar,
		FileText,
		Tag,
		Users,
		Tv,
		Sword,
		Ellipsis,
		RotateCcw,
		Check,
		Heart
	} from '@lucide/svelte';

	let { data, form } = $props();
	const a = $derived(data.archive);

	let markingRead = $state(false);
	let markingUnread = $state(false);
	let favoriting = $state(false);

	// Archive.artists is just names (matching everywhere else metadata is
	// stored free-text) - resolve to the matching artist's id, if any, from
	// the already-loaded full list so the name can link to its detail page.
	// An artist not found there (e.g. a brand new one not yet reflected in
	// allArtists) just renders as plain text instead of a broken link.
	function artistId(name: string) {
		return data.allArtists.find((artist) => artist.name === name)?.id;
	}

	// ArchiveCard links here with a `from` param carrying the exact list page
	// (library + sort/filters/pagination) the user came from, so "back" can
	// return there as left. The single-archive endpoint doesn't say which
	// library this archive belongs to, so without `from` we can't know which
	// one to link/label with either - fall back to the first library, same
	// as the rest of the app until that's exposed.
	const fromParam = $derived(page.url.searchParams.get('from'));
	const fromLibraryId = $derived(fromParam?.match(/^\/([^/?]+)/)?.[1]);
	// Same resolved id used for both the back-link label and, further down,
	// which library's metadata source settings the edit sheet's fetch dialog
	// should respect (see MetadataFetchDialog's disabled-source picker).
	const libraryId = $derived(fromLibraryId ?? data.libraries[0]?.id ?? '');

	const archivesHref = $derived(fromParam ?? (data.libraries[0] ? `/${data.libraries[0].id}` : '/'));
	const archivesLabel = $derived(
		data.libraries.find((l) => l.id === libraryId)?.name ?? 'Archives'
	);

	function formatDate(iso: string | null) {
		if (!iso) return null;
		return new Date(iso).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	// Resume from the saved position (current_page is 0-based, same as the
	// reader's own indexing; the URL segment is 1-based) rather than always
	// starting over - unless it's already finished, where starting over from
	// the last page read would be a worse default than the beginning.
	const readHref = $derived(
		`/a/${a.id}/${a.progress && !a.progress.completed ? a.progress.current_page + 1 : 1}`
	);
	const readLabel = $derived(
		!a.progress ? 'Read' : a.progress.completed ? 'Read again' : 'Continue reading'
	);
</script>

<svelte:head>
	<title>{a.title} | Shoka</title>
</svelte:head>

<div class="mx-auto max-w-screen-xl px-4 py-8 sm:px-6">
	<!-- Back -->
	<a
		href={archivesHref}
		class="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
	>
		<ChevronLeft class="size-4" />
		{archivesLabel}
	</a>

	<div class="flex flex-col gap-8 md:flex-row">
		<!-- Cover -->
		<div class="w-full shrink-0 md:w-56 lg:w-64">
			<div class="relative aspect-[2/3] w-full overflow-hidden rounded-lg border border-border bg-muted">
				<img src="/api/archives/{a.id}/cover" alt={a.title} class="size-full object-cover" />
				{#if a.progress}
					<div class="absolute inset-x-0 top-0 h-1.5">
						<div
							class="h-full bg-primary transition-all"
							style="width: {Math.round((a.progress.current_page / a.page_count) * 100)}%"
						></div>
					</div>
				{/if}
			</div>

			<Button href={readHref} class="mt-4 w-full gap-2" size="lg">
				<BookOpen class="size-4" />
				{readLabel}
			</Button>

			{#if (form?.action === 'markRead' || form?.action === 'markUnread' || form?.action === 'toggleFavorite') && form.error}
				<p class="mt-1.5 text-center text-xs text-destructive">{form.error}</p>
			{/if}
		</div>

		<!-- Metadata -->
		<div class="min-w-0 flex-1">
			<div class="flex items-start justify-between gap-3">
				<h1 class="text-2xl leading-tight font-bold">{a.title}</h1>
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
						>
							<Heart
								class={cn('size-4', a.is_favorited && 'fill-red-500 stroke-red-500')}
							/>
						</Button>
					</form>
					<EditArchiveSheet
						archive={a}
						{libraryId}
						categories={data.categories}
						languages={data.languages}
						allArtists={data.allArtists}
						allTags={data.allTags}
						allCharacters={data.allCharacters}
						allParodies={data.allParodies}
					/>
					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							{#snippet child({ props })}
								<Button {...props} variant="outline" size="icon" aria-label="More actions">
									<Ellipsis class="size-4" />
								</Button>
							{/snippet}
						</DropdownMenu.Trigger>
						<DropdownMenu.Content align="end">
							<form
								method="POST"
								action="?/markRead"
								use:enhance={() => {
									markingRead = true;
									return async ({ update }) => {
										markingRead = false;
										await update();
									};
								}}
							>
								<input type="hidden" name="page_count" value={a.page_count} />
								<DropdownMenu.Item disabled={a.progress?.completed || markingRead}>
									{#snippet child({ props })}
										<button
											type="submit"
											{...props}
											class={cn(props.class as string, 'w-full')}
										>
											<Check />
											{markingRead ? 'Marking…' : 'Mark as read'}
										</button>
									{/snippet}
								</DropdownMenu.Item>
							</form>
							<form
								method="POST"
								action="?/markUnread"
								use:enhance={() => {
									markingUnread = true;
									return async ({ update }) => {
										markingUnread = false;
										await update();
									};
								}}
							>
								<DropdownMenu.Item disabled={!a.progress || markingUnread}>
									{#snippet child({ props })}
										<button
											type="submit"
											{...props}
											class={cn(props.class as string, 'w-full')}
										>
											<RotateCcw />
											{markingUnread ? 'Marking…' : 'Mark as unread'}
										</button>
									{/snippet}
								</DropdownMenu.Item>
							</form>
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</div>
			</div>

			{#if a.artists?.length}
				<p class="mt-1 text-base text-muted-foreground">
					{#each a.artists as artist, i (artist)}
						{@const id = artistId(artist)}
						{#if id !== undefined}
							<a href={resolve(`/artist/${id}`)} class="hover:text-foreground hover:underline"
								>{artist}</a
							>
						{:else}
							{artist}
						{/if}{i < a.artists.length - 1 ? ', ' : ''}
					{/each}
				</p>
			{/if}

			<div class="mt-3 flex flex-wrap gap-2">
				{#if a.category}
					<Badge variant="secondary">{a.category}</Badge>
				{/if}
				{#if a.language}
					<Badge variant="outline">{a.language}</Badge>
				{/if}
				<Badge variant="outline" class="gap-1">
					<FileText class="size-3" />
					{a.page_count} pages
				</Badge>
			</div>

			{#if a.summary}
				<p class="mt-4 text-sm leading-relaxed text-muted-foreground">{a.summary}</p>
			{/if}

			<Separator class="my-5" />

			<dl class="grid grid-cols-1 gap-y-3 text-sm sm:grid-cols-2">
				{#if a.release_date}
					<div class="flex items-start gap-2">
						<Calendar class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
						<div>
							<dt class="text-xs text-muted-foreground">Released</dt>
							<dd>{formatDate(a.release_date)}</dd>
						</div>
					</div>
				{/if}

				{#if a.circles?.length}
					<div class="flex items-start gap-2">
						<Users class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
						<div>
							<dt class="text-xs text-muted-foreground">Circle</dt>
							<dd>{a.circles.join(', ')}</dd>
						</div>
					</div>
				{/if}

				{#if a.parodies?.length}
					<div class="flex items-start gap-2">
						<Tv class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
						<div>
							<dt class="text-xs text-muted-foreground">Parody</dt>
							<dd class="flex flex-wrap gap-x-1">
								{#each a.parodies as parody, i (parody)}
									<a href={resolve(`/parody/${encodeURIComponent(parody)}`)} class="hover:underline"
										>{parody}{i < a.parodies.length - 1 ? ',' : ''}</a
									>
								{/each}
							</dd>
						</div>
					</div>
				{/if}

				{#if a.characters?.length}
					<div class="flex items-start gap-2">
						<Sword class="mt-0.5 size-4 shrink-0 text-muted-foreground" />
						<div>
							<dt class="text-xs text-muted-foreground">Characters</dt>
							<dd class="flex flex-wrap gap-x-1">
								{#each a.characters as character, i (character)}
									<a
										href={resolve(`/character/${encodeURIComponent(character)}`)}
										class="hover:underline">{character}{i < a.characters.length - 1 ? ',' : ''}</a
									>
								{/each}
							</dd>
						</div>
					</div>
				{/if}
			</dl>

			{#if a.tags?.length}
				<div class="mt-5">
					<div class="mb-2 flex items-center gap-1.5 text-xs text-muted-foreground">
						<Tag class="size-3.5" />
						Tags
					</div>
					<div class="flex flex-wrap gap-1.5">
						{#each a.tags as tag (tag)}
							<a href={resolve(`/tag/${encodeURIComponent(tag)}`)}>
								<Badge variant="secondary" class="cursor-pointer text-xs hover:bg-secondary/80"
									>{tag}</Badge
								>
							</a>
						{/each}
					</div>
				</div>
			{/if}

			<Separator class="my-5" />

			<p class="text-xs text-muted-foreground">
				Added {formatDate(a.created_at)}
				{#if a.updated_at !== a.created_at}
					· Updated {formatDate(a.updated_at)}
				{/if}
			</p>
		</div>
	</div>

	<Separator class="my-8" />

	<PageThumbnailGallery archiveId={a.id} pageCount={a.page_count} thumbsReady={a.thumbs_ready} />
</div>
