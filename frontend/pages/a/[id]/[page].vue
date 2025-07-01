<script lang="ts" setup>
  definePageMeta({
    layout: "reader",
  });

  const { id, page } = useRoute().params;
</script>
<!--NOTE: class for fit to screen: max-h-screen -->
<template>
  <div class="flex justify-center">
    <NuxtImg :src="`/archive/${id}/${page}`" />
  </div>
</template>
