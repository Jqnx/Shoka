<script lang="ts" setup>
  import type { CarouselApi } from "@/components/ui/carousel";
  import {
    Carousel,
    CarouselContent,
    CarouselItem,
    CarouselNext,
    CarouselPrevious,
  } from "@/components/ui/carousel";
  import Autoplay from "embla-carousel-autoplay";

  defineProps<{
    title?: string;
  }>();

  const { data: archives } = await useFetch("/api/a/", {
    query: { page: 1, size: 14, sortby: "release_date" },
    key: "recentArchives",
  });

  const emblaMainApi = ref<CarouselApi>();
  const emblaThumbnailApi = ref<CarouselApi>();
  const selectedIndex = ref(0);

  function onSelect() {
    if (!emblaMainApi.value || !emblaThumbnailApi.value) return;
    selectedIndex.value = emblaMainApi.value.selectedScrollSnap();
    emblaThumbnailApi.value.scrollTo(emblaMainApi.value.selectedScrollSnap());
  }

  function onThumbClick(index: number) {
    if (!emblaMainApi.value || !emblaThumbnailApi.value) return;
    emblaMainApi.value.scrollTo(index);
  }

  watchOnce(emblaMainApi, (emblaMainApi) => {
    if (!emblaMainApi) return;

    onSelect();
    emblaMainApi.on("select", onSelect);
    emblaMainApi.on("reInit", onSelect);
  });
</script>

<template>
  <div class="flex flex-col items-center w-full gap-1">
    <Carousel
      class="relative w-full max-w-19/20"
      :opts="{ align: 'start', loop: true }"
      :plugins="[
        Autoplay({
          delay: 2500,
        }),
      ]"
      @init-api="(val) => (emblaMainApi = val)">
      <CarouselContent>
        <CarouselItem
          v-for="(archive, index) in archives.archives"
          :key="index"
          class="md:basis-1/2 lg:basis-1/6">
          <div class="p-1">
            <GalleryItem :id="archive.archive_id" :title="archive.title" />
          </div>
        </CarouselItem>
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>
    <Carousel
      class="relative max-w-28 w-full"
      :opts="{
        align: 'center',
        loop: true,
      }"
      @init-api="(val) => (emblaThumbnailApi = val)">
      <CarouselContent class="flex gap-0 ml-0">
        <CarouselItem
          v-for="(_, index) in archives.archives"
          :key="index"
          class="pl-0 basis-1/5 cursor-pointer"
          @click="onThumbClick(index)">
          <CarouselDot
            :class="index === selectedIndex ? 'bg-primary' : 'bg-secondary'" />
        </CarouselItem>
      </CarouselContent>
    </Carousel>
  </div>
</template>
