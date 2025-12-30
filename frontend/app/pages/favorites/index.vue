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
  title: "Favorites",
});

const { sortBy, sortDir, filters } = storeToRefs(useFavoriteFiltersStore());

const token = await useAuth().getToken();
const currentPage = ref(0);
const query = useRoute().query;
const page = parseInt(query.page as string, 10);
if (isNaN(page)) {
  currentPage.value = 1;
} else {
  currentPage.value = page;
}

const pageSize = useRuntimeConfig().public.pageSize;

const { data: archives } = await useAsyncData(
  "favorites",
  () =>
    $fetch("/api/user/favorites", {
      method: "POST",
      onRequest({ options }) {
        options.headers.set("Authorization", `Bearer ${token}`);
      },
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
    }),
  { watch: [sortBy, sortDir, currentPage] },
);

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: "smooth" });
};

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
      class="py-4 grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6"
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
</template>
