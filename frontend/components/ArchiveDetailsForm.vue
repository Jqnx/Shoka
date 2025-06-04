<script setup lang="ts">
  import { z } from "zod/v3";
  import {
    DateFormatter,
    getLocalTimeZone,
    parseAbsolute,
    parseZonedDateTime,
    today,
    ZonedDateTime,
  } from "@internationalized/date";
  import { toDate } from "reka-ui/date";
  import { cn } from "@/lib/utils";
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
  import { Textarea } from "@/components/ui/textarea";
  import { toast } from "vue-sonner";
  import { Calendar } from "@/components/ui/calendar";
  import { CalendarIcon } from "lucide-vue-next";

  const { data: archive } = useNuxtData("archive");

  const df = new DateFormatter("en-GB", {
    dateStyle: "long",
  });

  const tags: string[] = [];
  if (archive.value.tags) {
    for (const tag of archive.value.tags) {
      tags.push(tag.tag);
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
      parodies.push(parody.parody);
    }
  }

  const characters: string[] = [];
  if (archive.value.character) {
    for (const character of archive.value.character) {
      characters.push(character.character);
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
    })
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
      values.release_date
        ? parseAbsolute(values.release_date, "UTC")
        : undefined,
    set: (val) => val,
  });

  const placeholder = ref();

  const id = useRoute().params.id;
  const onSubmit = handleSubmit((values) => {
    const res = $fetch(`/api/a/${id}`, {
      method: "PUT",
      body: JSON.stringify(values, null, 2),
      onResponse() {
        // refresh nuxt data on response
        console.log(res);
      },
    });
    toast("Form Submitted", {
      description: h(
        "pre",
        h("code", { class: "text-white" }, JSON.stringify(values, null, 2))
      ),
    });
  });
</script>

<template>
  <div class="flex flex-col gap-1 w-full overflow-auto">
    <!-- Updateable: Title, Summary, Tags, Artists, Parodies, Characters, Languages, Categories, Release Date, URLs -->
    <form @submit="onSubmit">
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
      <FormField v-slot="{ componentField }" name="artist">
        <FormItem>
          <FormLabel>Artist</FormLabel>
          <FormControl>
            <TagsInput
              :model-value="componentField.modelValue"
              class="bg-transparent"
              @update:model-value="componentField['onUpdate:modelValue']">
              <TagsInputItem
                v-for="item in componentField.modelValue"
                :key="item"
                :value="item">
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
      <FormField v-slot="{ componentField }" name="tags">
        <FormItem>
          <FormLabel>Tags</FormLabel>
          <FormControl>
            <TagsInput
              :model-value="componentField.modelValue"
              class="bg-transparent"
              @update:model-value="componentField['onUpdate:modelValue']">
              <TagsInputItem
                v-for="item in componentField.modelValue"
                :key="item"
                :value="item">
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
      <div class="grid sm:grid-cols-2 w-full gap-2">
        <FormField v-slot="{ componentField }" name="parody">
          <FormItem>
            <FormLabel>Parody</FormLabel>
            <FormControl>
              <TagsInput
                :model-value="componentField.modelValue"
                class="bg-transparent"
                @update:model-value="componentField['onUpdate:modelValue']">
                <TagsInputItem
                  v-for="item in componentField.modelValue"
                  :key="item"
                  :value="item">
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
        <FormField v-slot="{ componentField }" name="character">
          <FormItem>
            <FormLabel>Character</FormLabel>
            <FormControl>
              <TagsInput
                :model-value="componentField.modelValue"
                class="bg-transparent"
                @update:model-value="componentField['onUpdate:modelValue']">
                <TagsInputItem
                  v-for="item in componentField.modelValue"
                  :key="item"
                  :value="item">
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
      <div class="grid sm:grid-cols-3 gap-2">
        <FormField v-slot="{ componentField }" name="language">
          <FormItem>
            <FormLabel>Language</FormLabel>
            <FormControl>
              <Input type="text" v-bind="componentField" />
            </FormControl>
            <FormDescription />
            <FormMessage />
          </FormItem>
        </FormField>
        <FormField v-slot="{ componentField }" name="category">
          <FormItem>
            <FormLabel>Category</FormLabel>
            <FormControl>
              <Input type="text" v-bind="componentField" />
            </FormControl>
            <FormDescription />
            <FormMessage />
          </FormItem>
        </FormField>
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
                        !value && 'text-muted-foreground'
                      )
                    ">
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
                          parseZonedDateTime(v.toString()).toAbsoluteString()
                        );
                      } else {
                        setFieldValue('release_date', undefined);
                      }
                    }
                  " />
              </PopoverContent>
            </Popover>
            <FormDescription />
            <FormMessage />
          </FormItem>
        </FormField>
      </div>
      <FormField v-slot="{ componentField }" name="url">
        <FormItem>
          <FormLabel>URL</FormLabel>
          <FormControl>
            <TagsInput
              :model-value="componentField.modelValue"
              class="bg-transparent"
              @update:model-value="componentField['onUpdate:modelValue']">
              <TagsInputItem
                v-for="item in componentField.modelValue"
                :key="item"
                :value="item">
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
      <Button class="w-full" type="submit">Submit</Button>
    </form>
  </div>
</template>
