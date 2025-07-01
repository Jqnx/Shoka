<script lang="ts" setup>
  import { ChevronDown, ChevronUp, Shuffle } from "lucide-vue-next";

  const { sortBy, sortDir, sortLabel } = storeToRefs(useFiltersStore());

  const sortList = [
    { value: "title", label: "Title" },
    { value: "page_count", label: "Page Count" },
    { value: "created_at", label: "Created At" },
    { value: "updated_at", label: "Updated At" },
    { value: "release_date", label: "Release Date" },
  ];
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
              {{ sortLabel }}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuItem
              v-for="(item, index) in sortList"
              :key="index"
              @select="
                () => {
                  sortBy = item.value;
                  sortLabel = item.label;
                }
              ">
              {{ item.label }}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <Button
          v-if="sortDir == 'asc'"
          variant="secondary"
          class="rounded-l-none"
          size="icon"
          @click="sortDir = 'desc'">
          <ChevronUp />
        </Button>
        <Button
          v-else
          variant="secondary"
          class="rounded-l-none"
          size="icon"
          @click="sortDir = 'asc'">
          <ChevronDown />
        </Button>
      </div>
    </div>
    <div>
      <Button variant="secondary">
        <!--TODO: Shuffle functionality-->
        <Shuffle />
      </Button>
    </div>
  </div>
</template>
