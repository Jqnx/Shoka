<script setup lang="ts">
  import {
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationItem,
    PaginationNext,
    PaginationPrevious,
    PaginationFirst,
    PaginationLast,
  } from "@/components/ui/pagination";

  const { currentPage, pageSize } = storeToRefs(usePageStore());
  const { sortBy, sortDir, filters } = storeToRefs(useFavoriteFiltersStore());

  const { token } = useAuth();

  const { data: archives } = await useAsyncData(
    "favorites",
    () =>
      $fetch("/api/user/favorites", {
        method: "POST",
        onRequest({ options }) {
          options.headers.set("Authorization", `${token.value}`);
        },
        query: {
          page: currentPage.value,
          size: pageSize.value,
          sortby: sortBy.value,
          sortdir: sortDir.value,
        },
        body: {
          tags: filters.value.tags,
          artists: filters.value.artists,
          characters: filters.value.characters,
          parodies: filters.value.parodies,
          languages: filters.value.languages,
          categories: filters.value.categories,
        },
      }),
    { watch: [sortBy, sortDir, currentPage] }
  );

  const router = useRouter();
  router.beforeResolve((_) => {
    currentPage.value = 1;
  });
</script>

<template>
  <div class="flex flex-1 flex-col gap-4 p-4 pt-0">
    <!--TODO: Filter options here -->
    <ListOptions />
    <div
      class="py-4 grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6">
      <div v-for="archive in archives.archives" :key="archive.id">
        <GalleryItem :id="archive.archive_id" :title="archive.title" />
      </div>
    </div>
    <Pagination
      v-model:page="currentPage"
      :show-edges="true"
      :sibling-count="2"
      :items-per-page="pageSize"
      :total="archives.total"
      :default-page="1">
      <PaginationContent v-slot="{ items }">
        <PaginationFirst />
        <PaginationPrevious />

        <template v-for="(item, index) in items" :key="index">
          <PaginationItem
            v-if="item.type == 'page'"
            :key="index"
            :value="item.value"
            :is-active="item.value == currentPage">
            {{ item.value }}
          </PaginationItem>
          <PaginationEllipsis v-else :key="item.type" :index="index" />
        </template>

        <PaginationNext />
        <PaginationLast />
      </PaginationContent>
    </Pagination>
  </div>
</template>
