<script lang="ts" setup>
  import AppSidebar from "@/components/AppSidebar.vue";
  import { Separator } from "@/components/ui/separator";
  import {
    SidebarInset,
    SidebarProvider,
    SidebarTrigger,
  } from "@/components/ui/sidebar";
  const defaultOpen = useCookie<boolean>("sidebar:state");
</script>

<template>
  <SidebarProvider :defaultOpen="defaultOpen">
    <AppSidebar />
    <SidebarInset>
      <header class="flex h-16 shrink-0 items-center gap-2">
        <div class="flex items-center gap-2 px-4">
          <SidebarTrigger class="-ml-1" />
          <Separator
            orientation="vertical"
            class="mr-2 data-[orientation=vertical]:h-4" />
        </div>
        <div class="flex flex-1 justify-center items-center gap-2 px-4">
          <div class="relative w-full max-w-xl items-center">
            <Input
              id="search"
              type="text"
              placeholder="Search..."
              class="pl-10" />
            <span
              class="absolute start-0 inset-y-0 flex items-center justify-center px-2">
              <Icon
                name="lucide:search"
                size="1.25em"
                class="text-muted-foreground" />
            </span>
          </div>
        </div>
      </header>
      <slot />
    </SidebarInset>
  </SidebarProvider>
</template>
