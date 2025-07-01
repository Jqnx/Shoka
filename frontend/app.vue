<script setup lang="ts">
  import { Toaster } from "./components/ui/sonner";
  import "vue-sonner/style.css";
</script>

<template>
  <NuxtLoadingIndicator />
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <Toaster rich-colors close-button theme="dark" />
</template>
