<script setup lang="ts">
import type { NuxtError } from "#app";

const props = defineProps<{ error: NuxtError }>();
</script>

<template>
  <div class="h-screen flex flex-col justify-center items-center">
    <h1 class="text-xl">{{ error.status }}</h1>
    <NuxtLink class="underline underline-offset-2 hover:cursor-pointer">
      Back to the home page
    </NuxtLink>
  </div>
</template>
