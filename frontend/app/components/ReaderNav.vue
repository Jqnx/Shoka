<script lang="ts" setup>
import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Undo2,
} from "lucide-vue-next";
import { useBreakpoints } from "@vueuse/core";

const breakpoints = useBreakpoints(
  {
    mobile: 412,
  },
  { ssrWidth: 1536 },
);
const largerMobile = breakpoints.greater("mobile");

const params = computed(() => {
  return useRoute().params;
});
const pageInt = computed(() => {
  return Number(params.value.page);
});

const clamp = (num: number, min: number, max: number) => {
  return Math.min(Math.max(num, min), max);
};

const { data } = await useFetch(`/api/a/${params.value.id}`, {
  key: "archive_page_count",
  pick: ["page_count"] as any,
});

const goBack = () => {
  navigateTo({
    name: "a-id-page",
    params: {
      id: params.value.id,
      page: clamp(pageInt.value - 1, 1, data.value.page_count),
    },
  });
};

const goNext = () => {
  if (pageInt.value === data.value.page_count) {
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
        page: clamp(pageInt.value + 1, 1, data.value.page_count),
      },
    });
  }
};

const handleKeyPress = (event: KeyboardEvent) => {
  if (event.key === "ArrowLeft") {
    goBack();
  } else if (event.key === "ArrowRight") {
    goNext();
  }
};

onMounted(() => {
  window.addEventListener("keydown", handleKeyPress);
});
</script>

<template>
  <div class="flex w-full max-w-200 p-3 m-auto">
    <div
      class="bg-secondary flex grow justify-between w-full items-center h-12 rounded-md"
    >
      <nuxt-link
        title="Return to Gallery"
        class="h-full px-3 rounded-l-md flex items-center justify-center hover:bg-muted-foreground/20"
        :to="{ name: 'a-id', params: { id: params.id } }"
      >
        <Undo2 class="size-5" />
      </nuxt-link>
      <div class="flex items-center gap-3 h-full">
        <div class="flex h-full">
          <nuxt-link
            title="First page"
            class="h-full px-3 flex items-center justify-center hover:bg-muted-foreground/20"
            :to="{ name: 'a-id-page', params: { id: params.id, page: 1 } }"
          >
            <ChevronsLeft class="size-5" />
          </nuxt-link>
          <nuxt-link
            title="Previous page"
            class="h-full px-3 flex items-center justify-center hover:bg-muted-foreground/20"
            :to="{
              name: 'a-id-page',
              params: {
                id: params.id,
                page: clamp(pageInt - 1, 1, data.page_count),
              },
            }"
          >
            <ChevronLeft class="size-5" />
          </nuxt-link>
        </div>
        <div class="flex gap-1 flex-wrap">
          <span class="font-semibold">{{ pageInt }}</span>
          <span v-if="largerMobile">of</span>
          <span v-if="largerMobile" class="font-semibold">{{
            data.page_count
          }}</span>
        </div>
        <div class="flex h-full">
          <nuxt-link
            title="Next page"
            class="h-full px-3 flex items-center justify-center hover:bg-muted-foreground/20"
            :to="{
              name: 'a-id-page',
              params: {
                id: params.id,
                page: clamp(pageInt + 1, 1, data.page_count),
              },
            }"
          >
            <ChevronRight class="size-5" />
          </nuxt-link>
          <nuxt-link
            class="h-full px-3 flex items-center justify-center hover:bg-muted-foreground/20"
            title="Last page"
            :to="{
              name: 'a-id-page',
              params: { id: params.id, page: data.page_count },
            }"
          >
            <ChevronsRight class="size-5" />
          </nuxt-link>
        </div>
      </div>
      <ReaderSettingsDialog />
    </div>
  </div>
</template>
