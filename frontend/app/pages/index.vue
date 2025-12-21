<script setup lang="ts">
const { token } = useAuth();

const { data: recentlyReleased } = await useFetch("/api/a/", {
  query: { page: 1, size: 14, sortby: "release_date" },
  key: "recentlyReleased",
  onRequest({ options }) {
    options.headers.set("Authorization", `${token.value}`);
  },
  onResponse({ response }) {
    if (response._data.total != 0) {
      generateCover(response._data.archives);
    }
  },
});

const { data: recentlyAdded } = await useFetch("/api/a/", {
  query: { page: 1, size: 14, sortby: "created_at", sortdir: "desc" },
  key: "recentlyAdded",
  onRequest({ options }) {
    options.headers.set("Authorization", `${token.value}`);
  },
  onResponse({ response }) {
    if (response._data.total != 0) {
      generateCover(response._data.archives);
    }
  },
});

const { data: recentlyRead } = await useFetch("/api/a/recent", {
  key: "recentlyRead",
  onRequest({ options }) {
    options.headers.set("Authorization", `${token.value}`);
  },
  onResponse({ response }) {
    if (response._data.total != 0) {
      generateCover(response._data.archives);
    }
  },
});
</script>

<template>
  <div class="flex flex-col gap-4">
    <div
      v-if="recentlyRead && recentlyRead.total > 0"
      class="flex flex-col gap-2"
    >
      <CarouselTitle
        class="px-4 md:px-14 lg:px-16 xl:px-18"
        title="Recently read"
        to="a"
      />
      <ArchiveCarousel
        :archives="recentlyRead"
        :auto-play="false"
        class="overflow-x-hidden md:px-10 xl:px-8 2xl:px-6"
      />
    </div>
    <div
      v-if="recentlyReleased && recentlyReleased.total > 0"
      class="flex flex-col gap-2"
    >
      <CarouselTitle
        class="px-4 md:px-14 lg:px-16 xl:px-18"
        title="Recent Releases"
        to="a"
      />
      <ArchiveCarousel
        :archives="recentlyReleased"
        :auto-play="true"
        class="overflow-x-hidden md:px-10 xl:px-8 2xl:px-6"
      />
    </div>
    <div
      v-if="recentlyAdded && recentlyAdded.total > 0"
      class="flex flex-col gap-2"
    >
      <CarouselTitle
        class="px-4 md:px-14 lg:px-16 xl:px-18"
        title="Recently Added"
        to="a"
      />
      <ArchiveCarousel
        :archives="recentlyAdded"
        :auto-play="true"
        class="overflow-x-hidden md:px-10 xl:px-8 2xl:px-6"
      />
    </div>
  </div>
</template>
