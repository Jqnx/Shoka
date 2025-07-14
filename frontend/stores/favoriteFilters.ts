import { defineStore } from 'pinia'


export const useFavoriteFiltersStore = defineStore('favoriteFilters', {
  state: () => {
    const tags : string[] = [];
    const artists : string[] = [];
    const characters : string[] = [];
    const parodies : string[] = [];
    const languages : string[] = [];
    const categories : string[] = [];
    return {
      sortBy: "favorited_at",
      sortDir: "desc",
      sortLabel: "Favorited At",
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