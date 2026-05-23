<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import { BookOpen, ChevronLeft, Calendar, FileText, Tag, Users, Tv, Sword } from '@lucide/svelte';

	let { data } = $props();
	const a = $derived(data.archive);

	function formatDate(iso: string | null) {
		if (!iso) return null;
		return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' });
	}
</script>

<div class="mx-auto max-w-screen-xl px-4 py-8 sm:px-6">
	<!-- Back -->
	<a
		href="/a"
		class="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
	>
		<ChevronLeft class="size-4" />
		Archives
	</a>

	<div class="flex flex-col gap-8 md:flex-row">
		<!-- Cover -->
		<div class="w-full shrink-0 md:w-56 lg:w-64">
			<div class="aspect-[2/3] w-full overflow-hidden rounded-lg border border-border bg-muted">
				<img
					src="http://localhost:8080/api/archives/{a.id}/cover"
					alt={a.title}
					class="size-full object-cover"
				/>
			</div>

			<Button class="mt-4 w-full gap-2" size="lg">
				<BookOpen class="size-4" />
				Read
			</Button>

			{#if a.progress}
				<div class="mt-3 rounded-md border border-border bg-card p-3 text-sm">
					<p class="font-medium">Reading progress</p>
					<p class="mt-1 text-muted-foreground">
						Page {a.progress.current_page} of {a.page_count}
					</p>
					<div class="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-muted">
						<div
							class="h-full rounded-full bg-primary transition-all"
							style="width: {Math.round((a.progress.current_page / a.page_count) * 100)}%"
						></div>
					</div>
					{#if a.progress.completed}
						<p class="mt-1.5 text-xs text-muted-foreground">Completed</p>
					{/if}
				</div>
			{/if}
		</div>

		<!-- Metadata -->
		<div class="flex-1 min-w-0">
			<h1 class="text-2xl font-bold leading-tight">{a.title}</h1>

			{#if a.artists?.length}
				<p class="mt-1 text-base text-muted-foreground">{a.artists.join(', ')}</p>
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
				<p class="mt-4 text-sm text-muted-foreground leading-relaxed">{a.summary}</p>
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
									<a href="/parody/{encodeURIComponent(parody)}" class="hover:underline">{parody}{i < a.parodies.length - 1 ? ',' : ''}</a>
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
									<a href="/character/{encodeURIComponent(character)}" class="hover:underline">{character}{i < a.characters.length - 1 ? ',' : ''}</a>
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
							<a href="/tag/{encodeURIComponent(tag)}">
								<Badge variant="secondary" class="cursor-pointer text-xs hover:bg-secondary/80">{tag}</Badge>
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
</div>
