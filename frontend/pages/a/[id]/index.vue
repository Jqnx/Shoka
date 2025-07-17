<script lang="ts" setup>
  import { HeartMinus, HeartPlus, Pencil } from "lucide-vue-next";

  const { id } = useRoute().params;

  const { token } = useAuth();

  const { data: archive } = await useFetch(`/api/a/${id}`, {
    onRequest({ options }) {
      options.headers.set("Authorization", `${token.value}`);
    },
    onResponse({ response }) {
      if (response._data.pages != response._data.page_count) {
        useFetch(`/api/a/${id}/thumb`, {
          method: "POST",
        });
      }
    },
    key: "archive",
  });

  async function favorite() {
    return $fetch(`/api/a/${id}/favorite`, {
      method: "post",
      onRequest({ options }) {
        options.headers.set("Authorization", `${token.value}`);
        archive.value.is_favorite = true;
      },
      onResponseError() {
        archive.value.is_favorite = false;
      },
      async onResponse() {
        refreshNuxtData("archive");
      },
    });
  }

  const ArchiveDetails = resolveComponent("ArchiveDetails");
  const ArchiveDetailsForm = resolveComponent("ArchiveDetailsForm");
  const toggle = ref(true);
</script>

<!--TODO: Mobile UI -->
<!--TODO: Fix cover scaling with archive details -->

<template>
  <div>
    <div class="flex flex-col gap-4 p-4 xl:px-24">
      <div
        class="flex flex-col justify-center py-8 bg-muted/70 rounded-xl lg:flex-row">
        <div class="w-full 2xl:w-2/5">
          <!-- Archive Cover -->
          <figure class="w-3/4 pb-4 m-auto">
            <NuxtLink :to="{ name: 'a-id-page', params: { id: id, page: 1 } }">
              <NuxtImg
                :src="`/archive/${id}/cover`"
                sizes="450px"
                class="rounded-sm m-auto">
                <!--TODO: FIX SKELETON WHILE IMAGE IS LOADING 
              <Skeleton v-else class="absolute rounded-xl size-full" />
              -->
              </NuxtImg>
            </NuxtLink>
          </figure>
        </div>
        <!-- Archive Details -->
        <div class="w-full 2xl:w-3/5 flex flex-col px-4 md:pr-4">
          <component
            :is="toggle ? ArchiveDetails : ArchiveDetailsForm"
            @back="toggle = !toggle" />
          <div class="flex gap-2 my-4">
            <Button
              v-if="toggle"
              class="rounded-sm items-center cursor-pointer"
              @click="toggle = !toggle">
              <Pencil />
              <span>Edit</span>
            </Button>
            <div v-if="toggle">
              <Button
                v-if="archive.is_favorite"
                class="rounded-sm items-center bg-destructive hover:bg-destructive/90 cursor-pointer"
                @click="favorite">
                <HeartMinus />
                <span>Favorite</span>
              </Button>
              <Button
                v-else
                class="rounded-sm items-center cursor-pointer"
                @click="favorite">
                <HeartPlus />
                <span>Favorite</span>
              </Button>
            </div>
          </div>
        </div>
      </div>
      <!-- Thumbnail Gallery -->
      <div
        class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-6 gap-4">
        <div
          v-for="page in archive.page_count"
          :key="page"
          class="hover:opacity-50">
          <NuxtLink
            :to="{ name: 'a-id-page', params: { id: id, page: page } }"
            class="flex justify-center"
            no-prefetch>
            <NuxtImg
              :src="`/archive/${id}/${page}`"
              sizes="300px"
              class="rounded-sm">
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
