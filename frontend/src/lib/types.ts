export type Archive = {
	id: number;
	title: string;
	artist: string;
	category: string;
	language: string;
	tags: string[];
	pages: number;
	coverHue: number;
};

export type SortOption = 'latest' | 'title' | 'release';
export type Category = 'Doujinshi' | 'Manga' | 'Artist CG' | 'Game CG' | 'Other';
export type Language = 'English' | 'Japanese';
