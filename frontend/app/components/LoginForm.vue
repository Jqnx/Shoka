<script setup lang="ts">
import type { HTMLAttributes } from "vue";
import { cn } from "~/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
} from "@/components/ui/form";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { z } from "zod";
import { toast } from "vue-sonner";

const props = defineProps<{
  class?: HTMLAttributes["class"];
}>();

const formSchema = toTypedSchema(
  z.object({
    username: z.string(),
    password: z.string(),
  }),
);

const { handleSubmit, errors } = useForm({
  validationSchema: formSchema,
});

const { signIn } = useAuth();

const onSubmit = handleSubmit(async (values) => {
  await signIn.username({
    username: values.username,
    password: values.password,
    fetchOptions: {
      onSuccess: () => {
        navigateTo("/");
      },
      onError: (ctx) => {
        toast.error(ctx.error.message);
      },
    },
  });
});
</script>

<template>
  <div :class="cn('flex flex-col gap-6', props.class)">
    <Card>
      <CardHeader class="text-center">
        <CardTitle class="text-xl"> Welcome back </CardTitle>
        <CardDescription />
      </CardHeader>
      <CardContent>
        <form @submit="onSubmit">
          <div class="grid gap-6">
            <div class="grid gap-6">
              <FormField v-slot="{ componentField }" name="username">
                <FormItem class="grid gap-3">
                  <FormLabel>Username</FormLabel>
                  <FormControl>
                    <Input type="username" v-bind="componentField" />
                  </FormControl>
                  <FormLabel v-if="errors.username">
                    <p class="text-destructive">
                      {{ errors.username }}
                    </p>
                  </FormLabel>
                </FormItem>
              </FormField>
              <FormField v-slot="{ componentField }" name="password">
                <FormItem class="grid gap-3">
                  <FormLabel>Password</FormLabel>
                  <FormControl>
                    <Input type="password" v-bind="componentField" />
                  </FormControl>
                  <FormLabel v-if="errors.password">
                    <p class="text-destructive">
                      {{ errors.password }}
                    </p>
                  </FormLabel>
                </FormItem>
              </FormField>
              <Button type="submit" class="w-full"> Login </Button>
            </div>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
