import type { Artist, Character, Parody, Tag, URL } from "./metadata";

export interface Archive {
  id: string;
  title: string;
  summary: string;
  language: string;
  category: string;
  pageCount: number;
  fileHash: string;
  type: string;
  createdAt: string;
  updatedAt: string;
  releaseDate: string;
  pagesOnDisk: number;
  tags: Tag[];
  artists: Artist[];
  parodies: Parody[];
  characters: Character[];
  url: URL[];
  status: string;
  progress: number;
  lastRead: string;
  isFavorite: boolean;
}
