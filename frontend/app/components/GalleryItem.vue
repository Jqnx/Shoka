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
  pageCount: {
    type: Number,
    default: 0,
  },
  progress: {
    type: Number,
    default: 0,
  },
});

const progressValue = computed(() => {
  return (props.progress / props.pageCount) * 100;
});
</script>

<template>
  <article class="group relative overflow-hidden">
    <NuxtLink :to="`/a/${id}`" class="block pt-[140%]">
      <Progress
        v-if="props.progress"
        class="z-1 rounded-2xl bg-transparent h-1 rounded-tl-none rounded-tr-none"
        :model-value="progressValue"
        :title="`${props.progress} of ${props.pageCount} pages read`"
      />
      <div class="absolute inset-0">
        <figure>
          <NuxtImg
            v-slot="{ src, isLoaded, imgAttrs }"
            :src="`/archive/${id}/cover`"
            sizes="500px"
            class="object-cover object-center rounded-2xl absolute inset-0 size-full group-hover:opacity-65"
            :custom="true"
          >
            <img v-if="isLoaded" v-bind="imgAttrs" :src="src" />
          </NuxtImg>
          <div
            class="absolute inset-0 bg-radial rounded-2xl from-transparent from-30% to-shadow/35"
          />
        </figure>
        <div layout class="relative flex flex-col h-full">
          <div class="grow" />
          <div
            layout
            class="relative flex h-16 p-3 bg-background rounded-t-2xl shadow-[0px_-3px_5px_0px_rgba(25,23,36,0.5)] group-hover:bg-background/65 group-hover:h-full group-hover:rounded-none transition-all ease-in-out duration-250"
          >
            <h4
              class="font-medium text-sm text-wrap tracking-normal transition-all group-hover:my-auto line-clamp-2 group-hover:line-clamp-none"
            >
              {{ title }}
            </h4>
          </div>
        </div>
      </div>
    </NuxtLink>
  </article>
</template>
