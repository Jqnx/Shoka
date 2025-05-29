<script setup lang="ts">
  //const { data: archives } = await useFetch("/api/a", {
  //  query: { page: 1, size: 10 },
  //  key: "archives",
  //});
</script>

<template>
  <div class="flex flex-col gap-4 p-4 pt-0">
    <div class="flex flex-col gap-2">
      <CarouselTitle
        title="Recent Doujinshi"
        :to="`/a`" />
      <ArchiveCarousel />
    </div>
  </div>
</template>
