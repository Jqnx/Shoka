<script setup lang="ts">
  const { data: recent } = await useFetch("/api/a/", {
    query: { page: 1, size: 14, sortby: "release_date" },
    key: "recentArchives",
  });
</script>

<template>
  <div class="flex flex-col gap-4">
    <div class="flex flex-col gap-2">
      <CarouselTitle
        class="px-4 md:px-14 lg:px-16 xl:px-18"
        title="Recent Releases"
        to="a" />
      <ArchiveCarousel
        :archives="recent"
        class="overflow-x-hidden md:px-10 xl:px-8 2xl:px-6" />
    </div>
  </div>
</template>
