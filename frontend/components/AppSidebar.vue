<script setup lang="ts">
  import NavMain from "@/components/NavMain.vue";

  import NavUser from "@/components/NavUser.vue";
  import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    type SidebarProps,
  } from "@/components/ui/sidebar";
  import {
    BookOpen,
    Command,
    Settings2,
    Book,
    Library,
    Image,
    Images,
    Album,
    BookImage,
    BookUser,
    Users,
    User,
    Tag,
    PersonStanding,
  } from "lucide-vue-next";

  const props = withDefaults(defineProps<SidebarProps>(), {
    variant: "inset",
  });

  const data = {
    user: {
      name: "user",
      email: "user@example.com",
      avatar: "",
    },
    navMain: [
      {
        title: "Archives",
        url: "/a",
        icon: Library,
        isActive: true,
      },
      {
        title: "Covers",
        url: "#",
        icon: Image,
      },
      {
        title: "Illustrations",
        url: "#",
        icon: Images,
      },
      {
        title: "Spreads",
        url: "#",
        icon: Album,
      },
      {
        title: "Tankoubons",
        url: "#",
        icon: BookImage,
      },
    ],
    navSecondary: [
      {
        title: "Artists",
        url: "/artist",
        icon: User,
      },
      {
        title: "Groups",
        url: "/group",
        icon: Users,
      },
      {
        title: "Tags",
        url: "/tag",
        icon: Tag,
      },
      {
        title: "Characters",
        url: "/character",
        icon: PersonStanding,
      },
      {
        title: "Parodies",
        url: "/parody",
        icon: BookUser,
      },
    ],
  };
</script>

<template>
  <Sidebar v-bind="props">
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton
            size="lg"
            as-child>
            <NuxtLink to="/">
              <div
                class="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-sidebar-primary-foreground">
                <Album class="size-4" />
              </div>
              <div class="grid flex-1 text-left text-sm leading-tight">
                <span class="truncate font-medium">Shoka</span>
              </div>
            </NuxtLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
    <SidebarContent>
      <NavMain
        label="Media"
        :items="data.navMain" />
      <NavMain
        label="Metadata"
        :items="data.navSecondary" />

      <!--
      <NavSecondary :items="data.navSecondary" class="mt-auto" />
      -->
    </SidebarContent>
    <SidebarFooter>
      <NavUser :user="data.user" />
    </SidebarFooter>
  </Sidebar>
</template>
