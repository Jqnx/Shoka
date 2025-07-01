<script lang="ts" setup>
  import AppSidebar from "@/components/AppSidebar.vue";
  import { Separator } from "@/components/ui/separator";
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
  } from "@/components/ui/combobox";
  const defaultOpen = useCookie<boolean>("sidebar:state");
  const { currentPage, pageSize } = storeToRefs(usePageStore());
  const { searchQuery } = storeToRefs(useSearchStore());
  const route = useRoute();

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
      page: currentPage.value,
      size: pageSize.value,
    },
    key: "searchResults",
  });
</script>

<template>
  <SidebarProvider :default-open="defaultOpen">
    <AppSidebar />
    <SidebarInset>
      <header class="flex h-16 shrink-0 items-center gap-2">
        <div class="flex items-center gap-2 px-4">
          <SidebarTrigger class="-ml-1" />
          <Separator
            orientation="vertical"
            class="mr-2 data-[orientation=vertical]:h-4" />
        </div>
        <div class="flex flex-1 justify-center items-center gap-2 px-4">
          <div class="relative w-full max-w-xl items-center">
            <Combobox :ignore-filter="true">
              <ComboboxAnchor class="w-full">
                <ComboboxInput v-model="searchQuery" as-child>
                  <Input
                    placeholder="Search..."
                    class="pl-10 rounded-xl"
                    type="text" />
                </ComboboxInput>
                <span
                  class="absolute start-0 inset-y-0 flex items-center justify-center px-2">
                  <Icon
                    name="lucide:search"
                    size="1.25em"
                    class="text-muted-foreground" />
                </span>
              </ComboboxAnchor>
              <ComboboxList
                class="w-(--reka-combobox-trigger-width) max-h-[80dvh]">
                <ComboboxEmpty>Nothing found.</ComboboxEmpty>
                <ComboboxGroup
                  v-if="searchQuery !== ''"
                  class="overflow-y-scroll">
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
                          params: { id: item.archive_id },
                        });
                      }
                    ">
                    {{ item.title }}
                  </ComboboxItem>
                </ComboboxGroup>
              </ComboboxList>
            </Combobox>
          </div>
        </div>
      </header>
      <slot />
    </SidebarInset>
  </SidebarProvider>
</template>
