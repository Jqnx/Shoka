<script lang="ts" setup>
const params = computed(() => {
  return useRoute().params;
});
const pageInt = computed(() => {
  return Number(params.value.page);
});

const clamp = (num: number, min: number, max: number) => {
  return Math.min(Math.max(num, min), max);
};

const data = useNuxtData("archive_reader");
const goBack = () => {
  navigateTo({
    name: "a-id-page",
    params: {
      id: params.value.id,
      page: clamp(pageInt.value - 1, 1, data.data.value.page_count),
    },
  });
};

const goNext = () => {
  if (pageInt.value === data.data.value.page_count) {
    navigateTo({
      name: "a-id",
      params: {
        id: params.value.id,
      },
    });
  } else {
    navigateTo({
      name: "a-id-page",
      params: {
        id: params.value.id,
        page: clamp(pageInt.value + 1, 1, data.data.value.page_count),
      },
    });
  }
};

const clickPage = (event: MouseEvent) => {
  const loc = event?.clientX <= window.innerWidth / 2;
  if (loc) {
    goBack();
  } else {
    goNext();
  }
};
</script>

<template>
  <div>
    <div class="cursor-pointer" @click="clickPage">
      <slot />
    </div>
    <ReaderNav />
  </div>
</template>
