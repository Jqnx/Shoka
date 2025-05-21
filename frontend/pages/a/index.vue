<script setup lang="ts">
const page = ref(1)
const pageSize = ref(30)
const { data: archives } = await useFetch('/api/a', {
  query: { page: page, size: pageSize }
})
</script>

<template>
  <div class="flex flex-1 flex-col gap-4 p-4 pt-0">
    <!--TODO: Filter options here -->
    <div class="px-24 py-4 grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6">
      <div v-for="archive in archives" :key="archive.id">
        <GalleryItem :id="archive.archive_id" :title="archive.title" />
      </div>
    </div>
  </div>
</template>
