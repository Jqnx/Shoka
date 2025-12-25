<script lang="ts" setup>
definePageMeta({
  layout: "reader",
});

const token = await useAuth().getToken();
const data = useNuxtData("archive_reader");

const params = computed(() => {
  return useRoute().params;
});
const pageInt = computed(() => {
  return Number(params.value.page);
});
const clamp = (num: number, min: number, max: number) => {
  return Math.min(Math.max(num, min), max);
};

useHead({
  title: `${data.data.value.title} - Page ${pageInt.value}`,
});

const { preload, fit } = storeToRefs(useReaderSettingsStore());

onBeforeRouteUpdate(() => {
  $fetch(`/api/a/${params.value.id}/${params.value.page}`, {
    method: "post",
    onRequest({ options }) {
      options.headers.set("Authorization", `Bearer ${token}`);
    },
  });
});

onBeforeRouteLeave(() => {
  $fetch(`/api/a/${params.value.id}/${params.value.page}`, {
    method: "post",
    onRequest({ options }) {
      options.headers.set("Authorization", `${token}`);
    },
  });
});
</script>

<template>
  <div class="flex flex-col items-center justify-center min-h-screen">
    <div v-for="(_, index) in preload" :key="index">
      <NuxtImg
        :src="`/archive/${params.id}/${clamp(
          pageInt + index,
          pageInt,
          data.data.value.page_count,
        )}`"
        preload
        hidden
      />
    </div>
    <NuxtImg
      draggable="false"
      :class="fit"
      :src="`/archive/${params.id}/${params.page}`"
    />
  </div>
</template>
