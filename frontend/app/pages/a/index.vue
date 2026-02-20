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

definePageMeta({
  middleware: [
    function (_, from) {
      if (from.name == "favorites") {
        navigateTo({ query: {} });
      }
    },
  ],
});

useHead({
  title: "Archives",
});

const router = useRouter();
const token = await useAuth().getToken();
const sortList = useSortOptions().SortOptions;
const { sortBy, sortDir, sortLabel, filters, page, pageSize } =
  storeToRefs(useFiltersStore());

const siblingCount = computed(() => {
  return useDevice().isMobile ? 0 : 1;
});

const { data: archives } = await useAsyncData<ArchiveList | undefined>(
  "archives",
  () =>
    $fetch("/api/a/filter", {
      method: "GET",
      query: {
        page: page.value,
        size: pageSize.value,
        sortby: sortBy.value,
        sortdir: sortDir.value,
        tags: filters.value.tags,
        artists: filters.value.artists,
        characters: filters.value.characters,
        parodies: filters.value.parodies,
        languages: filters.value.languages,
        categories: filters.value.categories,
      },
      onRequest({ options }) {
        options.headers.set("Authorization", `Bearer ${token}`);
      },
      onResponse({ response }) {
        if (response._data.total != 0) {
          generateCover(response._data.archives);
        }
      },
    }),
  { watch: [sortBy, sortDir, page, filters.value] },
);

watch([page, sortBy, sortDir, filters.value], () => {
  router.push({
    query: {
      page: page.value,
      size: pageSize.value,
      sortby: sortBy.value,
      sortdir: sortDir.value,
      tags: filters.value.tags,
      artists: filters.value.artists,
      characters: filters.value.characters,
      parodies: filters.value.parodies,
      languages: filters.value.languages,
      categories: filters.value.categories,
    },
  });
});

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: "smooth" });
};
</script>

<template>
  <div v-if="archives" class="flex flex-1 flex-col gap-2">
    <ListOptions
      v-model:sort-by="sortBy"
      v-model:sort-label="sortLabel"
      v-model:sort-dir="sortDir"
      v-model:filters="filters"
      :sort-list="sortList"
      use-filters
    />
    <div
      class="px-2 py-2 grid gap-2 xl:gap-4 grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
    >
      <div v-for="archive in archives.archives" :key="archive.id">
        <GalleryItem
          :id="archive.id"
          :title="archive.title"
          :progress="archive.progress"
          :page-count="archive.pageCount"
        />
      </div>
    </div>
    <Pagination
      v-model:page="page"
      :show-edges="true"
      :sibling-count="siblingCount"
      :items-per-page="pageSize"
      :total="archives.total"
      :default-page="1"
    >
      <PaginationContent v-slot="{ items }">
        <PaginationFirst @click="scrollToTop" />
        <PaginationPrevious @click="scrollToTop" />

        <template v-for="(item, index) in items" :key="index">
          <PaginationItem
            v-if="item.type === 'page'"
            :key="index"
            :value="item.value"
            :is-active="item.value === page"
            @click="scrollToTop"
          >
            {{ item.value }}
          </PaginationItem>
          <PaginationEllipsis v-else :key="item.type" :index="index" />
        </template>

        <PaginationNext @click="scrollToTop" />
        <PaginationLast @click="scrollToTop" />
      </PaginationContent>
    </Pagination>
  </div>
  <div v-else class="flex justify-center pt-12">
    <h1 class="text-3xl font-bold">No Archives Found</h1>
  </div>
</template>
