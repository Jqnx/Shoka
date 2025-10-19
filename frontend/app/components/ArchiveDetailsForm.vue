<script setup lang="ts">
import { z } from "zod";
import {
  DateFormatter,
  getLocalTimeZone,
  parseAbsolute,
  parseZonedDateTime,
  today,
  ZonedDateTime,
} from "@internationalized/date";
import { toDate } from "reka-ui/date";
import { cn } from "~/lib/utils";
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
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
import { CalendarIcon } from "lucide-vue-next";
import { useFilter } from "reka-ui";
import ComboboxViewport from "./ui/combobox/ComboboxViewport.vue";
import { toast } from "vue-sonner";
import Button from "./ui/button/Button.vue";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import Input from "./ui/input/Input.vue";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

const sources = ref([
  { id: 1, value: "nhentai", label: "NHentai" },
  { id: 2, value: "hentag", label: "Hentag" },
]);

const { data: archive } = useNuxtData("archive");

const df = new DateFormatter("en-GB", {
  dateStyle: "long",
});

const tags: string[] = [];
if (archive.value.tags) {
  for (const tag of archive.value.tags) {
    tags.push(tag.name);
  }
}

const artists: string[] = [];
if (archive.value.artist) {
  for (const artist of archive.value.artist) {
    artists.push(artist.name);
  }
}

const parodies: string[] = [];
if (archive.value.parody) {
  for (const parody of archive.value.parody) {
    parodies.push(parody.name);
  }
}

const characters: string[] = [];
if (archive.value.character) {
  for (const character of archive.value.character) {
    characters.push(character.name);
  }
}

const urls: string[] = [];
if (archive.value.url) {
  for (const url of archive.value.url) {
    urls.push(url.url);
  }
}

const release_date = archive.value.release_date
  ? archive.value.release_date
  : undefined;

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
    title: archive.value.title,
    summary: archive.value.summary,
    tags: tags,
    artist: artists,
    parody: parodies,
    character: characters,
    language: archive.value.language,
    category: archive.value.category,
    url: urls,
    release_date: release_date,
  },
});

