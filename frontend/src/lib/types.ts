export type Progress = {
	current_page: number;
	last_read: string;
	completed: boolean;
};

export type Archive = {
	id: string;
	title: string;
	summary: string | null;
	language: string | null;
	category: string | null;
	release_date: string | null;
	page_count: number;
	file_path: string;
	created_at: string;
	updated_at: string;

	// The backend builds these by appending onto a zero-value Go slice, so an
	// archive with zero of a given relation serializes it as JSON `null`, not
	// `[]` - genuinely nullable, not just optional. Every consumer needs to
	// account for that (see EditArchiveSheet's `?? []` when seeding editable
	// state - the bug that happens when you don't).
	artists: string[] | null;
	tags: string[] | null;
	parodies: string[] | null;
	circles: string[] | null;
	characters: string[] | null;

	progress: Progress | null;
	thumbs_ready: boolean;
	is_favorited: boolean;
	rating: number | null;
};

export type ArchiveListResponse = {
	items: Archive[];
	total: number;
	page: number;
	limit: number;
};

export type SortOption = 'latest' | 'title' | 'release';

// The shape returned by POST /api/archives/{id}/metadata and
// POST /api/archives/{id}/metadata/{source} - these only ever run the
// pipeline/source and hand back a *preview* of what it found. Nothing is
// persisted until the archive's own PATCH endpoint is called (i.e. the
// normal Save button), so a field being null here just means "this
// source/pipeline didn't provide a value for it", not "clear it".
// page_count is intentionally omitted - it isn't editable (see
// UpdateArchiveRequest's comment on the backend), so there's nothing to
// diff/apply it against.
export type FetchedMetadata = {
	title: string | null;
	summary: string | null;
	language: string | null;
	category: string | null;
	release_date: string | null;
	artists: string[] | null;
	tags: string[] | null;
	parodies: string[] | null;
	circles: string[] | null;
	characters: string[] | null;
};

// GET /api/metadata/sources?library_id=X - the enabled state is per-library
// (each library has its own library_source rows), so this always needs a
// library_id, unlike the archive-scoped fetch endpoints above.
export type MetadataSourceInfo = {
	name: string;
	priority: number;
	enabled: boolean;
};

// GET /api/archives/{id}/metadata/{source}/search - a lightweight preview
// card for manual selection, NOT the full metadata.Result. The backend only
// populates the rest of a result's fields (summary/category/release_date/
// artists/parodies/circles/characters) when it actually applies a pick via
// POST .../metadata/{source}/{source_id} - which persists immediately, with
// no preview mode. MetadataFetchDialog deliberately never calls that
// endpoint, so a diff built from a picked search result can only ever cover
// what's here: title/language/tags.
export type MetadataSearchResult = {
	id: string;
	title: string;
	cover_url: string;
	language: string;
	tags: string[] | null;
	page_count: number;
	updated_at: string | null;
};

// Both endpoints reflect only what's actually in use across libraries right
// now (nothing is seeded ahead of time) - see GetCategories/GetLanguages.
export type ArchiveLanguage = {
	code: string;
	name: string;
};

export type Tag = {
	id: number;
	name: string;
	description: string | null;
	count: number;
};

export type TagListResponse = {
	items: Tag[];
	total: number;
	page: number;
	limit: number;
};

export type Character = {
	id: number;
	name: string;
	count: number;
};

export type CharacterListResponse = {
	items: Character[];
	total: number;
	page: number;
	limit: number;
};

export type Parody = {
	id: number;
	name: string;
	count: number;
};

export type ParodyListResponse = {
	items: Parody[];
	total: number;
	page: number;
	limit: number;
};

export type Artist = {
	id: number;
	name: string;
	count: number;
};

export type ArtistListResponse = {
	items: Artist[];
	total: number;
	page: number;
	limit: number;
};

export type Library = {
	id: string;
	name: string;
	path: string;
	type: string;
	enabled: boolean;
	created_at: string;
	updated_at: string;
};

export type LibrarySource = {
	source: string;
	enabled: boolean;
	cookies?: string;
	api_key?: string;
	magazine_blocklist: string[];
	misc_blocklist: string[];
};

export type LibraryType = {
	type: string;
	supported: boolean;
};

export type ArchiveSortOption = {
	value: string;
	display_name: string;
};

// GET/PATCH /api/reader-settings - per-user reader preferences, stored
// backend-side (not localStorage) so they sync across devices/browsers, same
// as reading progress. GET always returns a full, usable object even if the
// user has never customized anything (the backend has its own defaults and
// only creates a row on the first PATCH) - there's no "unset" state to
// handle on this side.
//
// view_mode and reading_direction are orthogonal: view_mode picks paged vs.
// scroll, reading_direction only decides which physical side is
// "next" while paged (irrelevant in scroll mode, which always scrolls
// top-to-bottom).
export type ViewMode = 'paged' | 'scroll';
export type ReadingDirection = 'ltr' | 'rtl';
export type PageLayout = 'single' | 'double';
export type FitMode = 'width' | 'height' | 'original';
export type ReaderBackground = 'black' | 'white' | 'gray';

export type ReaderSettings = {
	view_mode: ViewMode;
	reading_direction: ReadingDirection;
	page_layout: PageLayout;
	fit_mode: FitMode;
	background: ReaderBackground;
};
