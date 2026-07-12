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

	artists: string[];
	tags: string[];
	parodies: string[];
	circles: string[];
	characters: string[];

	progress: Progress | null;
	thumbs_ready: boolean;
};

export type ArchiveListResponse = {
	items: Archive[];
	total: number;
	page: number;
	limit: number;
};

export type SortOption = 'latest' | 'title' | 'release';
export type Category = 'Doujinshi' | 'Manga' | 'Artist CG' | 'Game CG' | 'Other';
export type Language = 'English' | 'Japanese';

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
