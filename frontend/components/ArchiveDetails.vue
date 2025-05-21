<script lang="ts" setup>
const { data: archive } = useNuxtData('archive')
</script>

<!--TODO: Links & Pages for metadata e.g tags, characters etc-->

<template>
   <div class="flex flex-col gap-0.5">
      <h1 class="scroll-m-20 text-4xl font-extrabold tracking-tight text-text lg:text-2xl">{{ archive.title }}</h1>
      <p class="scroll-m-20 text-4xl font-semibold tracking-tight text-text py-1 lg:text-xl">{{ archive.summary
      }}</p>
      <p class="text-md font-semibold py-2 text-text">#{{ archive.archive_id }}</p>
      <div class="flex gap-1.5">
         <p class="text-md font-semibold text-text">Tags:</p>
         <p v-if="!archive.tags"
            class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">null</p>
         <div v-for="tag in archive.tags" v-else :key="tag.id">
            <p class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">{{
               tag }}</p>
         </div>
      </div>
      <div class="flex gap-1.5">
         <p class="text-md font-semibold text-text">Artists:</p>
         <p v-if="!archive.artist"
            class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">null</p>
         <div v-for="artist in archive.artist" v-else :key="artist.id">
            <p class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">{{
               artist }}
            </p>
         </div>
      </div>
      <div class="flex gap-1.5">
         <p class="text-md font-semibold text-text">Parodies:</p>
         <p v-if="!archive.parody"
            class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">null</p>
         <div v-for="parody in archive.parody" v-else :key="parody.id">
            <p class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">{{
               parody }}
            </p>
         </div>
      </div>
      <div class="flex gap-1.5">
         <p class="text-md font-semibold text-text">Characters:</p>
         <p v-if="!archive.character"
            class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">null</p>
         <div v-for="character in archive.character" v-else :key="character.id">
            <p class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">{{
               character }}
            </p>
         </div>
      </div>
      <div class="flex gap-1.5">
         <p class="text-md font-semibold text-text">Languages:</p>
         <p v-if="!archive.language"
            class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">null</p>
         <div v-for="language in archive.language" v-else :key="language">
            <p class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">{{
               language }}
            </p>
         </div>
      </div>
      <div class="flex gap-1.5">
         <p class="text-md font-semibold text-text">Categories:</p>
         <p v-if="!archive.category"
            class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">null</p>
         <div v-for="category in archive.category" v-else :key="category.id">
            <p class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">{{
               category }}
            </p>
         </div>
      </div>
      <div class="flex gap-1.5">
         <p class="text-md font-semibold text-text">Pages:</p>
         <p class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">{{
            archive.page_count }}</p>
      </div>
      <div class="flex gap-1.5">
         <p class="text-md font-semibold text-text">URLs:</p>
         <p v-if="!archive.url"
            class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">null</p>
         <div v-for="url in archive.url" v-else :key="url.id">
            <p class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">{{
               url }}</p>
         </div>
      </div>
      <div class="flex gap-1.5">
         <p class="text-md font-semibold text-text">Type:</p>
         <p v-if="!archive.type"
            class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">null</p>
         <p v-else class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">
            {{
               archive.type }}</p>
      </div>
      <div class="flex gap-1.5">
         <p class="text-md font-semibold text-text">Created:</p>
         <p class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">{{
            archive.created_at }}</p>
      </div>
      <div class="flex gap-1.5">
         <p class="text-md font-semibold text-text">Updated:</p>
         <p class="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold text-text">{{
            archive.updated_at }}</p>
      </div>
   </div>
</template>
