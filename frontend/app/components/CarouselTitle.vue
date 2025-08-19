<script lang="ts" setup>
import { ArrowRight } from "lucide-vue-next";
defineProps<{
  title: string;
  to: string;
}>();
</script>

<template>
  <div class="flex flex-1 justify-between px-2">
    <NuxtLink :to="to" class="font-semibold text-xl tracking-tight">
      {{ title }}
    </NuxtLink>
    <NuxtLink :to="to">
      <ArrowRight class="size-5" />
    </NuxtLink>
  </div>
</template>
