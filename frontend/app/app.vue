<script setup lang="ts">
import type { LayoutKey } from "#build/types/layouts";
import { Toaster } from "./components/ui/sonner";
import "vue-sonner/style.css";

const { isMobile } = useDevice();
const route = useRoute();

const layoutName = computed(() => {
  if (route.meta.layout && route.meta.layout !== ("false" as LayoutKey)) {
    return route.meta.layout;
  }

  return isMobile ? "mobile" : "default";
});

useHead({
  titleTemplate: (name) => {
    return name ? `${name} | Shoka` : "Shoka";
  },
});
</script>

<template>
  <NuxtLoadingIndicator />
  <NuxtLayout :name="layoutName">
    <NuxtPage />
  </NuxtLayout>
  <Toaster rich-colors close-button theme="dark" />
</template>
