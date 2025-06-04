<script lang="ts" setup>
  import {
    DateFormatter,
    getLocalTimeZone,
    parseAbsolute,
  } from "@internationalized/date";
  import { toDate } from "reka-ui/date";
  const { data: archive } = useNuxtData("archive");

  const df = new DateFormatter("en-GB", {
    dateStyle: "short",
    timeStyle: "long",
  });
</script>

<template>
  <div class="flex flex-col gap-1 overflow-auto">
    <!-- Title -->
    <h1 class="scroll-m-20 text-2xl font-extrabold tracking-tight text-text">
      {{ archive.title }}
    </h1>

    <!-- Summary -->
    <p class="scroll-m-20 text-lg font-semibold tracking-tight text-text py-1">
      {{ archive.summary }}
    </p>

    <!-- ID -->
    <p class="text-md font-semibold py-2 text-text">
      #{{ archive.archive_id }}
    </p>

    <!-- Artists -->
    <div class="flex gap-1.5">
      <p class="text-md font-semibold text-text">Artists:</p>
      <Badge v-if="!archive.artist" class="bg-slate-700 hover:bg-slate-600">
        null
      </Badge>
      <div v-for="artist in archive.artist" v-else :key="artist.id">
        <Badge class="bg-slate-700 hover:bg-slate-600">
          <NuxtLink :to="`/artist/${artist.name}`">{{ artist.name }}</NuxtLink>
        </Badge>
      </div>
    </div>

    <!-- Tags -->
    <div class="flex gap-1.5 flex-wrap">
      <p class="text-md font-semibold text-text">Tags:</p>
      <Badge v-if="!archive.tags" class="bg-slate-700 hover:bg-slate-600">
        null</Badge
      >
      <div v-for="tag in archive.tags" v-else :key="tag.id">
        <Badge class="bg-slate-700 hover:bg-slate-600">
          <NuxtLink :to="`/tag/${tag.tag}`">{{ tag.tag }}</NuxtLink>
        </Badge>
      </div>
    </div>

    <!-- Parodies -->
    <div class="flex gap-1.5">
      <p class="text-md font-semibold text-text">Parodies:</p>
      <Badge v-if="!archive.parody" class="bg-slate-700 hover:bg-slate-600">
        null
      </Badge>
      <div v-for="parody in archive.parody" v-else :key="parody.id">
        <Badge class="bg-slate-700 hover:bg-slate-600">
          <NuxtLink :to="`/parody/${parody.parody}`">{{
            parody.parody
          }}</NuxtLink>
        </Badge>
      </div>
    </div>

    <!-- Characters -->
    <div class="flex gap-1.5">
      <p class="text-md font-semibold text-text">Characters:</p>
      <Badge v-if="!archive.character" class="bg-slate-700 hover:bg-slate-600">
        null
      </Badge>
      <div v-for="character in archive.character" v-else :key="character.id">
        <Badge class="bg-slate-700 hover:bg-slate-600">
          <NuxtLink :to="`/character/${character.character}`">{{
            character.character
          }}</NuxtLink>
        </Badge>
      </div>
    </div>

    <!-- Languages -->
    <div class="flex gap-1.5">
      <p class="text-md font-semibold text-text">Languages:</p>
      <Badge v-if="!archive.language" class="bg-slate-700 hover:bg-slate-600">
        null
      </Badge>
      <div v-else :key="archive.archive_id">
        <Badge class="bg-slate-700 hover:bg-slate-600">
          <NuxtLink :to="`/lang/${archive.language}`">{{
            archive.language
          }}</NuxtLink>
        </Badge>
      </div>
    </div>

    <!-- Categories -->
    <div class="flex gap-1.5">
      <p class="text-md font-semibold text-text">Categories:</p>
      <Badge v-if="!archive.category" class="bg-slate-700 hover:bg-slate-600">
        null
      </Badge>
      <div v-else :key="archive.archive_id">
        <Badge class="bg-slate-700 hover:bg-slate-600">
          <NuxtLink :to="`/cat/${archive.category}`">{{
            archive.category
          }}</NuxtLink>
        </Badge>
      </div>
    </div>
    <div class="flex gap-1.5">
      <p class="text-md font-semibold text-text">Pages:</p>
      <Badge class="bg-slate-700 hover:bg-slate-600">
        {{ archive.page_count }}
      </Badge>
    </div>

    <!-- Created At -->
    <div class="flex gap-1.5">
      <p class="text-md font-semibold text-text">Created:</p>
      <Badge class="bg-slate-700 hover:bg-slate-600">
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

    <!-- Updated At -->
    <div class="flex gap-1.5">
      <p class="text-md font-semibold text-text">Updated:</p>
      <Badge class="bg-slate-700 hover:bg-slate-600">
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

    <!-- Release Date -->
    <div class="flex gap-1.5">
      <p class="text-md font-semibold text-text">Release Date:</p>
      <Badge
        v-if="!archive.release_date"
        class="bg-slate-700 hover:bg-slate-600">
        null
      </Badge>
      <Badge v-else class="bg-slate-700 hover:bg-slate-600">
        <NuxtTime
          :datetime="archive.release_date"
          locale="en-GB"
          :title="
            df.format(
              toDate(parseAbsolute(archive.release_date, getLocalTimeZone()))
            )
          " />
      </Badge>
    </div>

    <!-- URLs -->
    <!--TODO: URLs with icons probably-->
    <div class="flex gap-1.5">
      <p class="text-md font-semibold text-text">URLs:</p>
      <Badge v-if="!archive.url" class="bg-slate-700 hover:bg-slate-600">
        null
      </Badge>
      <div v-for="url in archive.url" v-else :key="url.id">
        <Badge class="bg-slate-700 hover:bg-slate-600">
          <NuxtLink :to="`${url.url}`"> {{ url.url }}</NuxtLink>
        </Badge>
      </div>
    </div>
  </div>
</template>
