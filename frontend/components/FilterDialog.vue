<script lang="ts" setup>
  import { z } from "zod/v3";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
    DialogClose,
  } from "@/components/ui/dialog";
  import { ChevronLeft, ChevronRight, ListFilter } from "lucide-vue-next";

  const { filters } = storeToRefs(useFiltersStore());

  const formSchema = toTypedSchema(
    z.object({
      tags: z.array(z.string()).optional(),
      artists: z.array(z.string()).optional(),
      parodies: z.array(z.string()).optional(),
      characters: z.array(z.string()).optional(),
      languages: z.array(z.string()).optional(),
      categories: z.array(z.string()).optional(),
    })
  );

  const { handleSubmit, values, setFieldValue } = useForm({
    validationSchema: formSchema,
    initialValues: {
      tags: filters.value.tags,
      artists: filters.value.artists,
      parodies: filters.value.parodies,
      characters: filters.value.characters,
      languages: filters.value.languages,
      categories: filters.value.categories,
    },
    keepValuesOnUnmount: true,
  });

  const onSubmit = handleSubmit((values) => {
    filters.value.tags = values.tags || [];
    filters.value.artists = values.artists || [];
    filters.value.characters = values.characters || [];
    filters.value.parodies = values.parodies || [];
    filters.value.languages = values.languages || [];
    filters.value.categories = values.categories || [];

    refreshNuxtData("archives");
  });

  const active = ref(0);

  const list = [
    {
      id: 1,
      field: "tags",
      endpoint: "tag",
      len: computed(() => {
        return values.tags?.length;
      }),
    },
    {
      id: 2,
      field: "artists",
      endpoint: "artist",
      len: computed(() => {
        return values.artists?.length;
      }),
    },
    {
      id: 3,
      field: "characters",
      endpoint: "character",
      len: computed(() => {
        return values.characters?.length;
      }),
    },
    {
      id: 4,
      field: "parodies",
      endpoint: "parody",
      len: computed(() => {
        return values.parodies?.length;
      }),
    },
    {
      id: 5,
      field: "languages",
      endpoint: "lang",
      len: computed(() => {
        return values.languages?.length;
      }),
    },
    {
      id: 6,
      field: "categories",
      endpoint: "category",
      len: computed(() => {
        return values.categories?.length;
      }),
    },
  ];

  function clearFilters() {
    setFieldValue("tags", []);
    setFieldValue("artists", []);
    setFieldValue("characters", []);
    setFieldValue("parodies", []);
    setFieldValue("languages", []);
    setFieldValue("categories", []);
  }
</script>

<template>
  <Dialog>
    <DialogTrigger as-child>
      <Button variant="secondary">
        <ListFilter />
      </Button>
    </DialogTrigger>
    <DialogContent class="max-h-[75dvh] flex flex-col justify-start">
      <DialogHeader class="flex flex-row">
        <Button
          v-if="active !== 0"
          variant="ghost"
          size="icon"
          @click="active = 0">
          <ChevronLeft />
        </Button>
        <div v-if="active === 0">
          <DialogTitle>Select Filters</DialogTitle>
          <DialogDescription
            >Select metadata by which to filter archives.</DialogDescription
          >
        </div>
        <div v-else>
          <div v-for="item in list" :key="item.id">
            <div v-if="active === item.id">
              <DialogTitle>Select {{ item.field }}</DialogTitle>
              <DialogDescription
                >Select {{ item.field }} by which to filter
                archives.</DialogDescription
              >
            </div>
          </div>
        </div>
      </DialogHeader>
      <form id="filters" @submit="onSubmit">
        <div class="h-[50dvh] flex flex-col overflow-y-scroll">
          <div v-if="active === 0" class="flex flex-col gap-1">
            <Button
              v-for="item in list"
              :key="item.id"
              variant="secondary"
              class="justify-between"
              @click="active = item.id">
              <div class="flex gap-1">
                <p>{{ item.field }}</p>
                <Badge>{{ item.len }}</Badge>
              </div>
              <ChevronRight />
            </Button>
          </div>
          <div v-for="item in list" :key="item.id">
            <FilterSelectDialog
              v-if="active === item.id"
              :field="item.field"
              :endpoint="item.endpoint"
              :values="values" />
          </div>
        </div>
      </form>
      <DialogClose class="flex gap-1" @click="active = 0">
        <Button
          class="flex-1"
          type="submit"
          form="filters"
          variant="secondary"
          @click="clearFilters()">
          Clear Filters</Button
        >
        <Button class="flex-1" type="submit" form="filters">Apply</Button>
      </DialogClose>
    </DialogContent>
  </Dialog>
</template>
