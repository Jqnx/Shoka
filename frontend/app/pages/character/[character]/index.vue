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

const { character } = useRoute().params;
useHead({
  title: `Character: ${character}`,
});

const token = await useAuth().getToken();
const router = useRouter();
const route = useRoute();
const { SortOptions: sortList, SortOptionsStore } = useSortOptions();

const sortBy =
  route.query.sortby != undefined ? ref(route.query.sortby) : ref("title");
const sortDir =
  route.query.sortdir != undefined ? ref(route.query.sortdir) : ref("asc");
const sortLabel =
  route.query.sortby != undefined
    ? ref(SortOptionsStore[sortBy.value as string])
    : ref("Title");
const page = parseInt(route.query.page as string, 10);
const currentPage = ref(isNaN(page) ? 1 : page);
const pageSize = ref(useRuntimeConfig().public.pageSize);

const siblingCount = computed(() => {
  return useDevice().isMobile ? 0 : 1;
});

const { data: archives } = await useFetch(`/api/character/${character}`, {
  onRequest({ options }) {
    options.headers.set("Authorization", `Bearer ${token}`);
  },
  query: {
    page: currentPage,
    size: pageSize,
    sortby: sortBy,
    sortdir: sortDir,
  },
  key: `archives-${character}`,
});

watch([currentPage, sortBy, sortDir], () => {
  router.push({
    query: {
      page: currentPage.value,
      size: pageSize.value,
      sortby: sortBy.value,
      sortdir: sortDir.value,
    },
  });
});

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: "smooth" });
};
</script>

<template>
  <div class="flex flex-1 flex-col gap-2">
    <ListOptions
      v-model:sort-by="sortBy"
      v-model:sort-dir="sortDir"
      v-model:sort-label="sortLabel"
      :sort-list="sortList"
    />
    <div
      class="px-2 py-2 grid gap-2 xl:gap-4 grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
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
            v-if="item.type == 'page'"
            :key="index"
            :value="item.value"
            :is-active="item.value == currentPage"
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
</template>
