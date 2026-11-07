<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { enhance } from '$app/forms';
	import { cn } from '$lib/utils.js';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import AudioImageGallery from '$lib/components/AudioImageGallery.svelte';
	import AudioWorkTracks from '$lib/components/AudioWorkTracks.svelte';
	import EditAudioSheet from '$lib/components/EditAudioSheet.svelte';
	import { audioCoverSrc } from '$lib/types';
	import { formatDuration } from '$lib/audio';
	import { AudioLines, ChevronLeft, Heart, Star } from '@lucide/svelte';
	import { can } from '$lib/auth/permissions';

	let { data, form } = $props();

	const a = $derived(data.audio);
	const defaultFolder = $derived(
		data.libraries.find((library) => library.id === a.library_id)?.default_audio_folder ?? ''
	);
	const canEdit = $derived(can(page.data.user, { audio: ['update'] }));

	let favoriting = $state(false);
	let settingRating = $state(false);
	let hoveredStar = $state<number | null>(null);
	let section = $state('tracks');

	const backHref = $derived(`/${a.library_id}` as `/${string}`);
	const languageName = $derived(
		a.language ? (data.languages.find((l) => l.code === a.language)?.name ?? a.language) : null
	);

	function artistId(name: string) {
		return data.allArtists.find((artist) => artist.name === name)?.id;
	}

	function formatDate(iso: string | null) {
		if (!iso) return null;
		return new Date(iso).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}
</script>

<svelte:head>
	<title>{a.title} | Shoka</title>
</svelte:head>

<div class="mx-auto max-w-screen-2xl px-4 py-8 sm:px-6">
	<a
		href={resolve(backHref)}
		class="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
	>
		<ChevronLeft class="size-4" />
		Library
	</a>

	<div class="flex flex-col gap-8 md:flex-row">
		<div class="w-full max-w-xs shrink-0 md:max-w-64 lg:max-w-[27rem]">
			<div class="relative w-full overflow-hidden rounded-lg border border-border bg-muted">
				{#if a.cover_ready}
					<img
						src={audioCoverSrc(a)}
						alt={a.title}
						width={a.cover_width || undefined}
						height={a.cover_height || undefined}
						class="block h-auto w-full"
					/>
				{:else}
					<div class="flex aspect-square items-center justify-center">
						<AudioLines class="size-10 text-muted-foreground/40" />
					</div>
				{/if}
			</div>
		</div>

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
							aria-pressed={a.is_favorited}
						>
							<Heart class={cn('size-4', a.is_favorited && 'fill-red-500 stroke-red-500')} />
						</Button>
					</form>
					{#if canEdit}
						<EditAudioSheet
							audio={a}
							languages={data.languages}
							allArtists={data.allArtists}
							allTags={data.allTags}
							allCircles={data.allCircles}
							allVoiceActors={data.allVoiceActors}
						/>
					{/if}
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
							class="p-0.5 text-muted-foreground transition-colors hover:text-primary disabled:opacity-50"
							aria-pressed={a.rating === n}
							aria-label={a.rating === n
								? `Clear rating (${n} star${n === 1 ? '' : 's'})`
								: `Rate ${n} star${n === 1 ? '' : 's'}`}
							onmouseenter={() => (hoveredStar = n)}
						>
							<Star
								class={cn(
									'size-5',
									(hoveredStar ?? a.rating ?? 0) >= n ? 'fill-primary stroke-primary' : 'fill-none'
								)}
							/>
						</button>
					</form>
				{/each}
			</div>

			<dl class="mt-4 grid grid-cols-1 gap-x-6 gap-y-2 text-sm sm:grid-cols-[7rem_1fr]">
				<dt class="text-muted-foreground">Duration</dt>
				<dd>{formatDuration(a.duration_seconds)}</dd>

				{#if formatDate(a.release_date)}
					<dt class="text-muted-foreground">Release date</dt>
					<dd>{formatDate(a.release_date)}</dd>
				{/if}

				{#if languageName}
					<dt class="text-muted-foreground">Language</dt>
					<dd>{languageName}</dd>
				{/if}

				{#if a.circle}
					<dt class="text-muted-foreground">Circle</dt>
					<dd>
						<a href={resolve(`/circle/${encodeURIComponent(a.circle)}`)} class="hover:underline">
							{a.circle}
						</a>
					</dd>
				{/if}

				{#if a.artists.length}
					<dt class="text-muted-foreground">Artists</dt>
					<dd class="flex flex-wrap gap-x-2">
						{#each a.artists as artist (artist)}
							{@const id = artistId(artist)}
							{#if id !== undefined}
								<a href={resolve(`/artist/${id}`)} class="hover:underline">{artist}</a>
							{:else}
								<span>{artist}</span>
							{/if}
						{/each}
					</dd>
				{/if}

				{#if a.tags.length}
					<dt class="text-muted-foreground">Tags</dt>
					<dd class="flex flex-wrap gap-1.5">
						{#each a.tags as tag (tag)}
							<Badge href={resolve(`/tag/${encodeURIComponent(tag)}`)} variant="secondary">
								{tag}
							</Badge>
						{/each}
					</dd>
				{/if}

				{#if a.voice_actors.length}
					<dt class="text-muted-foreground">Voice actors</dt>
					<dd class="flex flex-wrap gap-x-2">
						{#each a.voice_actors as voiceActor (voiceActor)}
							<a
								href={resolve(`/voice-actor/${encodeURIComponent(voiceActor)}`)}
								class="hover:underline"
							>
								{voiceActor}
							</a>
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

	{#snippet player()}
		<AudioWorkTracks audio={a} {defaultFolder} />
	{/snippet}

	{#if a.images.length}
		<Tabs.Root bind:value={section} class="gap-6">
			<Tabs.List class="self-center">
				<Tabs.Trigger value="tracks">Tracks</Tabs.Trigger>
				<Tabs.Trigger value="gallery">Gallery</Tabs.Trigger>
			</Tabs.List>

			<Tabs.Content value="tracks">
				{@render player()}
			</Tabs.Content>

			<Tabs.Content value="gallery">
				<AudioImageGallery audioId={a.id} updatedAt={a.updated_at} images={a.images} />
			</Tabs.Content>
		</Tabs.Root>
	{:else}
		{@render player()}
	{/if}
</div>
