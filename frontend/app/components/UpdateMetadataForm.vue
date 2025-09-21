<script setup lang="ts">
import { z } from "zod";
import {
  DateFormatter,
  getLocalTimeZone,
  parseAbsolute,
  parseZonedDateTime,
  today,
  ZonedDateTime,
  type DateValue,
} from "@internationalized/date";
import { toDate } from "reka-ui/date";
import { cn } from "~/lib/utils";
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormMessage,
} from "@/components/ui/form";
import {
  TagsInput,
  TagsInputInput,
  TagsInputItem,
  TagsInputItemDelete,
  TagsInputItemText,
} from "@/components/ui/tags-input";
import {
  Combobox,
  ComboboxAnchor,
  ComboboxEmpty,
  ComboboxGroup,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { Textarea } from "@/components/ui/textarea";
import { Calendar } from "@/components/ui/calendar";
import { CalendarIcon, XIcon, CheckIcon } from "lucide-vue-next";
import { useFilter } from "reka-ui";
import ComboboxViewport from "./ui/combobox/ComboboxViewport.vue";
import Button from "./ui/button/Button.vue";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import Input from "./ui/input/Input.vue";

const props = defineProps(["new"]);
const meta = props.new[0];

const emit = defineEmits(["save", "close"]);

const { data: archive } = useNuxtData("archive");

const df = new DateFormatter("en-GB", {
  dateStyle: "long",
});

function checkString(a: string) {
  if (!a) {
    return "";
  } else {
    return a;
  }
}

function arraysEqual(
  arr1: string[] | undefined,
  arr2: string[] | undefined,
): boolean {
  if (arr1 === undefined || arr2 === undefined) return false;
  if (arr1?.length !== arr2?.length) return false;
  const sortedArr1 = arr1?.sort() as string[];
  const sortedArr2 = arr2?.sort() as string[];
  return sortedArr1.every((val, index) => val === sortedArr2[index]);
}

const oldArtists = toArray(archive.value.artist);
const oldTags = toArray(archive.value.tags);
const oldParodies = toArray(archive.value.parody);
const oldCharacters = toArray(archive.value.character);
const oldUrls = toArray(archive.value.url);
const oldReleaseDate = computed(() => {
  return archive.value.release_date
    ? parseAbsolute(archive.value.release_date, "UTC")
    : undefined;
});

function checkDate(updated: DateValue, old: DateValue | undefined): boolean {
  if (old === undefined) {
    return true;
  }
  return df.format(toDate(updated)) !== df.format(toDate(old));
}

const newArtists = toArray(meta.artists);
const newTags = toArray(meta.tags);
const newParodies = toArray(meta.parodies);
const newCharacters = toArray(meta.characters);
const newUrls = toArray(meta.urls);

const formSchema = toTypedSchema(
  z.object({
    title: z.string().optional(),
    summary: z.string().optional().nullable(),
    tags: z.array(z.string()).optional(),
    artist: z.array(z.string()).optional(),
    parody: z.array(z.string()).optional(),
    character: z.array(z.string()).optional(),
    language: z.string().optional().nullable(),
    category: z.string().optional().nullable(),
    url: z.array(z.string().url()).optional(),
    release_date: z.string().optional(),
  }),
);

const { handleSubmit, setFieldValue, values } = useForm({
  validationSchema: formSchema,
  initialValues: {
    title: meta.title,
    summary: meta.summary,
    tags: newTags,
    artist: newArtists,
    parody: newParodies,
    character: newCharacters,
    language: meta.language,
    category: meta.category,
    url: newUrls,
    release_date: meta.release_date,
  },
});

const value = computed(() => {
  return parseAbsolute(values.release_date as string, "UTC");
});

const placeholder = ref();
const { contains } = useFilter({ sensitivity: "base" });

const { data: allArtists } = await useLazyFetch("/api/artist");
const openArtist = ref(false);
const searchArtist = ref("");
const filteredArtists = computed(() => {
  const options = allArtists.value.filter(
    (i: { name: string }) => !values.artist?.includes(i.name),
  );
  return searchArtist.value
    ? options.filter((option: { name: string }) =>
        contains(option.name, searchArtist.value),
      )
    : options;
});

const { data: allTags } = await useLazyFetch("/api/tag");
const openTag = ref(false);
const searchTag = ref("");
const filteredTags = computed(() => {
  const options = allTags.value.filter(
    (i: { name: string }) => !values.tags?.includes(i.name),
  );
  return searchTag.value
    ? options.filter((option: { name: string }) =>
        contains(option.name, searchTag.value),
      )
    : options;
});

const { data: allParodies } = await useLazyFetch("/api/parody");
const openParody = ref(false);
const searchParody = ref("");
const filteredParodies = computed(() => {
  const options = allParodies.value.filter(
    (i: { name: string }) => !values.parody?.includes(i.name),
  );
  return searchParody.value
    ? options.filter((option: { name: string }) =>
        contains(option.name, searchParody.value),
      )
    : options;
});

const { data: allCharacters } = await useLazyFetch("/api/character");
const openCharacter = ref(false);
const searchCharacter = ref("");
const filteredCharacters = computed(() => {
  const options = allCharacters.value.filter(
    (i: { name: string }) => !values.character?.includes(i.name),
  );
  return searchCharacter.value
    ? options.filter((option: { name: string }) =>
        contains(option.name, searchCharacter.value),
      )
    : options;
});

const { data: allLanguages } = await useLazyFetch("/api/lang");
const openLanguage = ref(false);
const searchLanguage = ref("");
const filteredLanguages = computed(() => {
  const options = allLanguages.value.filter(
    (i: string) => !values.language?.includes(i),
  );
  return searchLanguage.value
    ? options.filter((option: string) => contains(option, searchLanguage.value))
    : options;
});

const { data: allCategories } = await useLazyFetch("/api/category");
const openCategory = ref(false);
const searchCategory = ref("");
const filteredCategories = computed(() => {
  const options = allCategories.value.filter(
    (i: string) => !values.category?.includes(i),
  );
  return searchCategory.value
    ? options.filter((option: string) => contains(option, searchCategory.value))
    : options;
});

const activeTitle = ref(false);
const activeSummary = ref(false);
const activeArtists = ref(false);
const activeTags = ref(false);
const activeParodies = ref(false);
const activeCharacters = ref(false);
const activeLanguage = ref(false);
const activeCategory = ref(false);
const activeReleaseDate = ref(false);
const activeURLs = ref(false);

console.log(oldArtists);

//const id = useRoute().params.id;
const onSubmit = handleSubmit((values) => {
  emit("save", values);
  emit("close");
});
</script>

<template>
  <div id="main" class="w-full overflow-y-auto">
    <!-- Updateable: Title, Summary, Tags, Artists, Parodies, Characters, Languages, Categories, Release Date, URLs -->
    <form @submit="onSubmit">
      <!-- Title -->
      <div
        v-if="values.title && values.title !== archive.title"
        class="grid grid-cols-7"
      >
        <Label>Title</Label>
        <div class="flex items-center justify-center gap-2 col-span-6">
          <div class="flex flex-1 pb-2">
            <Button
              variant="outline"
              class="rounded-r-none h-auto"
              :disabled="checkString(archive.title) === ''"
              type="button"
              @click="
                () => {
                  if (!activeTitle) {
                    activeTitle = true;
                    setFieldValue('title', checkString(archive.title));
                  }
                }
              "
            >
              <CheckIcon v-if="activeTitle" class="stroke-success" />
              <XIcon v-else class="stroke-destructive" />
            </Button>
            <Input
              type="text"
              :model-value="checkString(archive.title)"
              class="disabled:opacity-80 text-muted-foreground"
              disabled
            />
          </div>
          <FormField v-slot="{ componentField }" name="title">
            <FormItem class="flex-1">
              <FormControl>
                <Button
                  variant="outline"
                  class="rounded-r-none h-auto"
                  type="button"
                  @click="
                    () => {
                      if (activeTitle) {
                        activeTitle = false;
                        setFieldValue('title', componentField.modelValue);
                      }
                    }
                  "
                >
                  <CheckIcon v-if="!activeTitle" class="stroke-success" />
                  <XIcon v-else class="stroke-destructive" />
                </Button>
                <Input type="text" v-bind="componentField" />
              </FormControl>
              <FormDescription />
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
      </div>

      <!-- Summary -->
      <div
        v-if="values.summary && values.summary !== archive.summary"
        class="grid grid-cols-7"
      >
        <Label>Summary</Label>
        <div class="flex items-center justify-center gap-2 col-span-6">
          <div class="flex flex-1 pb-2">
            <Button
              variant="outline"
              class="rounded-r-none h-auto"
              :disabled="checkString(archive.summary) === ''"
              type="button"
              @click="
                () => {
                  if (!activeSummary) {
                    activeSummary = true;
                    setFieldValue('summary', checkString(archive.summary));
                  }
                }
              "
            >
              <CheckIcon v-if="activeSummary" class="stroke-success" />
              <XIcon v-else class="stroke-destructive" />
            </Button>
            <Textarea
              type="text"
              class="resize-none disabled:opacity-80 text-muted-foreground rounded-l-none"
              :model-value="checkString(archive.summary)"
              disabled
            />
          </div>
          <FormField v-slot="{ componentField }" name="summary">
            <FormItem class="flex flex-1 gap-0 pb-2">
              <FormControl>
                <Button
                  variant="outline"
                  class="rounded-r-none h-auto"
                  type="button"
                  @click="
                    () => {
                      if (activeSummary) {
                        activeSummary = false;
                        setFieldValue('summary', componentField.modelValue);
                      }
                    }
                  "
                >
                  <CheckIcon v-if="!activeSummary" class="stroke-success" />
                  <XIcon v-else class="stroke-destructive" />
                </Button>
                <Textarea
                  type="text"
                  class="rounded-l-none"
                  v-bind="componentField"
                />
              </FormControl>
              <FormDescription />
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
      </div>

      <!-- Artists -->
      <div
        v-if="values.artist && !arraysEqual(newArtists, oldArtists)"
        class="grid grid-cols-7"
      >
        <Label>Artist</Label>
        <div class="flex items-center justify-center gap-2 col-span-6">
          <div class="flex flex-1 pb-2">
            <Button
              variant="outline"
              class="rounded-r-none h-auto"
              :disabled="oldArtists === undefined"
              type="button"
              @click="
                () => {
                  if (!activeArtists) {
                    activeArtists = true;
                    setFieldValue('artist', oldArtists);
                  }
                }
              "
            >
              <CheckIcon v-if="activeArtists" class="stroke-success" />
              <XIcon v-else class="stroke-destructive" />
            </Button>
            <TagsInput
              :model-value="oldArtists"
              class="w-full bg-input/30 opacity-80 min-h-10 text-muted-foreground rounded-l-none"
              disabled
            >
              <TagsInputItem
                v-for="item in oldArtists"
                :key="item"
                :value="item"
              >
                <TagsInputItemText />
                <TagsInputItemDelete />
              </TagsInputItem>
            </TagsInput>
          </div>

          <FormField v-slot="{ componentField }" name="artist">
            <FormItem class="flex flex-1 gap-0 pb-2">
              <Button
                variant="outline"
                class="rounded-r-none h-auto"
                type="button"
                @click="
                  () => {
                    if (activeArtists) {
                      activeArtists = false;
                      setFieldValue('artist', componentField.modelValue);
                    }
                  }
                "
              >
                <CheckIcon v-if="!activeArtists" class="stroke-success" />
                <XIcon v-else class="stroke-destructive" />
              </Button>
              <Combobox
                v-model="componentField.modelValue"
                v-model:open="openArtist"
                :ignore-filter="true"
              >
                <FormControl>
                  <ComboboxAnchor as-child>
                    <TagsInput
                      :model-value="componentField.modelValue"
                      class="bg-input/30 w-full rounded-l-none"
                      @update:model-value="
                        componentField['onUpdate:modelValue']
                      "
                    >
                      <TagsInputItem
                        v-for="item in componentField.modelValue"
                        :key="item"
                        :value="item"
                      >
                        <TagsInputItemText />
                        <TagsInputItemDelete />
                      </TagsInputItem>

                      <ComboboxInput v-model="searchArtist" as-child>
                        <TagsInputInput @keydown.enter="searchArtist = ''" />
                      </ComboboxInput>
                    </TagsInput>

                    <ComboboxList
                      :collision-padding="4"
                      :avoid-collisions="false"
                      class="max-h-(--reka-combobox-content-available-height) w-(--reka-combobox-trigger-width)"
                    >
                      <ComboboxViewport class="max-h-[40vh]">
                        <ComboboxEmpty />
                        <ComboboxGroup>
                          <ComboboxItem
                            v-for="artist in filteredArtists"
                            :key="artist.id"
                            :value="artist.name"
                            @select.prevent="
                              (ev) => {
                                if (typeof ev.detail.value === 'string') {
                                  searchArtist = '';
                                  componentField.modelValue.push(
                                    ev.detail.value,
                                  );
                                }

                                //if (artists.length === 0) {
                                //  openArtist = false;
                                //}
                              }
                            "
                          >
                            {{ artist.name }}
                          </ComboboxItem>
                        </ComboboxGroup>
                      </ComboboxViewport>
                    </ComboboxList>
                  </ComboboxAnchor>
                </FormControl>
              </Combobox>
              <FormDescription />
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
      </div>

      <!-- Tags -->
      <div
        v-if="values.tags && !arraysEqual(newTags, oldTags)"
        class="grid grid-cols-7"
      >
        <Label>Tags</Label>
        <div class="flex items-center justify-center gap-2 col-span-6">
          <div class="flex flex-1 pb-2">
            <Button
              variant="outline"
              class="rounded-r-none h-auto"
              :disabled="oldTags === undefined"
              type="button"
              @click="
                () => {
                  if (!activeTags) {
                    activeTags = true;
                    setFieldValue('tags', oldTags);
                  }
                }
              "
            >
              <CheckIcon v-if="activeTags" class="stroke-success" />
              <XIcon v-else class="stroke-destructive" />
            </Button>
            <TagsInput
              :model-value="oldTags"
              class="w-full bg-input/30 opacity-80 min-h-10 rounded-l-none"
              disabled
            >
              <TagsInputItem
                v-for="item in oldTags"
                :key="item"
                :value="item"
                class="text-muted-foreground"
              >
                <TagsInputItemText />
                <TagsInputItemDelete />
              </TagsInputItem>
            </TagsInput>
          </div>

          <FormField v-slot="{ componentField }" name="tags">
            <FormItem class="flex flex-1 pb-2 gap-0">
              <Button
                variant="outline"
                class="rounded-r-none h-auto"
                type="button"
                @click="
                  () => {
                    if (activeTags) {
                      activeTags = false;
                      setFieldValue('tags', componentField.modelValue);
                    }
                  }
                "
              >
                <CheckIcon v-if="!activeTags" class="stroke-success" />
                <XIcon v-else class="stroke-destructive" />
              </Button>
              <Combobox
                v-model="componentField.modelValue"
                v-model:open="openTag"
                :ignore-filter="true"
              >
                <FormControl>
                  <ComboboxAnchor as-child>
                    <TagsInput
                      :model-value="componentField.modelValue"
                      class="bg-input/30 w-full rounded-l-none"
                      @update:model-value="
                        componentField['onUpdate:modelValue']
                      "
                    >
                      <TagsInputItem
                        v-for="item in componentField.modelValue"
                        :key="item"
                        :value="item"
                      >
                        <TagsInputItemText />
                        <TagsInputItemDelete />
                      </TagsInputItem>

                      <ComboboxInput v-model="searchTag" as-child>
                        <TagsInputInput @keydown.enter="searchTag = ''" />
                      </ComboboxInput>
                    </TagsInput>

                    <ComboboxList
                      :collision-padding="4"
                      :avoid-collisions="false"
                      class="max-h-(--reka-combobox-content-available-height) w-(--reka-combobox-trigger-width)"
                    >
                      <ComboboxViewport class="max-h-[40vh]">
                        <ComboboxEmpty />
                        <ComboboxGroup>
                          <ComboboxItem
                            v-for="tag in filteredTags"
                            :key="tag.id"
                            :value="tag.name"
                            @select.prevent="
                              (ev) => {
                                if (typeof ev.detail.value === 'string') {
                                  searchTag = '';
                                  componentField.modelValue.push(
                                    ev.detail.value,
                                  );
                                }

                                if (allTags.length === 0) {
                                  openTag = false;
                                }
                              }
                            "
                          >
                            {{ tag.name }}
                          </ComboboxItem>
                        </ComboboxGroup>
                      </ComboboxViewport>
                    </ComboboxList>
                  </ComboboxAnchor>
                </FormControl>
              </Combobox>
              <FormDescription />
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
      </div>

      <!-- Parodies -->
      <div
        v-if="values.parody && !arraysEqual(newParodies, oldParodies)"
        class="grid grid-cols-7"
      >
        <Label>Parodies</Label>
        <div class="flex items-center justify-center gap-2 col-span-6">
          <div class="flex flex-1 pb-2">
            <Button
              variant="outline"
              class="rounded-r-none h-auto"
              :disabled="oldParodies === undefined"
              type="button"
              @click="
                () => {
                  if (!activeParodies) {
                    activeParodies = true;
                    setFieldValue('parody', oldParodies);
                  }
                }
              "
            >
              <CheckIcon v-if="activeParodies" class="stroke-success" />
              <XIcon v-else class="stroke-destructive" />
            </Button>
            <TagsInput
              :model-value="oldParodies"
              class="w-full bg-input/30 opacity-80 min-h-10 rounded-l-none"
            >
              <TagsInputItem
                v-for="item in oldParodies"
                :key="item"
                :value="item"
                class="text-muted-foreground"
              >
                <TagsInputItemText />
                <TagsInputItemDelete />
              </TagsInputItem>
            </TagsInput>
          </div>

          <FormField v-slot="{ componentField }" name="parody">
            <FormItem class="flex flex-1 gap-0 pb-2">
              <Button
                variant="outline"
                class="rounded-r-none h-auto"
                type="button"
                @click="
                  () => {
                    if (activeParodies) {
                      activeParodies = false;
                      setFieldValue('parody', componentField.modelValue);
                    }
                  }
                "
              >
                <CheckIcon v-if="!activeParodies" class="stroke-success" />
                <XIcon v-else class="stroke-destructive" />
              </Button>
              <Combobox
                v-model="componentField.modelValue"
                v-model:open="openParody"
                :ignore-filter="true"
              >
                <FormControl>
                  <ComboboxAnchor as-child>
                    <TagsInput
                      :model-value="componentField.modelValue"
                      class="bg-input/30 w-full rounded-l-none"
                      @update:model-value="
                        componentField['onUpdate:modelValue']
                      "
                    >
                      <TagsInputItem
                        v-for="item in componentField.modelValue"
                        :key="item"
                        :value="item"
                      >
                        <TagsInputItemText />
                        <TagsInputItemDelete />
                      </TagsInputItem>

                      <ComboboxInput v-model="searchParody" as-child>
                        <TagsInputInput @keydown.enter="searchParody = ''" />
                      </ComboboxInput>
                    </TagsInput>

                    <ComboboxList
                      :collision-padding="4"
                      :avoid-collisions="false"
                      class="max-h-(--reka-combobox-content-available-height) w-(--reka-combobox-trigger-width)"
                    >
                      <ComboboxViewport class="max-h-[40vh]">
                        <ComboboxEmpty />
                        <ComboboxGroup>
                          <ComboboxItem
                            v-for="parody in filteredParodies"
                            :key="parody.id"
                            :value="parody.name"
                            @select.prevent="
                              (ev) => {
                                if (typeof ev.detail.value === 'string') {
                                  searchParody = '';
                                  componentField.modelValue.push(
                                    ev.detail.value,
                                  );
                                }

                                //if (parodies.length === 0) {
                                //  openParody = false;
                                //}
                              }
                            "
                          >
                            {{ parody.name }}
                          </ComboboxItem>
                        </ComboboxGroup>
                      </ComboboxViewport>
                    </ComboboxList>
                  </ComboboxAnchor>
                </FormControl>
              </Combobox>
              <FormDescription />
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
      </div>

      <!-- Characters -->
      <div
        v-if="values.character && !arraysEqual(newCharacters, oldCharacters)"
        class="grid grid-cols-7"
      >
        <Label>Characters</Label>
        <div class="flex items-center justify-center gap-2 col-span-6">
          <div class="flex flex-1 pb-2">
            <Button
              variant="outline"
              class="rounded-r-none h-auto"
              :disabled="oldCharacters === undefined"
              type="button"
              @click="
                () => {
                  if (!activeCharacters) {
                    activeCharacters = true;
                    setFieldValue('character', oldCharacters);
                  }
                }
              "
            >
              <CheckIcon v-if="activeCharacters" class="stroke-success" />
              <XIcon v-else class="stroke-destructive" />
            </Button>
            <TagsInput
              :model-value="oldCharacters"
              class="w-full bg-input/30 opacity-80 min-h-10 rounded-l-none"
            >
              <TagsInputItem
                v-for="item in oldCharacters"
                :key="item"
                :value="item"
                class="text-muted-foreground"
              >
                <TagsInputItemText />
                <TagsInputItemDelete />
              </TagsInputItem>
            </TagsInput>
          </div>

          <FormField v-slot="{ componentField }" name="character">
            <FormItem class="flex flex-1 gap-0 pb-2">
              <Button
                variant="outline"
                class="rounded-r-none h-auto"
                type="button"
                @click="
                  () => {
                    if (activeCharacters) {
                      activeCharacters = false;
                      setFieldValue('character', componentField.modelValue);
                    }
                  }
                "
              >
                <CheckIcon v-if="!activeCharacters" class="stroke-success" />
                <XIcon v-else class="stroke-destructive" />
              </Button>
              <Combobox
                v-model="componentField.modelValue"
                v-model:open="openCharacter"
                :ignore-filter="true"
              >
                <FormControl>
                  <ComboboxAnchor as-child>
                    <TagsInput
                      :model-value="componentField.modelValue"
                      class="bg-input/30 w-full rounded-l-none"
                      @update:model-value="
                        componentField['onUpdate:modelValue']
                      "
                    >
                      <TagsInputItem
                        v-for="item in componentField.modelValue"
                        :key="item"
                        :value="item"
                      >
                        <TagsInputItemText />
                        <TagsInputItemDelete />
                      </TagsInputItem>

                      <ComboboxInput v-model="searchCharacter" as-child>
                        <TagsInputInput @keydown.enter="searchCharacter = ''" />
                      </ComboboxInput>
                    </TagsInput>

                    <ComboboxList
                      :collision-padding="4"
                      :avoid-collisions="false"
                      class="max-h-(--reka-combobox-content-available-height) w-(--reka-combobox-trigger-width)"
                    >
                      <ComboboxViewport class="max-h-[40vh]">
                        <ComboboxEmpty />
                        <ComboboxGroup>
                          <ComboboxItem
                            v-for="character in filteredCharacters"
                            :key="character.id"
                            :value="character.name"
                            @select.prevent="
                              (ev) => {
                                if (typeof ev.detail.value === 'string') {
                                  searchCharacter = '';
                                  componentField.modelValue.push(
                                    ev.detail.value,
                                  );
                                }

                                if (allCharacters.length === 0) {
                                  openCharacter = false;
                                }
                              }
                            "
                          >
                            {{ character.name }}
                          </ComboboxItem>
                        </ComboboxGroup>
                      </ComboboxViewport>
                    </ComboboxList>
                  </ComboboxAnchor>
                </FormControl>
              </Combobox>
              <FormDescription />
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
      </div>

      <!-- Language -->
      <div
        v-if="values.language && values.language !== archive.language"
        class="grid grid-cols-7"
      >
        <Label>Language</Label>
        <div class="flex items-center justify-center gap-2 col-span-6">
          <div class="flex flex-1 pb-2">
            <Button
              variant="outline"
              class="rounded-r-none h-auto"
              :disabled="checkString(archive.language) === ''"
              type="button"
              @click="
                () => {
                  if (!activeLanguage) {
                    activeLanguage = true;
                    setFieldValue('language', checkString(archive.language));
                  }
                }
              "
            >
              <CheckIcon v-if="activeLanguage" class="stroke-success" />
              <XIcon v-else class="stroke-destructive" />
            </Button>
            <Input
              class="w-full text-muted-foreground disabled:opacity-80 h-10 rounded-l-none"
              type="text"
              :model-value="checkString(archive.language)"
              disabled
            />
          </div>
          <FormField v-slot="{ componentField }" name="language">
            <FormItem class="flex flex-1 gap-0 pb-2 items-center">
              <Button
                variant="outline"
                class="rounded-r-none min-h-10 h-auto"
                type="button"
                @click="
                  () => {
                    if (activeLanguage) {
                      activeLanguage = false;
                      setFieldValue('language', componentField.modelValue);
                    }
                  }
                "
              >
                <CheckIcon v-if="!activeLanguage" class="stroke-success" />
                <XIcon v-else class="stroke-destructive" />
              </Button>
              <Combobox
                v-model:model-value="componentField.modelValue"
                v-model:open="openLanguage"
                :ignore-filter="true"
              >
                <FormControl>
                  <ComboboxAnchor as-child>
                    <ComboboxInput v-model="searchLanguage" as-child>
                      <Input
                        class="w-full rounded-l-none"
                        type="text"
                        v-bind="componentField"
                      />
                    </ComboboxInput>
                  </ComboboxAnchor>
                </FormControl>

                <ComboboxList
                  :collision-padding="4"
                  :avoid-collisions="false"
                  class="max-h-(--reka-combobox-content-available-height) w-(--reka-combobox-trigger-width)"
                >
                  <ComboboxEmpty />

                  <ComboboxGroup>
                    <ComboboxItem
                      v-for="(language, index) in filteredLanguages"
                      :key="index"
                      :value="language"
                      @select="
                        () => {
                          setFieldValue('language', language);
                        }
                      "
                    >
                      {{ language }}
                    </ComboboxItem>
                  </ComboboxGroup>
                </ComboboxList>
              </Combobox>
              <FormDescription />
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
      </div>

      <!-- Category -->
      <div
        v-if="values.category && values.category !== archive.category"
        class="grid grid-cols-7"
      >
        <Label>Category</Label>
        <div class="flex items-center justify-center gap-2 col-span-6">
          <div class="flex flex-1 pb-2">
            <Button
              variant="outline"
              class="rounded-r-none h-auto"
              :disabled="checkString(archive.category) === ''"
              type="button"
              @click="
                () => {
                  if (!activeCategory) {
                    activeCategory = true;
                    setFieldValue('category', checkString(archive.category));
                  }
                }
              "
            >
              <CheckIcon v-if="activeCategory" class="stroke-success" />
              <XIcon v-else class="stroke-destructive" />
            </Button>
            <Input
              class="w-full text-muted-foreground disabled:opacity-80 h-10 rounded-l-none"
              type="text"
              :model-value="checkString(archive.category)"
              disabled
            />
          </div>
          <FormField v-slot="{ componentField }" name="category">
            <FormItem class="flex flex-1 gap-0 pb-2 items-center">
              <Button
                variant="outline"
                class="rounded-r-none min-h-10 h-auto"
                type="button"
                @click="
                  () => {
                    if (activeCategory) {
                      activeCategory = false;
                      setFieldValue('category', componentField.modelValue);
                    }
                  }
                "
              >
                <CheckIcon v-if="!activeCategory" class="stroke-success" />
                <XIcon v-else class="stroke-destructive" />
              </Button>
              <Combobox
                v-model:model-value="componentField.modelValue"
                v-model:open="openCategory"
                :ignore-filter="true"
              >
                <FormControl>
                  <ComboboxAnchor as-child>
                    <ComboboxInput v-model="searchCategory" as-child>
                      <Input
                        class="w-full rounded-l-none"
                        type="text"
                        v-bind="componentField"
                      />
                    </ComboboxInput>
                  </ComboboxAnchor>
                </FormControl>

                <ComboboxList
                  :collision-padding="4"
                  :avoid-collisions="false"
                  class="max-h-(--reka-combobox-content-available-height) w-(--reka-combobox-trigger-width)"
                >
                  <ComboboxEmpty />

                  <ComboboxGroup>
                    <ComboboxItem
                      v-for="(category, index) in filteredCategories"
                      :key="index"
                      :value="category"
                      @select="
                        () => {
                          setFieldValue('category', category);
                        }
                      "
                    >
                      {{ category }}
                    </ComboboxItem>
                  </ComboboxGroup>
                </ComboboxList>
              </Combobox>
              <FormDescription />
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
      </div>

      <!-- Release Date -->
      <div
        v-if="values.release_date && checkDate(value, oldReleaseDate)"
        class="grid grid-cols-7"
      >
        <Label>Release Date</Label>
        <div class="flex items-center justify-center gap-2 col-span-6">
          <div class="flex flex-1 pb-2">
            <Button
              variant="outline"
              class="rounded-r-none h-auto"
              :disabled="oldReleaseDate === undefined"
              type="button"
              @click="
                () => {
                  if (!activeReleaseDate) {
                    activeReleaseDate = true;
                    setFieldValue(
                      'release_date',
                      oldReleaseDate?.toAbsoluteString(),
                    );
                  }
                }
              "
            >
              <CheckIcon v-if="activeReleaseDate" class="stroke-success" />
              <XIcon v-else class="stroke-destructive" />
            </Button>
            <Button
              variant="outline"
              :class="
                cn(
                  'w-[86%] ps-3 text-start font-normal rounded-l-none min-h-10',
                  !value && 'text-muted-foreground disabled:opacity-80 ',
                )
              "
              disabled
            >
              <span>{{
                archive.release_date
                  ? df.format(toDate(oldReleaseDate as DateValue))
                  : ""
              }}</span>
              <CalendarIcon class="ms-auto h-4 w-4 opacity-50" />
            </Button>
          </div>
          <FormField name="release_date">
            <FormItem class="flex flex-1 gap-0 pb-2">
              <Button
                variant="outline"
                class="rounded-r-none h-auto"
                type="button"
                @click="
                  () => {
                    if (activeReleaseDate) {
                      activeReleaseDate = false;
                      setFieldValue('release_date', value.toAbsoluteString());
                    }
                  }
                "
              >
                <CheckIcon v-if="!activeReleaseDate" class="stroke-success" />
                <XIcon v-else class="stroke-destructive" />
              </Button>
              <div class="flex flex-col flex-1">
                <Popover>
                  <PopoverTrigger as-child>
                    <FormControl>
                      <Button
                        variant="outline"
                        :class="
                          cn(
                            'w-full ps-3 text-start font-normal rounded-l-none min-h-10',
                            !value && 'text-muted-foreground',
                          )
                        "
                      >
                        <span>{{
                          value ? df.format(toDate(value)) : "Pick a date"
                        }}</span>
                        <CalendarIcon class="ms-auto h-4 w-4 opacity-50" />
                      </Button>
                      <input hidden />
                    </FormControl>
                  </PopoverTrigger>
                  <PopoverContent class="w-auto p-0">
                    <Calendar
                      v-model:placeholder="placeholder"
                      :model-value="value"
                      calendar-label="Release Date"
                      initial-focus
                      :min-value="new ZonedDateTime(1900, 1, 1, 'UTC', 0)"
                      :max-value="today(getLocalTimeZone())"
                      @update:model-value="
                        (v) => {
                          if (v) {
                            setFieldValue(
                              'release_date',
                              parseZonedDateTime(
                                v.toString(),
                              ).toAbsoluteString(),
                            );
                          } else {
                            setFieldValue('release_date', undefined);
                          }
                        }
                      "
                    />
                  </PopoverContent>
                </Popover>
              </div>
              <FormDescription />
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
      </div>

      <!-- URLs -->
      <div
        v-if="values.url && !arraysEqual(newUrls, oldUrls)"
        class="grid grid-cols-7"
      >
        <Label>URLs</Label>
        <div class="flex items-center justify-center gap-2 col-span-6">
          <div class="flex flex-1 pb-2">
            <Button
              variant="outline"
              class="rounded-r-none h-auto"
              :disabled="oldUrls === undefined"
              type="button"
              @click="
                () => {
                  if (!activeURLs) {
                    activeURLs = true;
                    setFieldValue('url', oldUrls);
                  }
                }
              "
            >
              <CheckIcon v-if="activeURLs" class="stroke-success" />
              <XIcon v-else class="stroke-destructive" />
            </Button>
            <TagsInput
              class="bg-input/30 opacity-80 w-full min-h-10 rounded-l-none"
              disabled
              :model-value="oldUrls"
            >
              <TagsInputItem
                v-for="item in oldUrls"
                :key="item"
                :value="item"
                class="text-muted-foreground"
              >
                <TagsInputItemText />
                <TagsInputItemDelete />
              </TagsInputItem>
            </TagsInput>
          </div>
          <FormField v-slot="{ componentField }" name="url">
            <FormItem class="flex flex-1 gap-0 pb-2">
              <Button
                variant="outline"
                class="rounded-r-none h-auto"
                type="button"
                @click="
                  () => {
                    if (activeURLs) {
                      activeURLs = false;
                      setFieldValue('url', componentField.modelValue);
                    }
                  }
                "
              >
                <CheckIcon v-if="!activeURLs" class="stroke-success" />
                <XIcon v-else class="stroke-destructive" />
              </Button>
              <FormControl>
                <TagsInput
                  :model-value="componentField.modelValue"
                  class="h-10 bg-input/30 rounded-l-none"
                  @update:model-value="componentField['onUpdate:modelValue']"
                >
                  <TagsInputItem
                    v-for="item in componentField.modelValue"
                    :key="item"
                    :value="item"
                  >
                    <TagsInputItemText />
                    <TagsInputItemDelete />
                  </TagsInputItem>
                  <TagsInputInput />
                </TagsInput>
              </FormControl>
              <FormDescription />
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
      </div>

      <div class="flex gap-2 flex-1">
        <Button
          variant="secondary"
          class="w-full flex-1"
          type="button"
          @click="$emit('close')"
          >Close</Button
        >
        <Button class="w-full flex-1" type="submit">Save</Button>
      </div>
    </form>
  </div>
</template>
