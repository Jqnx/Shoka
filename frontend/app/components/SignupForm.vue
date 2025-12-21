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
  z
    .object({
      name: z
        .string({ required_error: "Name is required." })
        .min(6, { message: "Must be atleast 6 characters long." })
        .max(64, "Cannot be longer than 64 characters."),
      email: z.string({ required_error: "Email is required." }).email(),
      username: z
        .string({ required_error: "Username is required." })
        .min(2, { message: "Must be atleast 2 characters long." })
        .max(64, { message: "Cannot be longer than 64 characters." }),
      password: z
        .string({ required_error: "Password is required." })
        .min(8, { message: "Must be atleast 8 characters long." })
        .max(64, { message: "Cannot be longer than 64 characters." }),
      vpassword: z
        .string({ required_error: "Verify Password is required." })
        .min(8, { message: "Must be atleast 8 characters long." })
        .max(64, { message: "Cannot be longer than 64 characters." }),
    })
    .superRefine(({ password, vpassword }, ctx) => {
      if (vpassword !== password) {
        ctx.addIssue({
          code: "custom",
          message: "Passwords must match.",
          path: ["vpassword"],
        });
      }
    }),
);

const { handleSubmit, errors } = useForm({
  validationSchema: formSchema,
});

const onSubmit = handleSubmit((values) => {
  const { signUp } = useAuth();
  signUp.email({
    email: values.email,
    name: values.name,
    password: values.password,
    username: values.username,
    fetchOptions: {
      onSuccess: () => {
        navigateTo("/");
      },
      onError(ctx) {
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
        <CardTitle class="text-xl"> Sign Up </CardTitle>
        <CardDescription />
      </CardHeader>
      <CardContent>
        <form @submit="onSubmit">
          <div class="grid gap-6">
            <div class="grid gap-6">
              <FormField v-slot="{ componentField }" name="name">
                <FormItem class="grid gap-3">
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input type="name" v-bind="componentField" />
                  </FormControl>
                  <FormLabel v-if="errors.name">
                    <p class="text-destructive">
                      {{ errors.name }}
                    </p>
                  </FormLabel>
                </FormItem>
              </FormField>
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
              <FormField v-slot="{ componentField }" name="email">
                <FormItem class="grid gap-3">
                  <FormLabel>Email</FormLabel>
                  <FormControl>
                    <Input type="email" v-bind="componentField" />
                  </FormControl>
                  <FormLabel v-if="errors.email">
                    <p class="text-destructive">
                      {{ errors.email }}
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
              <FormField v-slot="{ componentField }" name="vpassword">
                <FormItem class="grid gap-3">
                  <FormLabel>Verify Password</FormLabel>
                  <FormControl>
                    <Input type="password" v-bind="componentField" />
                  </FormControl>
                  <FormLabel v-if="errors.vpassword">
                    <p class="text-destructive">
                      {{ errors.vpassword }}
                    </p>
                  </FormLabel>
                </FormItem>
              </FormField>
              <Button type="submit" class="w-full"> Sign Up </Button>
            </div>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
