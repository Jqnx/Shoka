<script lang="ts" setup>
  import {
    DateFormatter,
    getLocalTimeZone,
    parseAbsolute,
  } from "@internationalized/date";
  import { toDate } from "reka-ui/date";
  import Badge from "./ui/badge/Badge.vue";
  const { data: archive } = useNuxtData("archive");

  const df = new DateFormatter("en-GB", {
    dateStyle: "short",
    timeStyle: "long",
  });
</script>

<template>
  <div class="flex flex-col gap-1 overflow-auto">
    <!-- Artists -->
    <div
      class="flex gap-1.5 flex-wrap sm:gap-0 sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6">
      <p class="text-md font-medium text-slate-400">Artists</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge
          v-if="!archive.artist"
          class="bg-gray-700/60 rounded-sm hover:bg-gray-700">
          null
        </Badge>
        <div v-for="artist in archive.artist" v-else :key="artist.id">
          <Badge
            class="bg-gray-700/60 font-normal rounded-sm hover:bg-gray-700">
            <NuxtLink
              :to="{ name: 'artist-artist', params: { artist: artist.name } }"
              >{{ artist.name }}</NuxtLink
            >
          </Badge>
        </div>
      </div>
    </div>

    <!-- Tags -->
    <div
      class="flex gap-1.5 flex-wrap sm:gap-0 sm:grid sm:grid-cols-6 md:grid-cols-5 2xl:grid-cols-6">
      <p class="text-md font-medium text-slate-400">Tags</p>
      <div class="flex gap-1 flex-wrap col-span-full col-start-2">
        <Badge
          v-if="!archive.tags"
          class="bg-gray-700/60 rounded-sm hover:bg-gray-700">
          null</Badge
        >
        <div v-for="tag in archive.tags" v-else :key="tag.id">
          <Badge
            class="bg-gray-700/60 font-normal rounded-sm hover:bg-gray-700">
            <NuxtLink :to="{ name: 'tag-tag', params: { tag: tag.name } }">{{
              tag.name
            }}</NuxtLink>
          </Badge>
        </div>
      </div>
    </div>

    <!-- Parodies -->
    <div
      class="flex gap-1.5 flex-wrap sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6">
      <p class="text-md font-medium text-slate-400">Parodies</p>
      <div class="flex gap-1 flex-wrap col-start-2 col-span-full">
        <Badge
          v-if="!archive.parody"
          class="bg-gray-700/60 rounded-sm hover:bg-gray-700">
          null
        </Badge>
        <div v-for="parody in archive.parody" v-else :key="parody.id">
          <Badge
            class="bg-gray-700/60 font-normal rounded-sm hover:bg-gray-700">
            <NuxtLink
              :to="{ name: 'parody-parody', params: { parody: parody.name } }"
              >{{ parody.name }}</NuxtLink
            >
          </Badge>
        </div>
      </div>
    </div>

    <!-- Characters -->
    <div
      class="flex gap-1.5 flex-wrap sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6">
      <p class="text-md font-medium text-slate-400">Characters</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge
          v-if="!archive.character"
          class="bg-gray-700/60 rounded-sm hover:bg-gray-700">
          null
        </Badge>
        <div v-for="character in archive.character" v-else :key="character.id">
          <Badge
            class="bg-gray-700/60 font-normal rounded-sm hover:bg-gray-700">
            <NuxtLink
              :to="{
                name: 'character-character',
                params: { character: character.name },
              }"
              >{{ character.name }}</NuxtLink
            >
          </Badge>
        </div>
      </div>
    </div>

    <!-- Languages -->
    <div
      class="flex gap-1.5 sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6">
      <p class="text-md font-medium text-slate-400">Languages</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge
          v-if="!archive.language"
          class="bg-gray-700/60 rounded-sm hover:bg-gray-700">
          null
        </Badge>
        <div v-else :key="archive.archive_id">
          <Badge
            class="bg-gray-700/60 font-normal rounded-sm hover:bg-gray-700">
            <NuxtLink
              :to="{ name: 'lang-lang', params: { lang: archive.language } }"
              >{{ archive.language }}</NuxtLink
            >
          </Badge>
        </div>
      </div>
    </div>

    <!-- Categories -->
    <div
      class="flex gap-1.5 sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6">
      <p class="text-md font-medium text-slate-400">Categories</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge
          v-if="!archive.category"
          class="bg-gray-700/60 rounded-sm hover:bg-gray-700">
          null
        </Badge>
        <div v-else :key="archive.archive_id">
          <Badge
            class="bg-gray-700/60 font-normal rounded-sm hover:bg-gray-700">
            <NuxtLink
              :to="{
                name: 'category-category',
                params: { category: archive.category },
              }"
              >{{ archive.category }}</NuxtLink
            >
          </Badge>
        </div>
      </div>
    </div>

    <!-- Created At -->
    <div
      class="flex gap-1.5 sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6">
      <p class="text-md font-medium text-slate-400">Created</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge class="bg-gray-700/60 font-normal rounded-sm hover:bg-gray-700">
          <NuxtTime
            :datetime="archive.created_at"
            locale="en-GB"
            :title="
              df.format(
                toDate(parseAbsolute(archive.created_at, getLocalTimeZone()))
              )
            " />
        </Badge>
      </div>
    </div>

    <!-- Updated At -->
    <div
      class="flex gap-1.5 sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6">
      <p class="text-md font-medium text-slate-400">Updated</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge class="bg-gray-700/60 font-normal rounded-sm hover:bg-gray-700">
          <NuxtTime
            :datetime="archive.updated_at"
            locale="en-GB"
            :title="
              df.format(
                toDate(parseAbsolute(archive.updated_at, getLocalTimeZone()))
              )
            " />
        </Badge>
      </div>
    </div>

    <!-- URLs -->
    <!--TODO: URLs with icons probably-->
    <div
      class="flex gap-1.5 sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6">
      <p class="text-md font-medium text-slate-400">URLs</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge v-if="!archive.url" class="bg-gray-700 hover:bg-gray-700">
          null
        </Badge>
        <div v-for="url in archive.url" v-else :key="url.id">
          <Badge class="bg-gray-700/60 font-normal hover:bg-gray-700">
            <NuxtLink :to="`${url.url}`" target="_blank">
              {{ url.url }}</NuxtLink
            >
          </Badge>
        </div>
      </div>
    </div>
  </div>
</template>
