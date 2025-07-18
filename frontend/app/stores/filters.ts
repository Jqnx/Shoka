import { defineStore } from 'pinia'


export const useFiltersStore = defineStore('filters', {
  state: () => {
    const tags : string[] = [];
    const artists : string[] = [];
    const characters : string[] = [];
    const parodies : string[] = [];
    const languages : string[] = [];
    const categories : string[] = [];
    return {
      sortBy: "title",
      sortDir: "asc",
      sortLabel: "Title",
      filters: {
        tags: tags,
        artists: artists,
        characters: characters,
        parodies: parodies,
        languages: languages,
        categories: categories,
      }
    }
  },
  persist: {
    storage: piniaPluginPersistedstate.cookies(),
  },
})