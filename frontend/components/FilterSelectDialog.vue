<script lang="ts" setup>
  import { useFilter } from "reka-ui";
  import { Check, LoaderCircle, Minus, Plus } from "lucide-vue-next";

  const props = defineProps<{
    endpoint: string;
    field: string;
    values: {
      tags?: string[] | undefined;
      artists?: string[] | undefined;
      parodies?: string[] | undefined;
      characters?: string[] | undefined;
      languages?: string[] | undefined;
      categories?: string[] | undefined;
    };
  }>();

  const url = computed(() => {
    return `/api/${props.endpoint}`;
  });

  const { data, status } = await useAsyncData(
    props.field,
    () => $fetch(url.value),
    {}
  );

  const { contains } = useFilter({ sensitivity: "base" });

  const search = ref("");
  const filtered = computed(() => {
    let val: string[] | undefined;
    switch (props.field) {
      case "tags": {
        val = props.values.tags;
        break;
      }
      case "artists": {
        val = props.values.artists;
        break;
      }
      case "parodies": {
        val = props.values.parodies;
        break;
      }
      case "characters": {
        val = props.values.characters;
        break;
      }
      case "languages": {
        val = props.values.languages;
        break;
      }
      case "categories": {
        val = props.values.categories;
        break;
      }
    }

    if (props.field === "languages" || props.field === "categories") {
      const options = data.value.filter((i: string) => !val?.includes(i));
      return search.value
        ? options.filter((option: string) => contains(option, search.value))
        : options;
    } else {
      const options = data.value.filter(
        (i: { name: string }) => !val?.includes(i.name)
      );
      return search.value
        ? options.filter((option: { name: string }) =>
            contains(option.name, search.value)
          )
        : options;
    }
  });

  const hover = ref("");
</script>

<template>
  <div v-if="status === 'pending'" class="flex justify-center">
    <LoaderCircle />
  </div>
  <FormField
    v-if="status === 'success'"
    v-slot="{ componentField }"
    :name="field">
    <FormItem>
      <FormControl>
        <Input v-model="search" placeholder="Search..." />
        <div class="overflow-y-scroll h-(--reka-accordion-content-height)">
          <div class="flex flex-col gap-1 w-[98%]">
            <Button
              v-for="item in componentField.modelValue"
              :key="item"
              class="justify-start bg-chart-2 hover:bg-destructive"
              size="sm"
              @mouseover="hover = item"
              @mouseleave="hover = ''"
              @click="
                () => {
                  const index = componentField.modelValue.indexOf(item);
                  componentField.modelValue.splice(index, 1);
                }
              ">
              <Minus v-if="hover === item" />
              <Check v-else />
              {{ item }}
            </Button>
            <div
              v-if="field === 'languages' || field === 'categories'"
              class="flex flex-col gap-1">
              <Button
                v-for="item in filtered"
                :key="item"
                class="justify-start"
                size="sm"
                variant="secondary"
                @click="componentField.modelValue.push(item)">
                <Plus /> {{ item }}
              </Button>
            </div>
            <div v-else class="flex flex-col gap-1">
              <Button
                v-for="item in filtered"
                :key="item.name"
                class="justify-start"
                size="sm"
                variant="secondary"
                @click="componentField.modelValue.push(item.name)">
                <Plus /> {{ item.name }}
              </Button>
            </div>
          </div>
        </div>
      </FormControl>
      <FormDescription />
      <FormMessage />
    </FormItem>
  </FormField>
</template>
