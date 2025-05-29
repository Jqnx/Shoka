<script lang="ts" setup>
  import { NuxtImg } from "#components";

  const { id } = useRoute().params;

  const { data: archive } = await useFetch(`/api/a/${id}`, {
    onResponse({ response }) {
      if (response._data.pages != response._data.page_count) {
        useFetch(`/api/a/${id}/thumb`, {
          method: "POST",
        });
      }
    },
    key: "archive",
  });
</script>

<!--TODO: Mobile UI -->
<!--TODO: Fix cover scaling with archive details -->

<template>
  <div>
    <div class="flex flex-col gap-4 p-4 px-24">
      <div class="flex gap-10 justify-center py-8 bg-muted/75 rounded-xl">
        <div class="w-2/5 flex items-center justify-center">
          <figure class="flex items-center w-3/4">
            <NuxtLink :to="`${id}/1`">
              <NuxtImg
                v-slot="{ src, isLoaded, imgAttrs }"
                :src="`/archive/${id}/cover`"
                sizes="500px"
                class="rounded-sm"
                :custom="true">
                <img
                  v-if="isLoaded"
                  v-bind="imgAttrs"
                  :src="src" />
                <!--TODO: FIX SKELETON WHILE IMAGE IS LOADING 
              <Skeleton v-else class="absolute rounded-xl size-full" />
              -->
              </NuxtImg>
            </NuxtLink>
          </figure>
        </div>
        <div class="w-3/5 flex items-center mx-4">
          <ArchiveDetails />
        </div>
      </div>
      <div class="grid grid-cols-6 gap-4">
        <div
          v-for="page in archive.page_count"
          :key="page"
          class="hover:opacity-50">
          <NuxtLink
            :to="`${id}/${page}`"
            class="flex justify-center">
            <NuxtImg
              v-slot="{ src, isLoaded, imgAttrs }"
              :src="`/archive/${id}/${page}`"
              sizes="300px"
              class="rounded-sm"
              :custom="true">
              <img
                v-if="isLoaded"
                v-bind="imgAttrs"
                :src="src" />
              <!--TODO: FIX SKELETON WHILE IMAGE IS LOADING 
                <Skeleton v-else class="absolute rounded-xl size-full" />
                -->
            </NuxtImg>
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>
