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
      v-if="archive.artists"
      class="flex gap-1.5 flex-wrap sm:gap-0 sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6"
    >
      <p class="text-md font-medium text-foreground/80">Artists</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge
          v-for="artist in archive.artists"
          :key="artist.id"
          class="bg-border text-foreground font-normal rounded-sm hover:bg-border/80"
        >
          <NuxtLink
            :to="{ name: 'artist-artist', params: { artist: artist.name } }"
            >{{ artist.name }}</NuxtLink
          >
        </Badge>
      </div>
    </div>

    <!-- Tags -->
    <div
      v-if="archive.tags"
      class="flex gap-1.5 flex-wrap sm:gap-0 sm:grid sm:grid-cols-6 md:grid-cols-5 2xl:grid-cols-6"
    >
      <p class="text-md font-medium text-foreground/80">Tags</p>
      <div class="flex gap-1 flex-wrap col-span-full col-start-2">
        <Badge
          v-for="tag in archive.tags"
          :key="tag.id"
          class="bg-border text-foreground font-normal rounded-sm hover:bg-border/80"
        >
          <NuxtLink :to="{ name: 'tag-tag', params: { tag: tag.name } }">{{
            tag.name
          }}</NuxtLink>
        </Badge>
      </div>
    </div>

    <!-- Parodies -->
    <div
      v-if="archive.parodies"
      class="flex gap-1.5 flex-wrap sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6"
    >
      <p class="text-md font-medium text-foreground/80">Parodies</p>
      <div class="flex gap-1 flex-wrap col-start-2 col-span-full">
        <Badge
          v-for="parody in archive.parodies"
          :key="parody.id"
          class="bg-border text-foreground font-normal rounded-sm hover:bg-border/80"
        >
          <NuxtLink
            :to="{ name: 'parody-parody', params: { parody: parody.name } }"
            >{{ parody.name }}</NuxtLink
          >
        </Badge>
      </div>
    </div>

    <!-- Characters -->
    <div
      v-if="archive.characters"
      class="flex gap-1.5 flex-wrap sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6"
    >
      <p class="text-md font-medium text-foreground/80">Characters</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge
          v-for="character in archive.characters"
          :key="character.id"
          class="bg-border text-foreground font-normal rounded-sm hover:bg-border/80"
        >
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

    <!-- Languages -->
    <div
      v-if="archive.language"
      class="flex gap-1.5 sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6"
    >
      <p class="text-md font-medium text-foreground/80">Languages</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge
          class="bg-border text-foreground font-normal rounded-sm hover:bg-border/80"
        >
          <NuxtLink
            :to="{ name: 'lang-lang', params: { lang: archive.language } }"
            >{{ archive.language }}</NuxtLink
          >
        </Badge>
      </div>
    </div>

    <!-- Categories -->
    <div
      v-if="archive.category"
      class="flex gap-1.5 sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6"
    >
      <p class="text-md font-medium text-foreground/80">Categories</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge
          class="bg-border text-foreground font-normal rounded-sm hover:bg-border/80"
        >
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

    <!-- Added At -->
    <div
      class="flex gap-1.5 sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6"
    >
      <p class="text-md font-medium text-foreground/80">Added</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge
          class="bg-border text-foreground font-normal rounded-sm hover:bg-border/80"
        >
          <NuxtTime
            :datetime="archive.createdAt"
            locale="en-GB"
            :title="
              df.format(
                toDate(parseAbsolute(archive.createdAt, getLocalTimeZone())),
              )
            "
          />
        </Badge>
      </div>
    </div>

    <!-- Updated At -->
    <div
      class="flex gap-1.5 sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6"
    >
      <p class="text-md font-medium text-foreground/80">Updated</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge
          class="bg-border text-foreground font-normal rounded-sm hover:bg-border/80"
        >
          <NuxtTime
            :datetime="archive.updatedAt"
            locale="en-GB"
            :title="
              df.format(
                toDate(parseAbsolute(archive.updatedAt, getLocalTimeZone())),
              )
            "
          />
        </Badge>
      </div>
    </div>

    <!-- URLs -->
    <!--TODO: URLs with icons probably-->
    <div
      v-if="archive.url"
      class="flex gap-1.5 sm:grid grid-cols-6 md:grid-cols-5 2xl:grid-cols-6"
    >
      <p class="text-md font-medium text-foreground/80">URLs</p>
      <div class="flex flex-wrap gap-1 col-start-2 col-span-full">
        <Badge
          v-for="url in archive.url"
          :key="url.id"
          class="bg-border text-foreground font-normal hover:bg-border/80"
        >
          <NuxtLink :to="`${url.url}`" target="_blank"> {{ url.url }}</NuxtLink>
        </Badge>
      </div>
    </div>
  </div>
</template>
