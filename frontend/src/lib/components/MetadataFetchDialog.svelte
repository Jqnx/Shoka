<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Sparkles, Search, ArrowLeft, ImageOff } from '@lucide/svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import { errorMessage } from '$lib/api';
	import { METADATA_SOURCE_INFO, sourceInfo } from '$lib/metadata-sources';
	import type {
		Archive,
		ArchiveLanguage,
		FetchedMetadata,
		MetadataSearchResult,
		MetadataSourceInfo
	} from '$lib/types';

	type CurrentMetadata = {
		title: string;
		summary: string;
		category: string;
		language: string;
		releaseDate: string;
		artists: string[];
		tags: string[];
		parodies: string[];
		circles: string[];
		characters: string[];
		urls: string[];
	};

	type DiffField = {
		key: keyof CurrentMetadata;
		label: string;
		oldDisplay: string;
		newDisplay: string;
		newValue: string | string[];
	};

	let {
		archiveId,
		libraryId,
		current,
		languages,
		onApply
	}: {
		archiveId: string;
		libraryId: string;
		current: CurrentMetadata;
		languages: ArchiveLanguage[];
		onApply: (fields: Partial<CurrentMetadata>) => void;
	} = $props();

	const SOURCE_KEYS = Object.keys(METADATA_SOURCE_INFO);
	const ALL_SOURCES = 'all';

	let open = $state(false);
	// 'search' shows a picker of results for the user to compare and choose
	// from (when the source supports search); 'auto' skips that picker and
	// goes straight to the top result's diff - see runFetch().
	let mode = $state<'search' | 'auto'>('search');
	let sourceValue = $state(ALL_SOURCES);
	// Seeded from the archive's title (the backend's own default when no
	// override is given), then freely editable - lets a source's search be
	// retried with a better query when the title alone doesn't find it.
	// Only meaningful for a specific source; "All sources" has no search of
	// its own to query.
	// svelte-ignore state_referenced_locally
	let searchQuery = $state(current.title);
	let loading = $state(false);
	let picking = $state(false);
	let error = $state<string | null>(null);
	// Three steps, layered so "back" can reveal the previous one rather than
	// discarding it: pick a source -> (if searchable) pick a result -> diff.
	// searchResults stays populated while a diff is shown from a picked
	// result, so going back from the diff returns to the same list.
	let searchResults = $state<MetadataSearchResult[] | null>(null);
	let fetched = $state<FetchedMetadata | null>(null);
	const selected = new SvelteSet<string>();

	// Per-library enabled state for each source, so a disabled one can be
	// shown as such in the picker instead of only failing after the user
	// tries to fetch from it. Loaded lazily (once) the first time the dialog
	// opens - fails open (nothing marked disabled) if libraryId is unknown
	// or the request fails, since that's strictly better than blocking the
	// whole picker over a best-effort enhancement.
	let disabledSources = $state<Set<string>>(new Set());
	let sourcesLoaded = false;

	async function loadSourceAvailability() {
		if (sourcesLoaded || !libraryId) return;
		sourcesLoaded = true;
		try {
			const res = await fetch(`/api/metadata/sources?library_id=${libraryId}`);
			if (!res.ok) return;
			const sources: MetadataSourceInfo[] = await res.json();
			disabledSources = new Set(sources.filter((s) => !s.enabled).map((s) => s.name));
		} catch {
			// best-effort - leave everything selectable
		}
	}

	const sourceLabel = $derived(
		sourceValue === ALL_SOURCES ? 'All sources (pipeline)' : sourceInfo(sourceValue).label
	);

	function languageLabel(code: string) {
		return languages.find((l) => l.code === code)?.name ?? code;
	}

	// SearchResult.Language is expected to be a raw ISO code, same as every
	// Result.Language elsewhere - this just guards against a source sending
	// a display name instead (matching against both `code` and `name`),
	// falling back to the raw value if it's not recognized as either -
	// same graceful-degradation as languageLabel above.
	function normalizeLanguage(value: string) {
		if (!value) return '';
		const match = languages.find((l) => l.code === value || l.name === value);
		return match?.code ?? value;
	}

	function listsEqual(a: string[], b: string[]) {
		if (a.length !== b.length) return false;
		const sa = [...a].sort();
		const sb = [...b].sort();
		return sa.every((v, i) => v === sb[i]);
	}

	function computeDiff(cur: CurrentMetadata, result: FetchedMetadata): DiffField[] {
		const fields: DiffField[] = [];

		function addText(
			key: keyof CurrentMetadata,
			label: string,
			oldVal: string,
			newVal: string | null,
			format: (v: string) => string = (v) => v
		) {
			if (newVal === null || newVal === undefined || newVal === oldVal) return;
			fields.push({
				key,
				label,
				oldDisplay: oldVal ? format(oldVal) : '—',
				newDisplay: newVal ? format(newVal) : '—',
				newValue: newVal
			});
		}

		function addList(
			key: keyof CurrentMetadata,
			label: string,
			oldVal: string[],
			newVal: string[] | null
		) {
			if (newVal === null || newVal === undefined || listsEqual(oldVal, newVal)) return;
			fields.push({
				key,
				label,
				oldDisplay: oldVal.length ? oldVal.join(', ') : '—',
				newDisplay: newVal.length ? newVal.join(', ') : '—',
				newValue: newVal
			});
		}

		addText('title', 'Title', cur.title, result.title);
		addText('summary', 'Summary', cur.summary, result.summary);
		addText('category', 'Category', cur.category, result.category);
		addText('language', 'Language', cur.language, result.language, languageLabel);
		addText(
			'releaseDate',
			'Release date',
			cur.releaseDate,
			result.release_date?.slice(0, 10) ?? null
		);
		addList('artists', 'Artists', cur.artists, result.artists);
		addList('tags', 'Tags', cur.tags, result.tags);
		addList('parodies', 'Parodies', cur.parodies, result.parodies);
		addList('circles', 'Circles', cur.circles, result.circles);
		addList('characters', 'Characters', cur.characters, result.characters);
		// URLs are additive by nature — an nhentai gallery link doesn't
		// invalidate a hand-added e-hentai one — so the "new value" is the
		// union of current and fetched, not a replacement. Every other field
		// keeps replace behaviour.
		addList(
			'urls',
			'Source links',
			cur.urls,
			result.urls ? [...new Set([...cur.urls, ...result.urls])] : null
		);

		return fields;
	}

	const diffFields = $derived(fetched ? computeDiff(current, fetched) : []);

	async function fetchDirect(url: string) {
		try {
			const res = await fetch(url, { method: 'POST' });
			if (!res.ok) {
				error = await errorMessage(res, 'Failed to fetch metadata.');
				return;
			}
			const result: FetchedMetadata = await res.json();
			fetched = result;
			for (const field of computeDiff(current, result)) selected.add(field.key);
		} catch {
			error = 'Failed to fetch metadata.';
		}
	}

	async function runFetch() {
		loading = true;
		error = null;
		searchResults = null;
		fetched = null;
		selected.clear();

		// "All sources" runs the whole priority-ordered pipeline at once -
		// there's no per-source search equivalent for that, so it always
		// goes straight to a direct fetch.
		if (sourceValue !== ALL_SOURCES) {
			try {
				const q = searchQuery.trim();
				const searchUrl = `/api/archives/${archiveId}/metadata/${sourceValue}/search${q ? `?q=${encodeURIComponent(q)}` : ''}`;
				const res = await fetch(searchUrl);
				if (res.ok) {
					const results: MetadataSearchResult[] = await res.json();

					// Auto mode: skip the picker entirely, take the top result
					// (the source's own best-match ranking) straight to its diff.
					// loading stays true through pickResult() too, so the Fetch
					// button stays disabled/labeled for the whole round trip
					// instead of flashing enabled again mid-fetch.
					if (mode === 'auto') {
						if (results.length === 0) {
							loading = false;
							error = 'No matching results found.';
							return;
						}
						await pickResult(results[0]);
						loading = false;
						return;
					}

					searchResults = results;
					loading = false;
					return;
				}
				// 400 = this source doesn't implement search (comicinfo/filename/
				// e-hentai today) - fall through to the direct fetch below rather
				// than surfacing that as an error. Anything else (403 disabled,
				// 500, ...) is a real error.
				if (res.status !== 400) {
					error = await errorMessage(res, 'Failed to search for metadata.');
					loading = false;
					return;
				}
			} catch {
				error = 'Failed to search for metadata.';
				loading = false;
				return;
			}
		}

		const url =
			sourceValue === ALL_SOURCES
				? `/api/archives/${archiveId}/metadata`
				: `/api/archives/${archiveId}/metadata/${sourceValue}`;
		await fetchDirect(url);
		loading = false;
	}

	// A search result (MetadataSearchResult) is only ever a lightweight
	// preview card - title/language/tags, nothing else. To diff on every
	// field, fetch the picked result's full details via
	// POST .../metadata/{source}/{source_id}, which now only *previews*
	// what the archive would look like with it applied (the existing DB row
	// with the result's fields overlaid on top) rather than saving anything.
	async function pickResult(result: MetadataSearchResult) {
		picking = true;
		error = null;
		try {
			const res = await fetch(
				`/api/archives/${archiveId}/metadata/${sourceValue}/${encodeURIComponent(result.id)}`,
				{ method: 'POST' }
			);
			if (!res.ok) {
				error = await errorMessage(res, 'Failed to fetch details for this result.');
				return;
			}
			const preview: Archive = await res.json();
			const asMetadata: FetchedMetadata = {
				title: preview.title,
				summary: preview.summary,
				// Same conversion as the archive GET endpoint - a display name,
				// not the raw code the rest of this dialog works in.
				language: preview.language ? normalizeLanguage(preview.language) : null,
				category: preview.category,
				release_date: preview.release_date,
				artists: preview.artists,
				tags: preview.tags,
				parodies: preview.parodies,
				circles: preview.circles,
				characters: preview.characters,
				urls: preview.urls?.map((u) => u.url) ?? null
			};
			fetched = asMetadata;
			selected.clear();
			for (const field of computeDiff(current, asMetadata)) selected.add(field.key);
		} catch {
			error = 'Failed to fetch details for this result.';
		} finally {
			picking = false;
		}
	}

	function toggle(key: string) {
		if (selected.has(key)) selected.delete(key);
		else selected.add(key);
	}

	function apply() {
		const partial: Partial<CurrentMetadata> = {};
		for (const field of diffFields) {
			if (!selected.has(field.key)) continue;
			(partial as Record<string, string | string[]>)[field.key] = field.newValue;
		}
		onApply(partial);
		open = false;
		searchResults = null;
		fetched = null;
		error = null;
	}
