import { defineStore } from "pinia";

export const useFiltersStore = defineStore("filters", () => {
  const route = useRoute();
  const { SortOptionsStore } = useSortOptions();

  const sortBy =
    route.query.sortby != undefined ? ref(route.query.sortby) : ref("title");
  const sortDir =
    route.query.sortdir != undefined ? ref(route.query.sortdir) : ref("asc");
  const sortLabel =
    route.query.sortby != undefined
      ? ref(SortOptionsStore[sortBy.value as string])
      : ref("Title");

  const filters = ref({
    tags: (route.query.tags as string) || undefined,
    artists: (route.query.artists as string) || undefined,
    characters: (route.query.characters as string) || undefined,
    parodies: (route.query.parodies as string) || undefined,
    languages: (route.query.languages as string) || undefined,
    categories: (route.query.categories as string) || undefined,
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
    isNaN(parseInt(route.query.page as string, 10))
      ? 1
      : parseInt(route.query.page as string, 10),
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
