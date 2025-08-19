<script setup lang="ts">
  import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  } from "@/components/ui/tabs";
  import { Library } from "lucide-vue-next";
  definePageMeta({
    layout: "login",
  });
</script>

<template>
  <div
    class="bg-background flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
    <div class="flex w-full max-w-sm flex-col gap-6">
      <a href="#" class="flex items-center gap-2 self-center font-medium">
        <div
          class="bg-primary text-primary-foreground flex size-6 items-center justify-center rounded-md">
          <Library class="size-4" />
        </div>
        Shoka
      </a>

      <Tabs default-value="login">
        <TabsList class="bg-card grid w-full grid-cols-2">
          <TabsTrigger
            class="dark:data-[state=active]:bg-input h-[calc(100%-2px)] cursor-pointer"
            value="login"
            >Login</TabsTrigger
          >
          <TabsTrigger
            class="dark:data-[state=active]:bg-input h-[calc(100%-2px)] cursor-pointer"
            value="signup"
            >Sign Up</TabsTrigger
          >
        </TabsList>
        <TabsContent value="login">
          <LoginForm />
        </TabsContent>
        <TabsContent value="signup">
          <SignupForm />
        </TabsContent>
      </Tabs>
    </div>
  </div>
</template>
