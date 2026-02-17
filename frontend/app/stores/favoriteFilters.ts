import { defineStore } from "pinia";

export const useFavoriteFiltersStore = defineStore("favoriteFilters", () => {
  const route = useRoute();
  const { SortOptionsStore } = useSortOptions();

  const sortBy =
    route.query.sortby != undefined
      ? ref(route.query.sortby)
      : ref("favorited_at");
  const sortDir =
    route.query.sortdir != undefined ? ref(route.query.sortdir) : ref("desc");
  const sortLabel =
    route.query.sortby != undefined
      ? ref(SortOptionsStore[<string>sortBy.value])
      : ref("Favorited At");

  const filters = ref({
    tags: <string>route.query.tags || undefined,
    artists: <string>route.query.artists || undefined,
    characters: <string>route.query.characters || undefined,
    parodies: <string>route.query.parodies || undefined,
    languages: <string>route.query.languages || undefined,
    categories: <string>route.query.categories || undefined,
  });

  const filtersSplit = ref({
    tags: filters.value.tags?.split(",") || [],
    artists: filters.value.artists?.split(",") || [],
    characters: filters.value.characters?.split(",") || [],
    parodies: filters.value.parodies?.split(",") || [],
    languages: filters.value.languages?.split(",") || [],
    categories: filters.value.categories?.split(",") || [],
  });

  const page = ref(
    isNaN(parseInt(<string>route.query.page, 10))
      ? 1
      : parseInt(<string>route.query.page, 10),
  );

  const pageSize = ref(useRuntimeConfig().public.pageSize);

  return {
    sortBy,
    sortDir,
    sortLabel,
    filters,
    filtersSplit,
    page,
    pageSize,
  };
});

