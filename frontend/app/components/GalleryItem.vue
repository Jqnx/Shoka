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
  <NuxtLink :to="{ name: 'a-id', params: { id: id } }" class="group">
    <article class="group relative overflow-hidden">
      <div class="block pt-[140%]">
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
              :custom="true"
            >
              <img
                v-if="isLoaded"
                v-bind="imgAttrs"
                :src="src"
                class="object-cover object-center rounded-2xl absolute inset-0 size-full"
              />
            </NuxtImg>
            <div
              class="absolute inset-0 bg-radial rounded-2xl from-transparent from-30% to-shadow/35"
            />
          </figure>
        </div>
      </div>
    </article>
    <h4
      class="font-medium text-sm/6 text-wrap tracking-normal line-clamp-2 pt-2 text-muted-foreground transition-all group-hover:text-primary"
      :title="title"
    >
      {{ title }}
    </h4>
  </NuxtLink>
</template>
