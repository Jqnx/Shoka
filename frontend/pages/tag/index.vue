<script setup lang="ts">
  const { data: tags } = await useFetch("/api/tag", {
    key: "tags",
  });
</script>

<template>
  <div>
    <h1
      class="flex justify-center py-4 text-text font-semibold text-2xl tracking-tight">
      Tags
    </h1>
    <div class="grid grid-cols-6 h-screen gap-1.5 px-8 py-4">
      <Button
        v-for="(tag, index) in tags"
        :key="index"
        class="w-full">
        <div class="flex flex-1 justify-between">
          <p>{{ tag.tag }}</p>
          <p><!--TODO: Tag count--></p>
        </div>
      </Button>
    </div>
  </div>
</template>
