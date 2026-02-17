<script lang="ts" setup>
import {
  CalendarArrowUp,
  Files,
  Heart,
  EllipsisVertical,
} from "lucide-vue-next";
import { toast } from "vue-sonner";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  DateFormatter,
  getLocalTimeZone,
  parseAbsolute,
} from "@internationalized/date";
import { toDate } from "reka-ui/date";
import { useBreakpoints } from "@vueuse/core";
import { Progress } from "@/components/ui/progress";
import type { Archive } from "~~/shared/types/archive";

const { id } = useRoute().params;
const token = await useAuth().getToken();
const { copy } = useClipboard();

const breakpoints = useBreakpoints(
  {
    mobile: 412,
  },
  { ssrWidth: 1536 },
);
const largerMobile = breakpoints.greater("mobile");

const df = new DateFormatter("en-GB", {
  dateStyle: "short",
  timeStyle: "long",
});

const copyArchiveId = () => {
  copy(archive.value.id);
  toast.success(`Copied ${archive.value.id} to clipboard.`);
};

const { data: archive } = await useFetch<Archive | undefined>(`/api/a/${id}`, {
  onRequest({ options }) {
    options.headers.set("Authorization", `Bearer ${token}`);
  },
  onResponse({ response }) {
    if (response._data.pagesOnDisk != response._data.pageCount) {
      useFetch(`/api/a/${id}/thumb`, {
        method: "POST",
      });
    }
  },
  key: "archive",
});

useHead({
  title: `${archive.value.title}`,
});

async function favorite() {
  return $fetch(`/api/a/${id}/favorite`, {
    method: "post",
    onRequest({ options }) {
      options.headers.set("Authorization", `Bearer ${token}`);
      archive.value.isFavorite = true;
    },
    onResponseError() {
      archive.value.isFavorite = false;
    },
    async onResponse() {
      refreshNuxtData("archive");
    },
  });
}

const generateCover = () => {
  $fetch(`/api/a/${id}/cover`, {
    method: "post",
    onResponseError({ response }) {
      toast.error(response._data.data);
    },
    onResponse({ response }) {
      if (response.ok) {
        toast.success("Successfully generated cover");
      }
    },
  });
};

const toFile = () => {
  $fetch(`/api/a/${id}/meta/tofile`, {
    method: "post",
    onResponseError() {
      toast.error("Failed to save ComicInfo.xml file to archive.");
    },
    onResponse({ response }) {
      if (response.ok) {
        toast.success("Successfully saved ComicInfo.xml file to archive.");
      }
    },
  });
};

const toggle = ref(true);
const progressValue = computed(() => {
  return (archive.value.progress / archive.value.pageCount) * 100;
});
</script>

<!--TODO: Mobile UI -->
<!--TODO: Fix cover scaling with archive details -->

