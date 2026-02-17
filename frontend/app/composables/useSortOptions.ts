export function useSortOptions() {
  const Title: SortOption = {
    value: "title",
    label: "Title",
  };

  const PageCount: SortOption = {
    value: "page_count",
    label: "Page Count",
  };

  const CreatedAt: SortOption = {
    value: "created_at",
    label: "Date Added",
  };

  const ReleaseDate: SortOption = {
    value: "release_date",
    label: "Release Date",
  };

  const FavoritedAt: SortOption = {
    value: "favorited_at",
    label: "Favorited At",
  };

  const SortOptions: SortOption[] = [
    { value: Title.value, label: Title.label },
    { value: PageCount.value, label: PageCount.label },
    //{ value: "favorites", label: "Favorites" },
    { value: CreatedAt.value, label: CreatedAt.label },
    { value: ReleaseDate.value, label: ReleaseDate.label },
  ];

  const SortOptionsStore: { [option: string]: string } = {
    title: "Title",
    page_count: "Page Count",
    created_at: "Date Added",
    release_date: "Release Date",
  };

  const FavoriteSortOptions: SortOption[] = [
    { value: Title.value, label: Title.label },
    { value: PageCount.value, label: PageCount.label },
    //{ value: "favorites", label: "Favorites" },
    { value: CreatedAt.value, label: CreatedAt.label },
    { value: ReleaseDate.value, label: ReleaseDate.label },
    { value: FavoritedAt.value, label: FavoritedAt.label },
  ];

  const FavoriteSortOptionsStore: { [option: string]: string } = {
    title: "Title",
    page_count: "Page Count",
    created_at: "Date Added",
    release_date: "Release Date",
    favorited_at: "Favorited At",
  };

  return {
    Title,
    PageCount,
    CreatedAt,
    ReleaseDate,
    SortOptions,
    SortOptionsStore,
    FavoriteSortOptions,
    FavoriteSortOptionsStore,
  };
}
