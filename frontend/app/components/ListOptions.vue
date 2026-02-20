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
const token = await useAuth().getToken();

const props = defineProps({
  useFilters: Boolean,
  sortList: Object,
});

const sortBy = defineModel("sortBy");
const sortLabel = defineModel("sortLabel");
const sortDir = defineModel("sortDir");
//const filters = defineModel("filters");

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

const shuffle = () => {
  // generate number with total archives as max
  const num = Math.floor(Math.random() * totalArchives.value);
  // fetch archive from backend with number as query 'c' and favorite 'true' or 'false'
  $fetch("/api/a/shuffle", {
    method: "POST",
    onRequest({ options }) {
      if (isFav.value) {
        options.headers.set("Authorization", `Bearer ${token}`);
      }
    },
    query: {
      c: num,
      favorite: isFav.value,
    },
    //body: {
    //  tags: filters.value.filters.value.tags,
    //  artists: filters.value.filters.value.artists,
    //  characters: filters.value.filters.value.characters,
    //  parodies: filters.value.filters.value.parodies,
    //  languages: filters.value.filters.value.languages,
    //  categories: filters.value.filters.value.categories,
    //},
    onResponse({ response }) {
      navigateTo({ name: "a-id", params: { id: response._data } });
    },
  });
};
</script>

<template>
  <div class="flex justify-between">
    <div class="flex gap-2">
      <div v-if="props.useFilters">
        <FilterDialog />
      </div>
      <div class="flex">
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button
              variant="ghost"
              class="rounded-r-none text-muted-foreground"
              title="Sort By"
            >
              {{ sortLabel }}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuItem
              v-for="(item, index) in sortList"
              :key="index"
              class="text-muted-foreground font-medium"
              @select="
                () => {
                  sortBy = item.value;
                  sortLabel = item.label;
                }
              "
            >
              {{ item.label }}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <Button
          v-if="sortDir == 'asc'"
          variant="ghost"
          class="rounded-l-none text-muted-foreground hover:cursor-pointer"
          size="icon"
          title="Ascending"
          @click="sortDir = 'desc'"
        >
          <ChevronUp />
        </Button>
        <Button
          v-else
          variant="ghost"
          class="rounded-l-none text-muted-foreground hover:cursor-pointer"
          size="icon"
          title="Descending"
          @click="sortDir = 'asc'"
        >
          <ChevronDown />
        </Button>
      </div>
    </div>
    <div>
      <Button
        variant="ghost"
        class="text-muted-foreground hover:cursor-pointer"
        title="Shuffle"
        @click="shuffle"
      >
        <Shuffle />
      </Button>
    </div>
  </div>
</template>