<template>
  <div>
    <div class="flex flex-col gap-4 p-4">
      <div class="bg-secondary rounded-xl">
        <Progress
          class="bg-transparent rounded-b-none"
          :model-value="progressValue"
          :title="`${archive.progress} of ${archive.pageCount} pages read`"
        />
        <div class="flex flex-col justify-center lg:flex-row">
          <div class="w-full 2xl:w-2/5 flex flex-col justify-center">
            <!-- Archive Cover -->
            <figure class="w-full p-4">
              <NuxtLink
                :to="{ name: 'a-id-page', params: { id: id, page: 1 } }"
              >
                <NuxtImg
                  :src="`/archive/${id}/cover`"
                  sizes="450px"
                  class="rounded-sm m-auto"
                >
                  <!--TODO: FIX SKELETON WHILE IMAGE IS LOADING 
              <Skeleton v-else class="absolute rounded-xl size-full" />
              -->
                </NuxtImg>
              </NuxtLink>
            </figure>
          </div>
          <!-- Archive Details -->
          <div class="w-full 2xl:w-3/5 flex flex-col px-4 py-2 md:pr-4">
            <div class="flex flex-col gap-1">
              <!-- Title -->
              <h1 class="scroll-m-20 text-2xl font-extrabold tracking-tight">
                {{ archive.title }}
              </h1>

              <!-- Summary -->
              <p
                class="scroll-m-20 text-lg font-semibold tracking-tight text-foreground/70 py-1"
              >
                {{ archive.summary }}
              </p>

              <div class="flex flex-wrap justify-between gap-2">
                <div class="flex gap-2 xl:gap-4">
                  <!-- ID -->
                  <p
                    class="text-md font-medium cursor-pointer p-0.5 w-max hover:bg-accent hover:rounded-md"
                    title="Click to copy!"
                    @click="copyArchiveId"
                  >
                    <span class="text-foreground/40">#</span>
                    <span class="text-foreground/70">
                      {{ archive.id }}
                    </span>
                  </p>

                  <!-- Page Count -->
                  <div class="flex items-center gap-1.5">
                    <Files class="size-4 stroke-foreground/70 stroke-2" />
                    <p class="text-foreground/70 font-medium">
                      {{ archive.pageCount }}
                      <span v-if="largerMobile">Pages</span>
                    </p>
                  </div>

                  <!-- Release Date -->
                  <div
                    v-if="archive.releaseDate"
                    class="flex items-center gap-1.5"
                  >
                    <CalendarArrowUp
                      class="size-4 stroke-foreground/70 stroke-2"
                    />
                    <p class="text-foreground/70 font-medium">
                      <NuxtTime
                        :datetime="archive.releaseDate"
                        locale="en-GB"
                        :title="
                          df.format(
                            toDate(
                              parseAbsolute(
                                archive.releaseDate,
                                getLocalTimeZone(),
                              ),
                            ),
                          )
                        "
                      />
                    </p>
                  </div>
                </div>
                <div class="flex gap-2">
                  <div v-if="toggle" class="flex items-center">
                    <div v-if="archive.isFavorite" title="Unfavorite Archive">
                      <Heart
                        class="size-5 fill-destructive stroke-destructive cursor-pointer"
                        @click="favorite"
                      />
                    </div>
                    <div v-else class="flex gap-1" title="Favorite Archive">
                      <Heart
                        class="size-5 cursor-pointer stroke-foreground/70"
                        @click="favorite"
                      />
                    </div>
                  </div>
                  <div class="flex items-center justify-center">
                    <DropdownMenu>
                      <DropdownMenuTrigger>
                        <EllipsisVertical
                          class="size-5 stroke-foreground/70 cursor-pointer"
                        />
                      </DropdownMenuTrigger>
                      <DropdownMenuContent>
                        <DropdownMenuItem @click="toFile">
                          Export metadata to ComicInfo
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </div>
              </div>
            </div>

            <div class="py-2">
              <Tabs default-value="details">
                <TabsList class="grid w-full grid-cols-2">
                  <TabsTrigger
                    value="details"
                    class="dark:data-[state=active]:bg-primary dark:data-[state=active]:text-primary-foreground text-foreground h-[calc(100%-2px)] cursor-pointer"
                    >Details</TabsTrigger
                  >
                  <TabsTrigger
                    value="edit"
                    class="dark:data-[state=active]:bg-primary dark:data-[state=active]:text-primary-foreground text-foreground h-[calc(100%-2px)] cursor-pointer"
                    >Edit</TabsTrigger
                  >
                </TabsList>
                <TabsContent value="details">
                  <ArchiveDetails />
                </TabsContent>
                <TabsContent value="edit">
                  <ArchiveDetailsForm />
                </TabsContent>
              </Tabs>
            </div>
          </div>
        </div>
      </div>
      <!-- Thumbnail Gallery -->
      <div
        class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-6 gap-2"
      >
        <div
          v-for="page in archive.pageCount"
          :key="page"
          class="hover:opacity-50"
        >
          <NuxtLink
            :to="{ name: 'a-id-page', params: { id: id, page: page } }"
            class="flex justify-center"
            no-prefetch
          >
            <NuxtImg
              :src="`/archive/${id}/${page}`"
              sizes="300px"
              class="rounded-sm"
            >
              <!--TODO: FIX SKELETON WHILE IMAGE IS LOADING 
                <Skeleton v-else class="absolute rounded-xl size-full" />
                -->
            </NuxtImg>
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>
