<script lang="ts" setup>
import AppSidebar from "@/components/AppSidebar.vue";
import { Separator } from "@/components/ui/separator";
import Input from "~/components/ui/input/Input.vue";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import {
  Combobox,
  ComboboxAnchor,
  ComboboxEmpty,
  ComboboxGroup,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
  ComboboxSeparator,
} from "@/components/ui/combobox";
import Button from "~/components/ui/button/Button.vue";
import { Search, Sun, Moon } from "lucide-vue-next";

const defaultOpen = useCookie<boolean>("sidebar:state");

//const { currentPage, pageSize } = storeToRefs(usePageStore());
const { searchQuery } = storeToRefs(useSearchStore());

const route = useRoute();

const color = useColorMode();
const changeColorMode = () => {
  if (color.value === "dark") {
    color.preference = "light";
  } else if (color.value === "light") {
    color.preference = "dark";
  }
};

const search = refDebounced(searchQuery, 250);

const compSearch = computed(() => {
  return search;
});

const goSearch = () => {
  if (route.name === "search") {
    navigateTo({
      name: "search",
      query: {
        q: searchQuery.value,
      },
    });
    refreshNuxtData("searchPageResults");
  } else {
    navigateTo({
      name: "search",
      query: {
        q: searchQuery.value,
      },
    });
  }
};

const { data: result } = await useFetch("/api/search", {
  query: {
    q: compSearch.value,
    //page: currentPage.value,
    //size: pageSize.value,
  },
  key: "searchResults",
});
</script>

<template>
  <SidebarProvider :default-open="defaultOpen">
    <AppSidebar />
    <SidebarInset>
      <header class="flex h-16 shrink-0 items-center justify-between">
        <div class="flex items-center gap-2 px-4">
          <SidebarTrigger class="-ml-1" />
          <Separator
            orientation="vertical"
            class="data-[orientation=vertical]:h-4"
          />
        </div>
        <div class="flex flex-1 justify-center items-center mr-8 px-4">
          <div class="relative w-full max-w-xl items-center">
            <Combobox :ignore-filter="true">
              <ComboboxAnchor class="w-full">
                <ComboboxInput v-model="searchQuery" as-child>
                  <Input
                    placeholder="Search..."
                    class="pl-12 rounded-xl"
                    type="text"
                  />
                </ComboboxInput>
                <Button
                  variant="ghost"
                  class="absolute rounded-l-xl start-0 -inset-y-1.5 flex items-center justify-center size-10"
                  @click="goSearch()"
                >
                  <Search class="size-[1.25em] text-muted-foreground" />
                </Button>
              </ComboboxAnchor>
              <ComboboxList
                class="w-(--reka-combobox-trigger-width) max-h-[80dvh]"
              >
                <ComboboxEmpty>Nothing found.</ComboboxEmpty>
                <ComboboxGroup
                  v-if="searchQuery !== ''"
                  class="overflow-y-scroll"
                >
                  <!--TODO: Add covers to search results-->
                  <ComboboxItem :value="searchQuery" @select="goSearch()"
                    >See all results.</ComboboxItem
                  >
                  <ComboboxSeparator v-if="result.archives" class="my-1" />
                  <ComboboxItem
                    v-for="item in result.archives"
                    :key="item.id"
                    :value="item"
                    @select="
                      () => {
                        navigateTo({
                          name: 'a-id',
                          params: { id: item.id },
                        });
                      }
                    "
                  >
                    <NuxtImg
                      :src="`/archive/${item.id}/cover`"
                      width="48px"
                      class="rounded-sm"
                    />
                    {{ item.title }}
                  </ComboboxItem>
                </ComboboxGroup>
              </ComboboxList>
            </Combobox>
          </div>
        </div>
        <div class="pr-2">
          <Button size="icon" variant="ghost" @click="changeColorMode">
            <Sun v-if="color.preference === 'dark'" />
            <Moon v-if="color.preference === 'light'" />
          </Button>
        </div>
      </header>
      <main class="xl:px-40">
        <slot />
      </main>
    </SidebarInset>
  </SidebarProvider>
</template>