</script>

<div class="flex gap-2">
	<Button
		type="button"
		variant="outline"
		size="sm"
		class="flex-1 gap-1.5"
		onclick={() => {
			mode = 'search';
			open = true;
			loadSourceAvailability();
		}}
	>
		<Search class="size-4" />
		Search metadata
	</Button>
	<Button
		type="button"
		variant="outline"
		size="sm"
		class="flex-1 gap-1.5"
		onclick={() => {
			mode = 'auto';
			open = true;
			loadSourceAvailability();
		}}
	>
		<Sparkles class="size-4" />
		Fetch metadata
	</Button>
</div>

<Dialog.Root
	bind:open
	onOpenChange={(isOpen) => {
		if (!isOpen) {
			searchResults = null;
			fetched = null;
			error = null;
		}
	}}
>
	<Dialog.Content class="sm:max-w-lg">
		<Dialog.Header>
			<Dialog.Title>{mode === 'auto' ? 'Fetch metadata' : 'Search metadata'}</Dialog.Title>
			<Dialog.Description>
				{#if fetched}
					Choose which fields to apply. Nothing is saved until you hit Save changes below.
				{:else if searchResults}
					Pick a result to compare against the current form.
				{:else if mode === 'auto'}
					Automatically picks the best match from the selected source and shows you the differences.
					Nothing is saved until you hit Save changes below.
				{:else}
					Pull metadata from a source and choose which fields to apply. Nothing is saved until you
					hit Save changes below.
				{/if}
			</Dialog.Description>
		</Dialog.Header>

		{#if fetched}
			<Button
				type="button"
				variant="ghost"
				size="sm"
				class="w-fit gap-1.5"
				onclick={() => {
					fetched = null;
					error = null;
				}}
			>
				<ArrowLeft class="size-3.5" />
				Back
			</Button>

			{#if diffFields.length === 0}
				<p class="text-sm text-muted-foreground">No differences from the current form values.</p>
			{:else}
				<div class="flex max-h-80 flex-col gap-2 overflow-y-auto">
					{#each diffFields as field (field.key)}
						<div class="flex items-start gap-3 rounded-md border border-border p-2.5 text-sm">
							<Switch
								checked={selected.has(field.key)}
								onCheckedChange={() => toggle(field.key)}
								class="mt-0.5"
							/>
							<div class="min-w-0 flex-1">
								<p class="font-medium">{field.label}</p>
								<p class="text-muted-foreground line-through">{field.oldDisplay}</p>
								<p class="text-foreground">{field.newDisplay}</p>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		{:else if searchResults}
			<Button
				type="button"
				variant="ghost"
				size="sm"
				class="w-fit gap-1.5"
				onclick={() => {
					searchResults = null;
					error = null;
				}}
			>
				<ArrowLeft class="size-3.5" />
				Back
			</Button>

			{#if picking}
				<p class="text-sm text-muted-foreground">Loading details…</p>
			{/if}

			{#if searchResults.length === 0}
				<p class="text-sm text-muted-foreground">No results found.</p>
			{:else}
				<div class="flex max-h-96 flex-col gap-2 overflow-y-auto">
					{#each searchResults as result (result.id)}
						<button
							type="button"
							disabled={picking}
							class="group flex items-center gap-3 rounded-md border border-border p-2 text-left text-sm hover:bg-accent disabled:pointer-events-none disabled:opacity-50"
							onclick={() => pickResult(result)}
						>
							{#if result.cover_url}
								<img
									src={result.cover_url}
									alt=""
									class="h-16 w-12 shrink-0 rounded object-cover"
								/>
							{:else}
								<div class="flex h-16 w-12 shrink-0 items-center justify-center rounded bg-muted">
									<ImageOff class="size-4 text-muted-foreground" />
								</div>
							{/if}
							<div class="min-w-0 flex-1">
								<p class="truncate font-medium group-hover:whitespace-normal">{result.title}</p>
								<div class="mt-1 flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground">
									<span>{result.page_count} pages</span>
									{#if result.language}
										<Badge variant="outline" class="px-1.5 py-0 text-[10px] font-normal">
											{normalizeLanguage(result.language)}
										</Badge>
									{/if}
								</div>
							</div>
						</button>
					{/each}
				</div>
			{/if}
		{:else}
			<div class="flex flex-col gap-2">
				<div class="flex items-center gap-2">
					<Select.Root type="single" value={sourceValue} onValueChange={(v) => (sourceValue = v)}>
						<Select.Trigger class="flex-1">
							{sourceLabel}
						</Select.Trigger>
						<Select.Content>
							<Select.Item value={ALL_SOURCES} label="All sources (pipeline)">
								All sources (pipeline)
							</Select.Item>
							{#each SOURCE_KEYS as key (key)}
								<Select.Item
									value={key}
									label={sourceInfo(key).label}
									disabled={disabledSources.has(key)}
								>
									{sourceInfo(key).label}
									{#if disabledSources.has(key)}
										<span class="text-muted-foreground">(disabled)</span>
									{/if}
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
					<Button type="button" onclick={runFetch} disabled={loading}>
						{loading ? 'Fetching…' : 'Fetch'}
					</Button>
				</div>

				{#if sourceValue !== ALL_SOURCES}
					<Input
						bind:value={searchQuery}
						placeholder="Search query"
						disabled={loading}
						onkeydown={(e) => {
							if (e.key === 'Enter') {
								e.preventDefault();
								runFetch();
							}
						}}
					/>
				{/if}
			</div>
		{/if}

		{#if error}
			<p class="text-sm text-destructive">{error}</p>
		{/if}

		<Dialog.Footer showCloseButton>
			<Button type="button" onclick={apply} disabled={!fetched || diffFields.length === 0}>
				Apply selected
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
