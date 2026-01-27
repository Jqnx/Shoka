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
  token?: string;
}>();

const formSchema = toTypedSchema(
  z.object({
    url: z.string({ required_error: "URL is required." }).url(),
  }),
);

const { handleSubmit, errors, setFieldValue } = useForm({
  validationSchema: formSchema,
});

const { data } = useLazyFetch("/api/config/flaresolverr", {
  key: "flaresolverr",
  onRequest({ options }) {
    options.headers.set("Authorization", `Bearer ${props.token}`);
  },
});

watch(data, () => {
  setFieldValue("url", data.value.url);
});

const onSubmit = handleSubmit((values) => {
  $fetch(`/api/config/flaresolverr`, {
    method: "POST",
    body: JSON.stringify(values, null, 2),
    onRequest({ options }) {
      options.headers.set("Authorization", `Bearer ${props.token}`);
    },
    onResponseError({ response }) {
      const err = JSON.stringify(response._data.data, null, 2);
      const parsedErr = JSON.parse(err);
      toast.error(h("pre", parsedErr));
    },
    onResponse({ response }) {
      if (response.ok) {
        refreshNuxtData("flaresolverr");
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
          src="/assets/flaresolverr.png"
          format="webp"
          height="20"
          loading="lazy"
          class="rounded-sm"
        />
        <CardTitle class="text-xl"> Flaresolverr </CardTitle>
        <CardDescription />
      </CardHeader>
      <CardContent>
        <form @submit="onSubmit">
          <div class="grid gap-6">
            <div class="grid gap-6">
              <FormField v-slot="{ componentField }" name="url">
                <FormItem class="grid gap-3">
                  <FormLabel>URL</FormLabel>
                  <FormControl>
                    <Input type="url" v-bind="componentField" />
                  </FormControl>
                  <FormLabel v-if="errors.url">
                    <p class="text-destructive">
                      {{ errors.url }}
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
