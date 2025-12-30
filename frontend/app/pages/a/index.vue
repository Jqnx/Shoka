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

useHead({
  title: "Archives",
});

const currentPage = ref(0);
const query = useRoute().query;
const page = parseInt(query.page as string, 10);
if (isNaN(page)) {
  currentPage.value = 1;
} else {
  currentPage.value = page;
}

const pageSize = useRuntimeConfig().public.pageSize;

const { sortBy, sortDir, filters } = storeToRefs(useFiltersStore());
const token = await useAuth().getToken();

const { data: archives } = await useAsyncData(
  "archives",
  () =>
    $fetch("/api/a/filter", {
      method: "POST",
      query: {
        page: currentPage.value,
        size: pageSize,
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
      onRequest({ options }) {
        options.headers.set("Authorization", `Bearer ${token}`);
      },
      onResponse({ response }) {
        if (response._data.total != 0) {
          generateCover(response._data.archives);
        }
      },
    }),
  { watch: [sortBy, sortDir, currentPage] },
);

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: "smooth" });
};
</script>

<template>
  <div v-if="archives" class="flex flex-1 flex-col gap-4">
    <ListOptions />
    <div
      class="py-4 grid gap-8 grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
    >
      <div v-for="archive in archives.archives" :key="archive.id">
        <GalleryItem
          :id="archive.id"
          :title="archive.title"
          :progress="archive.page"
          :page-count="archive.page_count"
        />
      </div>
    </div>
    <Pagination
      v-model:page="currentPage"
      :show-edges="true"
      :sibling-count="2"
      :items-per-page="pageSize"
      :total="archives.total"
      :default-page="1"
      @update:page="navigateTo(`/a?page=${currentPage}`)"
    >
      <PaginationContent v-slot="{ items }">
        <PaginationFirst @click="scrollToTop" />
        <PaginationPrevious @click="scrollToTop" />

        <template v-for="(item, index) in items" :key="index">
          <PaginationItem
            v-if="item.type === 'page'"
            :key="index"
            :value="item.value"
            :is-active="item.value === currentPage"
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
