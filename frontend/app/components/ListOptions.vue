<script lang="ts" setup>
import { ChevronDown, ChevronUp, Shuffle } from "lucide-vue-next";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from "@/components/ui/dropdown-menu";
import Button from "@/components/ui/button/Button.vue";

const route = useRoute();

const filters = computed(() => {
  if (route.name === "favorites") {
    return storeToRefs(useFavoriteFiltersStore());
  } else {
    return storeToRefs(useFiltersStore());
  }
});

const totalArchives = computed(() => {
  if (route.name === "favorites") {
    const fav = useNuxtData("favorites");
    return fav.data.value.total;
  } else {
    const arch = useNuxtData("archives");
    return arch.data.value.total;
  }
});

const isFav = computed(() => {
  if (route.name === "favorites") {
    return true;
  } else {
    return false;
  }
});

const sortList = [
  { value: "title", label: "Title" },
  { value: "page_count", label: "Page Count" },
  { value: "created_at", label: "Created At" },
  { value: "updated_at", label: "Updated At" },
  { value: "release_date", label: "Release Date" },
];

const sortListFavorite = [
  { value: "title", label: "Title" },
  { value: "page_count", label: "Page Count" },
  { value: "created_at", label: "Created At" },
  { value: "updated_at", label: "Updated At" },
  { value: "release_date", label: "Release Date" },
  { value: "favorited_at", label: "Favorited At" },
];

const { token } = useAuth();

const shuffle = () => {
  // generate number with total archives as max
  const num = Math.floor(Math.random() * totalArchives.value);
  // fetch archive from backend with number as query 'c' and favorite 'true' or 'false'
  $fetch("/api/a/shuffle", {
    method: "POST",
    onRequest({ options }) {
      if (isFav.value) {
        options.headers.set("Authorization", `${token.value}`);
      }
    },
    query: {
      c: num,
      favorite: isFav.value,
    },
    body: {
      tags: filters.value.filters.value.tags,
      artists: filters.value.filters.value.artists,
      characters: filters.value.filters.value.characters,
      parodies: filters.value.filters.value.parodies,
      languages: filters.value.filters.value.languages,
      categories: filters.value.filters.value.categories,
    },
    onResponse({ response }) {
      navigateTo({ name: "a-id", params: { id: response._data } });
    },
  });
};
</script>

<template>
  <div class="flex justify-between">
    <div class="flex gap-2">
      <div>
        <FilterDialog />
      </div>
      <div class="flex">
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button variant="secondary" class="rounded-r-none">
              {{ filters.sortLabel }}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent v-if="route.name === 'favorites'">
            <DropdownMenuItem
              v-for="(item, index) in sortListFavorite"
              :key="index"
              @select="
                () => {
                  filters.sortBy.value = item.value;
                  filters.sortLabel.value = item.label;
                }
              "
            >
              {{ item.label }}
            </DropdownMenuItem>
          </DropdownMenuContent>
          <DropdownMenuContent v-else>
            <DropdownMenuItem
              v-for="(item, index) in sortList"
              :key="index"
              @select="
                () => {
                  filters.sortBy.value = item.value;
                  filters.sortLabel.value = item.label;
                }
              "
            >
              {{ item.label }}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <Button
          v-if="filters.sortDir.value == 'asc'"
          variant="secondary"
          class="rounded-l-none"
          size="icon"
          @click="filters.sortDir.value = 'desc'"
        >
          <ChevronUp />
        </Button>
        <Button
          v-else
          variant="secondary"
          class="rounded-l-none"
          size="icon"
          @click="filters.sortDir.value = 'asc'"
        >
          <ChevronDown />
        </Button>
      </div>
    </div>
    <div>
      <Button variant="secondary" @click="shuffle">
        <Shuffle />
      </Button>
    </div>
  </div>
</template>
