<script setup lang="ts">
  import { Progress } from "@/components/ui/progress";
  const props = defineProps({
    id: {
      type: String,
      default: "",
    },
    title: {
      type: String,
      default: "",
    },
    page_count: {
      type: Number,
      default: 0,
    },
    progress: {
      type: Number,
      default: 0,
    },
  });

  const progressValue = computed(() => {
    return (props.progress / props.page_count) * 100;
  });
</script>

<template>
  <article class="group relative rounded-2xl overflow-hidden">
    <NuxtLink :to="`/a/${id}`" class="block pt-[140%]">
      <Progress
        class="z-1 bg-transparent h-1"
        :model-value="progressValue"
        :title="`${props.progress} of ${props.page_count} pages read`" />
      <div class="absolute inset-0">
        <figure>
          <NuxtImg
            v-slot="{ src, isLoaded, imgAttrs }"
            :src="`/archive/${id}/cover`"
            sizes="500px"
            class="text-transparent object-cover object-center absolute inset-0 size-full group-hover:opacity-50"
            :custom="true">
            <img v-if="isLoaded" v-bind="imgAttrs" :src="src" />
          </NuxtImg>
          <div
            class="absolute inset-0 bg-radial from-transparent from-30% to-background/35" />
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