const value = computed({
  get: () =>
    values.release_date ? parseAbsolute(values.release_date, "UTC") : undefined,
  set: (val) => val,
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

function updateData(values) {
  // For some values want to overwrite for some not
  // URLs want to append
  setFieldValue("title", values.title);
  setFieldValue("summary", values.summary);
  setFieldValue("artist", values.artist);
  setFieldValue("tags", values.tags);
  setFieldValue("parody", values.parody);
  setFieldValue("character", values.character);
  setFieldValue("language", values.language);
  setFieldValue("category", values.category);
  setFieldValue("release_date", values.release_date);
  setFieldValue("url", values.url);
}

const id = useRoute().params.id;
const onSubmit = handleSubmit((values) => {
  $fetch(`/api/a/${id}`, {
    method: "PUT",
    body: JSON.stringify(values, null, 2),
    onResponseError({ response }) {
      const err = JSON.stringify(response._data.data, null, 2);
      const test = JSON.parse(err);
      if (test.language) {
        toast.error(h("pre", test.language));
      } else if (test.title) {
        toast.error(h("pre", test.title));
      }
    },
    onResponse({ response }) {
      if (response.ok) {
        refreshNuxtData("archive");
        toast.success("Successfully updated archive.");
      }
    },
  });
});
</script>

<template>
  <div id="main" class="flex flex-col gap-1 w-full overflow-auto">
    <!-- Updateable: Title, Summary, Tags, Artists, Parodies, Characters, Languages, Categories, Release Date, URLs -->
    <form @submit="onSubmit">
      <!-- Title -->
      <FormField v-slot="{ componentField }" name="title">
        <FormItem>
          <FormLabel>Title</FormLabel>
          <FormControl>
            <Input type="text" v-bind="componentField" />
          </FormControl>
          <FormDescription />
          <FormMessage />
        </FormItem>
      </FormField>

      <!-- Summary -->
      <FormField v-slot="{ componentField }" name="summary">
        <FormItem>
          <FormLabel>Summary</FormLabel>
          <FormControl>
            <Textarea type="text" class="resize-none" v-bind="componentField" />
          </FormControl>
          <FormDescription />
          <FormMessage />
        </FormItem>
      </FormField>

      <!-- Artists -->
      <FormField v-slot="{ componentField }" name="artist">
        <FormItem>
          <FormLabel>Artist</FormLabel>
          <Combobox
            v-model="componentField.modelValue"
            v-model:open="openArtist"
            :ignore-filter="true"
          >
            <FormControl>
              <ComboboxAnchor as-child>
                <TagsInput
                  :model-value="componentField.modelValue"
                  class="bg-input/30 w-full"
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
                              componentField.modelValue.push(ev.detail.value);
                            }

                            if (artists.length === 0) {
                              openArtist = false;
                            }
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

      <!-- Tags -->
      <FormField v-slot="{ componentField }" name="tags">
        <FormItem>
          <FormLabel>Tags</FormLabel>

          <Combobox
            v-model="componentField.modelValue"
            v-model:open="openTag"
            :ignore-filter="true"
          >
            <FormControl>
              <ComboboxAnchor as-child>
                <TagsInput
                  :model-value="componentField.modelValue"
                  class="bg-input/30 w-full"
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
                              componentField.modelValue.push(ev.detail.value);
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

      <!-- Parodies & Characters -->
      <div class="grid sm:grid-cols-2 w-full gap-2">
        <!-- Parodies -->
        <FormField v-slot="{ componentField }" name="parody">
          <FormItem>
            <FormLabel>Parody</FormLabel>

            <Combobox
              v-model="componentField.modelValue"
              v-model:open="openParody"
              :ignore-filter="true"
            >
              <FormControl>
                <ComboboxAnchor as-child>
                  <TagsInput
                    :model-value="componentField.modelValue"
                    class="bg-input/30 w-full"
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
                                componentField.modelValue.push(ev.detail.value);
                              }

                              if (parodies.length === 0) {
                                openParody = false;
                              }
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

        <!-- Characters -->
        <FormField v-slot="{ componentField }" name="character">
          <FormItem>
            <FormLabel>Character</FormLabel>
            <Combobox
              v-model="componentField.modelValue"
              v-model:open="openCharacter"
              :ignore-filter="true"
            >
              <FormControl>
                <ComboboxAnchor as-child>
                  <TagsInput
                    :model-value="componentField.modelValue"
                    class="bg-input/30 w-full"
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
                                componentField.modelValue.push(ev.detail.value);
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

      <!-- Language, Category & Release Date-->
      <div class="grid sm:grid-cols-3 gap-2">
        <!-- Language -->
        <FormField v-slot="{ componentField }" name="language">
          <FormItem>
            <FormLabel>Language</FormLabel>
            <Combobox
              v-model:model-value="componentField.modelValue"
              v-model:open="openLanguage"
              :ignore-filter="true"
            >
              <FormControl>
                <ComboboxAnchor as-child>
                  <ComboboxInput v-model="searchLanguage" as-child>
                    <Input class="w-full" type="text" v-bind="componentField" />
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

        <!-- Category -->
        <FormField v-slot="{ componentField }" name="category">
          <FormItem>
            <FormLabel>Category</FormLabel>
            <Combobox
              v-model:model-value="componentField.modelValue"
              v-model:open="openCategory"
              :ignore-filter="true"
            >
              <FormControl>
                <ComboboxAnchor as-child>
                  <ComboboxInput v-model="searchCategory" as-child>
                    <Input class="w-full" type="text" v-bind="componentField" />
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

        <!-- Release Date -->
        <FormField name="rd">
          <FormItem class="flex flex-col">
            <FormLabel>Release Date</FormLabel>
            <Popover>
              <PopoverTrigger as-child>
                <FormControl>
                  <Button
                    variant="outline"
                    :class="
                      cn(
                        'w-full ps-3 text-start font-normal',
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
                          parseZonedDateTime(v.toString()).toAbsoluteString(),
                        );
                      } else {
                        setFieldValue('release_date', undefined);
                      }
                    }
                  "
                />
              </PopoverContent>
            </Popover>
            <FormDescription />
            <FormMessage />
          </FormItem>
        </FormField>
      </div>

      <!-- URLs -->
      <FormField v-slot="{ componentField }" name="url">
        <FormItem>
          <FormLabel>URL</FormLabel>
          <FormControl>
            <TagsInput
              :model-value="componentField.modelValue"
              class="bg-input/30"
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
      <div class="flex gap-2 flex-1">
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button variant="secondary" class="flex-1 bg-muted"
              >Fetch with...</Button
            >
          </DropdownMenuTrigger>
          <DropdownMenuContent class="w-(--reka-dropdown-menu-trigger-width)">
            <UpdateMetadataDialog
              v-for="source in sources"
              :key="source.id"
              :source="source"
              @save="updateData"
            />
          </DropdownMenuContent>
        </DropdownMenu>
        <Button class="w-full flex-1" type="submit">Save</Button>
      </div>
    </form>
  </div>
</template>
