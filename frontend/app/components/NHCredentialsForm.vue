<script setup lang="ts">
import type { HTMLAttributes } from "vue";
import { cn } from "~/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
} from "@/components/ui/form";
import { z } from "zod";
import { toast } from "vue-sonner";

const props = defineProps<{
  class?: HTMLAttributes["class"];
}>();

const formSchema = toTypedSchema(
  z.object({
    csrftoken: z.string({ required_error: "CSRFToken is required." }),
    useragent: z.string({ required_error: "UserAgent is required." }),
  }),
);

const { handleSubmit, errors, setFieldValue } = useForm({
  validationSchema: formSchema,
});

const { data: creds } = useLazyFetch("/api/config/nh", {
  key: "nhcredentials",
});

watch(creds, () => {
  setFieldValue("csrftoken", creds.value.csrftoken);
  setFieldValue("useragent", creds.value.useragent);
});

const onSubmit = handleSubmit((values) => {
  $fetch(`/api/config/nh`, {
    method: "POST",
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
        refreshNuxtData("nhcredentials");
        toast.success("Successfully set credentials.");
      }
    },
  });
});
</script>

<template>
  <div :class="cn('flex flex-col gap-6', props.class)">
    <Card class="border-none shadow-none">
      <CardHeader class="flex justify-center items-center pt-2">
        <NuxtImg
          src="/assets/nhentai.png"
          format="webp"
          width="20"
          height="20"
          loading="lazy"
          class="rounded-sm"
        />
        <CardTitle class="text-xl"> NHentai </CardTitle>
        <CardDescription />
      </CardHeader>
      <CardContent>
        <form @submit="onSubmit">
          <div class="grid gap-6">
            <div class="grid gap-6">
              <FormField v-slot="{ componentField }" name="csrftoken">
                <FormItem class="grid gap-3">
                  <FormLabel>CSRFToken</FormLabel>
                  <FormControl>
                    <Input type="csrftoken" v-bind="componentField" />
                  </FormControl>
                  <FormLabel v-if="errors.csrftoken">
                    <p class="text-destructive">
                      {{ errors.csrftoken }}
                    </p>
                  </FormLabel>
                </FormItem>
              </FormField>
              <FormField v-slot="{ componentField }" name="useragent">
                <FormItem class="grid gap-3">
                  <FormLabel>User-Agent</FormLabel>
                  <FormControl>
                    <Input type="useragent" v-bind="componentField" />
                  </FormControl>
                  <FormLabel v-if="errors.useragent">
                    <p class="text-destructive">
                      {{ errors.useragent }}
                    </p>
                  </FormLabel>
                </FormItem>
              </FormField>
              <Button type="submit" class="w-full">Save</Button>
            </div>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
