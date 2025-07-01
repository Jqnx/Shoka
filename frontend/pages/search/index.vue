<script setup lang="ts">
  import {
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationItem,
    PaginationNext,
    PaginationPrevious,
  } from "@/components/ui/pagination";

  const { currentPage, pageSize } = storeToRefs(usePageStore());
  const { searchQuery } = storeToRefs(useSearchStore());
  const route = useRoute();
  //const { sortBy, sortDir, filters } = storeToRefs(useFiltersStore());

  const { data: archives } = await useFetch("/api/search", {
    query: {
      q: computed(() => {
        return route.query.q;
      }),
      page: currentPage.value,
      size: pageSize.value,
    },
    key: "searchPageResults",
  });

  const router = useRouter();
  router.beforeResolve((_) => {
    currentPage.value = 1;
  });

  router.beforeEach(() => {
    searchQuery.value = "";
  });

  onMounted(() => {
    searchQuery.value = route.query.q as string;
  });
</script>

<template>
  <div class="flex flex-1 flex-col gap-4 p-4 pt-0">
    <!--TODO: Re-enable when shuffle functionality is done, maybe only use shuffle -->
    <!--
    <ListOptions class="px-16" />
    -->
    <h1
      class="scroll-m-20 text-2xl font-semibold tracking-tight text-center pt-3">
      {{ archives.total }} results found.
    </h1>
    <div
      class="px-16 py-4 grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6">
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
