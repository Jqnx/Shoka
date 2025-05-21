<script lang="ts">
export default {
  props: {
    id: {
      type: String,
      default: "",
    },
    title: {
      type: String,
      default: "",
    },

  },
}
</script>

<template>
  <article class="group relative rounded-2xl overflow-hidden">
    <NuxtLink :to="`/a/${id}`" class="block pt-[140%]">
      <div class="absolute inset-0">
        <figure>
          <NuxtImg v-slot="{ src, isLoaded, imgAttrs }" :src="`/archive/${id}/cover`" sizes="500px"
            class="text-transparent object-cover object-center absolute inset-0 size-full group-hover:opacity-50"
            :custom="true">
            <img v-if="isLoaded" v-bind="imgAttrs" :src="src">
            <Skeleton v-else class="absolute rounded-xl size-full" />
          </NuxtImg>
          <div class="absolute inset-0 bg-radial from-transparent from-30% to-background/35" />
        </figure>
        <div class="relative flex flex-col h-full">
          <div class="grow h-0" />
          <div class="relative p-4">
            <h4
              class="font-normal text-shadow-md text-shadow-background/75 text-balance break-words tracking-tight text-text line-clamp-2 group-hover:line-clamp-3">
              {{ title }}
            </h4>
          </div>
        </div>
      </div>
    </NuxtLink>
  </article>
</template>
