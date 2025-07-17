<script lang="ts" setup>
  definePageMeta({
    layout: "reader",
  });

  const { token } = useAuth();

  const params = computed(() => {
    return useRoute().params;
  });
  const pageInt = computed(() => {
    return Number(params.value.page);
  });
  const clamp = (num: number, min: number, max: number) => {
    return Math.min(Math.max(num, min), max);
  };

  const data = useNuxtData("archive_page_count");

  const { preload, fit } = storeToRefs(useReaderSettingsStore());

  onBeforeRouteUpdate(() => {
    $fetch(`/api/a/${params.value.id}/${params.value.page}`, {
      method: "post",
      onRequest({ options }) {
        options.headers.set("Authorization", `${token.value}`);
      },
    });
  });

  onBeforeRouteLeave(() => {
    $fetch(`/api/a/${params.value.id}/${params.value.page}`, {
      method: "post",
      onRequest({ options }) {
        options.headers.set("Authorization", `${token.value}`);
      },
    });
  });
</script>
<template>
  <div class="flex justify-center">
    <div v-for="(_, index) in preload" :key="index">
      <NuxtImg
        :src="`/archive/${params.id}/${clamp(
          pageInt + index,
          pageInt,
          data.data.value.page_count
        )}`"
        preload
        hidden />
    </div>
    <NuxtImg :class="fit" :src="`/archive/${params.id}/${params.page}`" />
  </div>
</template>
