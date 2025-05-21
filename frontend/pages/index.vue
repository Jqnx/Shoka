<script setup lang="ts">
const page = ref(1)
const { data: archives } = await useFetch('/api/a', {
  query: { page: page, size: 10 }
})
</script>

<template>
  <div class="flex flex-1 flex-col gap-4 p-4 pt-0">
    <div class="flex flex-1 justify-between items-center px-2">
      <!--TODO: Highlight/Slideshow/Sliding Gallery component-->
      <a href="#" class="font-semibold text-xl tracking-tight">Recent Doujinshi</a>
      <a href="#">
        <Icon name="lucide:arrow-right" size="1.25em" />
      </a>
    </div>
    <div class="grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7">
      <div v-for="archive in archives" :key="archive.archive_id">
        <GalleryItem :id="archive.archive_id" :title="archive.title" />
      </div>
    </div>
  </div>
</template>
