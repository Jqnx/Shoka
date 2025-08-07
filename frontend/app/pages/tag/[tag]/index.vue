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

  const { currentPage } = storeToRefs(usePageStore());
  const { tag } = useRoute().params;
  const { token } = useAuth();

  useHead({
    title: `Tag: ${tag}`,
  });

  const pageSize = ref(30);
  const { data: archives } = await useFetch(`/api/tag/${tag}`, {
    onRequest({ options }) {
      options.headers.set("Authorization", `${token.value}`);
    },
    query: { page: currentPage, size: pageSize },
    key: "archives",
  });

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
    <div
      class="py-4 grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6">
      <div v-for="archive in archives.archives" :key="archive.id">
        <GalleryItem
          :id="archive.archive_id"
          :title="archive.title"
          :progress="archive.page"
          :page_count="archive.page_count" />
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
        <PaginationFirst @click="scrollToTop" />
        <PaginationPrevious @click="scrollToTop" />

        <template v-for="(item, index) in items" :key="index">
          <PaginationItem
            v-if="item.type == 'page'"
            :key="index"
            :value="item.value"
            :is-active="item.value == currentPage"
            @click="scrollToTop">
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
