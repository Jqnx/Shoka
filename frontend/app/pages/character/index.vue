<script setup lang="ts">
  import Button from "~/components/ui/button/Button.vue";
  import Separator from "~/components/ui/separator/Separator.vue";

  useHead({
    title: "Characters",
  });

  const { data: characters } = await useFetch("/api/character", {
    key: "characters",
  });
</script>

<template>
  <div>
    <h1
      class="flex justify-center py-4 text-text font-semibold text-2xl tracking-tight">
      Characters
    </h1>
    <div
      class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6 gap-1.5 px-8 py-4">
      <NuxtLink
        v-for="(character, index) in characters"
        :key="index"
        :to="{
          name: 'character-character',
          params: { character: character.name },
        }">
        <Button size="sm" class="w-full rounded-sm bg-stone-700 cursor-pointer">
          <div class="flex flex-1 justify-between">
            <p>{{ character.name }}</p>
            <div class="flex gap-2">
              <Separator orientation="vertical" />
              <p class="font-light text-stone-400">{{ character.count }}</p>
            </div>
          </div>
        </Button>
      </NuxtLink>
    </div>
  </div>
</template>
