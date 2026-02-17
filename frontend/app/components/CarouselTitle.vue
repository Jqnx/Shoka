<script lang="ts" setup>
import { ArrowRight } from "lucide-vue-next";
const { sortBy, sortLabel, page, pageSize, sortDir } =
  storeToRefs(useFiltersStore());
const router = useRouter();
const props = defineProps<{
  title: string;
  to: string;
  sort?: SortOption;
  sortdir?: string;
}>();

const nav = () => {
  sortBy.value = <string>props.sort?.value;
  sortLabel.value = <string>props.sort?.label;
  sortDir.value = <string>props.sortdir;
  router.push({
    name: "a",
    query: {
      page: page.value,
      size: pageSize.value,
      sortby: sortBy.value,
      sortdir: sortDir.value,
    },
  });
};
</script>

<template>
  <div class="flex flex-1 justify-between px-2">
    <p
      class="font-semibold text-xl tracking-tight cursor-pointer"
      @click="nav()"
    >
      {{ title }}
    </p>
    <p @click="nav()">
      <ArrowRight class="size-5 cursor-pointer" />
    </p>
  </div>
</template>
